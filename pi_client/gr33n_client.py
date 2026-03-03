#!/usr/bin/env python3
# gr33n Pi Client - sensor + actuator daemon

import logging
import sqlite3
import threading
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

import requests
import yaml

# ── Optional GPIO (stubs when running off-Pi) ───────────────────────────────
try:
    from gpiozero import OutputDevice
    GPIO_AVAILABLE = True
except (ImportError, RuntimeError):
    GPIO_AVAILABLE = False
    class OutputDevice:  # noqa: F811
        def __init__(self, pin, **kw): self.pin = pin; self._on = False
        def on(self):  self._on = True
        def off(self): self._on = False

try:
    import adafruit_dht, board
    DHT_AVAILABLE = True
except ImportError:
    DHT_AVAILABLE = False

try:
    import busio, adafruit_ads1x15.ads1115 as ADS
    from adafruit_ads1x15.analog_in import AnalogIn
    ADS_AVAILABLE = True
except ImportError:
    ADS_AVAILABLE = False

try:
    import serial
    SERIAL_AVAILABLE = True
except ImportError:
    SERIAL_AVAILABLE = False

try:
    import smbus2
    I2C_BUS_AVAILABLE = True
except ImportError:
    I2C_BUS_AVAILABLE = False

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(name)s - %(message)s',
)
log = logging.getLogger('gr33n.client')


# --- CONFIG -----------------------------------------------------------------
DEFAULT_CONFIG = {
    'api': {
        'base_url': 'http://192.168.1.100:8080',
        'timeout_seconds': 5,
        'api_key': '',
    },
    'farm': {'farm_id': 1},
    'sensors': [
        {'sensor_id': 1, 'sensor_type': 'temperature',   'source': 'dht22',   'pin': 4,  'interval_seconds': 60},
        {'sensor_id': 2, 'sensor_type': 'humidity',      'source': 'dht22',   'pin': 4,  'interval_seconds': 60},
        {'sensor_id': 3, 'sensor_type': 'soil_moisture', 'source': 'ads1115', 'channel': 0, 'interval_seconds': 300},
        {'sensor_id': 4, 'sensor_type': 'co2',           'source': 'mhz19',   'port': '/dev/ttyS0', 'interval_seconds': 60},
        {'sensor_id': 5, 'sensor_type': 'ec',            'source': 'ads1115', 'channel': 1, 'interval_seconds': 60},
        {'sensor_id': 6, 'sensor_type': 'ph',            'source': 'ads1115', 'channel': 2, 'interval_seconds': 60},
        {'sensor_id': 7, 'sensor_type': 'par',           'source': 'bh1750',  'interval_seconds': 60},
    ],
    'actuators': [
        {'actuator_id': 1, 'device_type': 'light',      'gpio_pin': 17},
        {'actuator_id': 2, 'device_type': 'irrigation', 'gpio_pin': 27},
        {'actuator_id': 3, 'device_type': 'fan',        'gpio_pin': 22},
    ],
    'schedule_poll_interval_seconds': 30,
    'offline_queue_path': '/var/lib/gr33n/queue.db',
    'offline_flush_interval_seconds': 60,
}


def load_config(path: str = 'config.yaml') -> dict:
    cfg = DEFAULT_CONFIG.copy()
    p = Path(path)
    if p.exists():
        with open(p) as fh:
            user_cfg = yaml.safe_load(fh) or {}
        for k, v in user_cfg.items():
            if isinstance(v, dict) and isinstance(cfg.get(k), dict):
                cfg[k].update(v)
            else:
                cfg[k] = v
    return cfg


# --- OFFLINE QUEUE (SQLite) -------------------------------------------------
class OfflineQueue:
    def __init__(self, db_path: str):
        Path(db_path).parent.mkdir(parents=True, exist_ok=True)
        self._conn = sqlite3.connect(db_path, check_same_thread=False)
        self._lock = threading.Lock()
        self._init_schema()

    def _init_schema(self):
        with self._lock, self._conn:
            self._conn.execute(
                'CREATE TABLE IF NOT EXISTS pending_readings ('
                '    id           INTEGER PRIMARY KEY AUTOINCREMENT,'
                '    sensor_id    INTEGER NOT NULL,'
                '    value_raw    REAL    NOT NULL,'
                '    reading_time TEXT    NOT NULL,'
                '    attempts     INTEGER DEFAULT 0,'
                '    created_at   TEXT    DEFAULT (datetime("now"))'
                ')'
            )

    def push(self, sensor_id: int, value_raw: float, reading_time: str):
        with self._lock, self._conn:
            self._conn.execute(
                'INSERT INTO pending_readings (sensor_id, value_raw, reading_time) VALUES (?,?,?)',
                (sensor_id, value_raw, reading_time),
            )
        log.debug('Queued offline reading sensor_id=%s value=%s', sensor_id, value_raw)

    def pop_batch(self, limit: int = 50) -> list:
        with self._lock:
            rows = self._conn.execute(
                'SELECT id, sensor_id, value_raw, reading_time FROM pending_readings '
                'ORDER BY id LIMIT ?', (limit,)
            ).fetchall()
        return [{'_qid': r[0], 'sensor_id': r[1], 'value_raw': r[2], 'reading_time': r[3]} for r in rows]

    def ack(self, qids: list):
        if not qids:
            return
        ph = ','.join('?' * len(qids))
        with self._lock, self._conn:
            self._conn.execute(f'DELETE FROM pending_readings WHERE id IN ({ph})', qids)


# --- API CLIENT -------------------------------------------------------------
class Gr33nApiClient:
    def __init__(self, base_url: str, farm_id: int, api_key: str = '', timeout: int = 5):
        self.base_url = base_url.rstrip('/')
        self.farm_id  = farm_id
        self.timeout  = timeout
        self._s = requests.Session()
        if api_key:
            self._s.headers['X-Api-Key'] = api_key
        self._s.headers['Content-Type'] = 'application/json'

    def is_reachable(self) -> bool:
        try:
            return self._s.get(f'{self.base_url}/health', timeout=self.timeout).status_code == 200
        except requests.RequestException:
            return False

    def post_reading(self, sensor_id: int, value_raw: float, reading_time: Optional[str] = None) -> bool:
        # Body mirrors gr33ncore.sensor_readings columns (schema FIX #4: PK is reading_time,sensor_id)
        payload = {
            'sensor_id':    sensor_id,
            'value_raw':    value_raw,
            'reading_time': reading_time or datetime.now(timezone.utc).isoformat(),
            'is_valid':     True,
        }
        try:
            r = self._s.post(f'{self.base_url}/sensors/{sensor_id}/readings',
                             json=payload, timeout=self.timeout)
            if r.status_code in (200, 201):
                return True
            log.warning('API rejected reading sensor_id=%s status=%s body=%s',
                        sensor_id, r.status_code, r.text[:200])
            return False
        except requests.RequestException as exc:
            log.debug('API unreachable: %s', exc)
            return False

    def get_devices(self) -> list:
        # GET /farms/{id}/devices - matches handleListDevices in routes.go
        try:
            r = self._s.get(f'{self.base_url}/farms/{self.farm_id}/devices', timeout=self.timeout)
            if r.status_code == 200:
                data = r.json()
                return data if isinstance(data, list) else []
        except requests.RequestException as exc:
            log.debug('Could not fetch devices: %s', exc)
        return []

    def patch_device_status(self, device_id: int, status: str) -> bool:
        # PATCH /devices/{id}/status - body: {status: string}
        # Allowed values (device_status_enum): online | offline | error_comms |
        #   error_hardware | maintenance_mode | initializing | unknown |
        #   decommissioned | pending_activation
        try:
            r = self._s.patch(f'{self.base_url}/devices/{device_id}/status',
                              json={'status': status}, timeout=self.timeout)
            return r.status_code == 200
        except requests.RequestException:
            return False

    def post_actuator_event(self, actuator_id: int, command: str,
                             source: str = 'schedule_trigger',
                             schedule_id: Optional[int] = None) -> bool:
        # Body mirrors gr33ncore.actuator_events columns
        # source enum: manual_api_call | schedule_trigger | automation_rule_trigger |
        #   device_internal_feedback_loop | system_initialization_routine | emergency_stop_signal
        # execution_status: command_sent_to_device (initial state)
        payload = {
            'actuator_id':              actuator_id,
            'command_sent':             command,
            'source':                   source,
            'event_time':               datetime.now(timezone.utc).isoformat(),
            'execution_status':         'command_sent_to_device',
            'triggered_by_schedule_id': schedule_id,
        }
        try:
            r = self._s.post(f'{self.base_url}/actuators/{actuator_id}/events',
                             json=payload, timeout=self.timeout)
            return r.status_code in (200, 201)
        except requests.RequestException:
            return False


# --- SENSOR READER ----------------------------------------------------------
class SensorReader:
    def __init__(self, cfg: dict):
        self.cfg = cfg
        self._dht = self._adc = self._uart = self._i2c = None
        self._init_hardware()

    def _init_hardware(self):
        src = self.cfg.get('source', '')
        if src == 'dht22' and DHT_AVAILABLE:
            pin = getattr(board, f'D{self.cfg["pin"]}', board.D4)
            self._dht = adafruit_dht.DHT22(pin, use_pulseio=False)
        elif src == 'ads1115' and ADS_AVAILABLE:
            i2c = busio.I2C(board.SCL, board.SDA)
            ads = ADS.ADS1115(i2c)
            ch_map = [ADS.P0, ADS.P1, ADS.P2, ADS.P3]
            self._adc = AnalogIn(ads, ch_map[self.cfg.get('channel', 0)])
        elif src == 'mhz19' and SERIAL_AVAILABLE:
            self._uart = serial.Serial(self.cfg.get('port', '/dev/ttyS0'), baudrate=9600, timeout=2)
        elif src == 'bh1750' and I2C_BUS_AVAILABLE:
            self._i2c = smbus2.SMBus(1)

    def read(self) -> Optional[float]:
        src   = self.cfg.get('source', '')
        stype = self.cfg.get('sensor_type', '')
        if src == 'dht22':
            if not self._dht:
                return self._mock(stype)
            try:
                return float(self._dht.temperature if stype == 'temperature' else self._dht.humidity)
            except RuntimeError:
                return None
        elif src == 'ads1115':
            if not self._adc:
                return self._mock(stype)
            v = self._adc.voltage
            if stype == 'soil_moisture':
                # Capacitive sensor: 3.0V=0%(dry), 1.5V=100%(wet)
                return round(max(0.0, min(100.0, (3.0 - v) / 1.5 * 100.0)), 1)
            elif stype == 'ec':
                # Linear 0-5 mS/cm on 0-3.3V
                return round(v / 3.3 * 5.0, 3)
            elif stype == 'ph':
                # Atlas Scientific analog: 7pH=2.5V, ~0.18V/pH
                return round(7.0 + (2.5 - v) / 0.18, 2)
            return v
        elif src == 'mhz19':
            if not self._uart:
                return self._mock('co2')
            cmd = bytes([0xFF, 0x01, 0x86, 0x00, 0x00, 0x00, 0x00, 0x00, 0x79])
            self._uart.write(cmd)
            resp = self._uart.read(9)
            if len(resp) == 9 and resp[0] == 0xFF and resp[1] == 0x86:
                return float(resp[2] * 256 + resp[3])
            return None
        elif src == 'bh1750':
            if not self._i2c:
                return self._mock('par')
            try:
                self._i2c.write_byte(0x23, 0x10)
                time.sleep(0.18)
                data = self._i2c.read_i2c_block_data(0x23, 0x10, 2)
                lux = (data[0] << 8 | data[1]) / 1.2
                return round(lux * 0.0185, 1)  # lux to PAR umol/m2/s
            except Exception:
                return None
        return self._mock(stype)

    @staticmethod
    def _mock(stype: str) -> float:
        return {'temperature': 22.5, 'humidity': 58.0, 'soil_moisture': 42.0,
                'co2': 820.0, 'ec': 1.4, 'ph': 6.2, 'par': 380.0}.get(stype, 0.0)


# --- ACTUATOR CONTROLLER ----------------------------------------------------
class ActuatorController:
    def __init__(self, cfg: dict):
        self.actuator_id = cfg['actuator_id']
        self.device_type = cfg['device_type']
        self.gpio_pin    = cfg['gpio_pin']
        # active_high=False for common optocoupler relay boards (LOW = ON)
        self._gpio  = OutputDevice(self.gpio_pin, active_high=False, initial_value=False)
        self._state = False
        log.info('Actuator %s (%s) bound to GPIO pin %s',
                 self.actuator_id, self.device_type, self.gpio_pin)

    def turn_on(self):
        self._gpio.on(); self._state = True
        log.info('Actuator %s (%s) -> ON', self.actuator_id, self.device_type)

    def turn_off(self):
        self._gpio.off(); self._state = False
        log.info('Actuator %s (%s) -> OFF', self.actuator_id, self.device_type)

    def execute(self, command: str):
        cmd = command.strip().lower()
        if cmd in ('on', 'actuator_on', 'turn_on', 'open', 'start'):
            self.turn_on()
        elif cmd in ('off', 'actuator_off', 'turn_off', 'close', 'stop'):
            self.turn_off()
        else:
            log.warning('Unknown command %r for actuator %s', command, self.actuator_id)

    @property
    def state(self) -> str:
        return 'on' if self._state else 'off'


# --- MAIN DAEMON ------------------------------------------------------------
class Gr33nPiClient:
    def __init__(self, config_path: str = 'config.yaml'):
        self.cfg  = load_config(config_path)
        self.api  = Gr33nApiClient(
            base_url = self.cfg['api']['base_url'],
            farm_id  = self.cfg['farm']['farm_id'],
            api_key  = self.cfg['api'].get('api_key', ''),
            timeout  = self.cfg['api']['timeout_seconds'],
        )
        self.queue      = OfflineQueue(self.cfg['offline_queue_path'])
        self._stop      = threading.Event()
        self._last_read: dict = {}
        self._readers: dict = {s['sensor_id']: SensorReader(s) for s in self.cfg['sensors']}
        self._actuators: dict = {a['actuator_id']: ActuatorController(a) for a in self.cfg['actuators']}

    def _sensor_loop(self):
        while not self._stop.is_set():
            now = time.time()
            for scfg in self.cfg['sensors']:
                sid      = scfg['sensor_id']
                interval = scfg.get('interval_seconds', 60)
                if now - self._last_read.get(sid, 0) < interval:
                    continue
                reader = self._readers.get(sid)
                value  = reader.read() if reader else None
                if value is None:
                    log.warning('No reading from sensor_id=%s', sid)
                    continue
                self._last_read[sid] = now
                ts = datetime.now(timezone.utc).isoformat()
                log.debug('sensor_id=%s  %s=%.3f', sid, scfg['sensor_type'], value)
                if self.api.is_reachable():
                    if not self.api.post_reading(sid, value, ts):
                        self.queue.push(sid, value, ts)
                else:
                    self.queue.push(sid, value, ts)
            time.sleep(1)

    def _flush_loop(self):
        flush_interval = self.cfg.get('offline_flush_interval_seconds', 60)
        while not self._stop.is_set():
            time.sleep(flush_interval)
            if not self.api.is_reachable():
                continue
            batch = self.queue.pop_batch(50)
            if not batch:
                continue
            acked = []
            for item in batch:
                if self.api.post_reading(item['sensor_id'], item['value_raw'], item['reading_time']):
                    acked.append(item['_qid'])
            self.queue.ack(acked)
            if acked:
                log.info('Flushed %d queued readings to API', len(acked))

    def _schedule_loop(self):
        # Poll GET /farms/{id}/devices. The API embeds pending_command in each
        # device's config JSONB when an automation_rule or schedule fires.
        # Expected config keys: pending_command (str), pending_schedule_id (int|null)
        poll_interval = self.cfg.get('schedule_poll_interval_seconds', 30)
        while not self._stop.is_set():
            time.sleep(poll_interval)
            if not self.api.is_reachable():
                continue
            for device in self.api.get_devices():
                did      = device.get('id')
                config   = device.get('config') or {}
                cmd      = config.get('pending_command')
                sched_id = config.get('pending_schedule_id')
                if not cmd:
                    continue
                actuator = self._actuators.get(did)
                if not actuator:
                    log.debug('No local actuator for device_id=%s', did)
                    continue
                log.info('Executing scheduled command %r for device_id=%s', cmd, did)
                actuator.execute(cmd)
                self.api.post_actuator_event(
                    actuator_id=did, command=cmd,
                    source='schedule_trigger', schedule_id=sched_id,
                )
                self.api.patch_device_status(did, 'online')

    def run(self):
        log.info('gr33n Pi Client starting - farm_id=%s  api=%s',
                 self.cfg['farm']['farm_id'], self.cfg['api']['base_url'])
        threads = [
            threading.Thread(target=self._sensor_loop,   name='sensor-loop',   daemon=True),
            threading.Thread(target=self._flush_loop,    name='flush-loop',    daemon=True),
            threading.Thread(target=self._schedule_loop, name='schedule-loop', daemon=True),
        ]
        for t in threads:
            t.start()
        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            log.info('Shutdown requested.')
            self._stop.set()
            for t in threads:
                t.join(timeout=5)
            log.info('gr33n Pi Client stopped.')


if __name__ == '__main__':
    import argparse
    p = argparse.ArgumentParser(description='gr33n Pi sensor/actuator client')
    p.add_argument('--config', default='config.yaml')
    Gr33nPiClient(p.parse_args().config).run()
