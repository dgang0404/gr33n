#!/usr/bin/env python3
# gr33n Pi Client - sensor + actuator daemon

import base64
import json as py_json
import logging
import math
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


def _device_config_dict(raw) -> dict:
    """Decode ``device['config']`` from GET /farms/{id}/devices.

    Go's ``encoding/json`` marshals ``[]byte`` as a base64 *string*; the Pi
    must decode that string to recover ``pending_command`` and other keys.
    If the server already returns an object (tests / future encoder), pass
    it through.
    """
    if raw is None:
        return {}
    if isinstance(raw, dict):
        return raw
    if isinstance(raw, str):
        try:
            b = base64.b64decode(raw)
            return py_json.loads(b.decode('utf-8'))
        except (ValueError, py_json.JSONDecodeError, UnicodeDecodeError):
            return {}
    return {}


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
        {'actuator_id': 1, 'device_id': 1, 'device_type': 'light',      'gpio_pin': 17},
        {'actuator_id': 2, 'device_id': 2, 'device_type': 'irrigation', 'gpio_pin': 27},
        {'actuator_id': 3, 'device_id': 3, 'device_type': 'fan',        'gpio_pin': 22},
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
                '    created_at   TEXT    DEFAULT CURRENT_TIMESTAMP'
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

    def post_readings_batch(self, items: list) -> bool:
        """POST /sensors/readings/batch — items are dicts with sensor_id, value_raw, optional reading_time, is_valid."""
        if not items:
            return True
        try:
            r = self._s.post(f'{self.base_url}/sensors/readings/batch', json=items, timeout=self.timeout)
            if r.status_code in (200, 201):
                return True
            log.warning('API rejected batch ingest status=%s body=%s', r.status_code, r.text[:200])
            return False
        except requests.RequestException as exc:
            log.debug('API unreachable: %s', exc)
            return False

    def get_devices(self) -> list:
        # GET /farms/{id}/devices — JWT or X-API-Key (Pi); see mqtt-edge-operator-playbook.md
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
                             schedule_id: Optional[int] = None,
                             rule_id: Optional[int] = None,
                             program_id: Optional[int] = None,
                             meta_data: Optional[dict] = None,
                             parameters_sent: Optional[dict] = None) -> bool:
        """Report command execution to the API (Pi feedback).

        Pass through provenance from ``pending_command`` so actuator_events
        rows join back to schedules, rules, and fertigation programs for
        audit trails. ``rule_id`` and ``program_id`` are mutually exclusive
        on the server; schedule-bound program fires include both
        ``schedule_id`` and ``program_id``.
        """
        payload = {
            'command_sent':     command,
            'source':           source,
            'event_time':       datetime.now(timezone.utc).isoformat(),
            'execution_status': 'command_sent_to_device',
        }
        if schedule_id is not None:
            payload['triggered_by_schedule_id'] = schedule_id
        if rule_id is not None:
            payload['triggered_by_rule_id'] = rule_id
        if program_id is not None:
            payload['program_id'] = program_id
        if meta_data:
            payload['meta_data'] = meta_data
        if parameters_sent:
            payload['parameters_sent'] = parameters_sent
        try:
            r = self._s.post(f'{self.base_url}/actuators/{actuator_id}/events',
                             json=payload, timeout=self.timeout)
            return r.status_code in (200, 201)
        except requests.RequestException:
            return False

    def clear_pending_command(self, device_id: int) -> bool:
        try:
            r = self._s.delete(f'{self.base_url}/devices/{device_id}/pending-command',
                               timeout=self.timeout)
            return r.status_code in (200, 204)
        except requests.RequestException:
            return False


# --- DERIVED SENSOR SUPPORT -------------------------------------------------
# A `source: derived` sensor computes its value from other sensors on the
# same Pi (e.g. dew_point from temperature + humidity). The physical readers
# below write their latest value into a shared ReadingCache after every
# successful read; derived readers query that cache instead of talking to
# hardware. Keeping computation on the edge means derived channels keep
# working when the network is flaky, and the backend never has to know the
# difference — dew_point is ingested exactly like temperature would be.

DEFAULT_DERIVED_INPUT_MAX_AGE_SECONDS = 120


class ReadingCache:
    """Thread-safe most-recent-value cache keyed by sensor_id.

    `put` records (value, monotonic_timestamp) so staleness checks are
    wall-clock-independent. `get` returns None when no reading is cached or
    when the cached reading is older than `max_age_s`. The `now` kwarg on
    both methods is for deterministic tests; production code leaves it
    unset and lets `time.monotonic()` flow.
    """

    def __init__(self):
        self._data: dict = {}
        self._lock = threading.Lock()

    def put(self, sensor_id: int, value: float, now: Optional[float] = None):
        ts = now if now is not None else time.monotonic()
        with self._lock:
            self._data[sensor_id] = (float(value), ts)

    def get(self, sensor_id: int, max_age_s: float, now: Optional[float] = None) -> Optional[float]:
        ts_now = now if now is not None else time.monotonic()
        with self._lock:
            rec = self._data.get(sensor_id)
        if rec is None:
            return None
        value, ts = rec
        if ts_now - ts > max_age_s:
            return None
        return value


def compute_dew_point_c(t_c: float, rh_pct: float) -> float:
    """Magnus-Tetens dew-point approximation (valid 0-60°C, RH > 1%).

    Reference: August-Roche-Magnus formula. Output °C.
    """
    if rh_pct <= 0:
        # Log-of-zero would blow up; treat as "extremely dry" floor.
        rh_pct = 0.01
    a, b = 17.625, 243.04
    gamma = math.log(rh_pct / 100.0) + (a * t_c) / (b + t_c)
    return round((b * gamma) / (a - gamma), 2)


def compute_vpd_kpa(t_c: float, rh_pct: float) -> float:
    """Vapour Pressure Deficit in kPa, leaf-temperature approximation.

    VPD = SVP * (1 - RH/100), with SVP via Tetens over water. Output kPa,
    rounded to 3 decimals — the resolution growers actually tune against.
    """
    svp = 0.6108 * math.exp((17.27 * t_c) / (t_c + 237.3))
    return round(svp * (1.0 - rh_pct / 100.0), 3)


def compute_heat_index_c(t_c: float, rh_pct: float) -> float:
    """Rothfusz heat-index regression, converted to Celsius.

    The regression is defined in Fahrenheit; below 80°F (~26.7°C) the NWS
    says "use the dry-bulb temperature" because the regression diverges.
    We follow that convention — low-temperature callers just get `t_c` back.
    """
    t_f = t_c * 9.0 / 5.0 + 32.0
    if t_f < 80.0:
        return round(t_c, 2)
    hi_f = (
        -42.379
        + 2.04901523 * t_f
        + 10.14333127 * rh_pct
        - 0.22475541 * t_f * rh_pct
        - 0.00683783 * t_f * t_f
        - 0.05481717 * rh_pct * rh_pct
        + 0.00122874 * t_f * t_f * rh_pct
        + 0.00085282 * t_f * rh_pct * rh_pct
        - 0.00000199 * t_f * t_f * rh_pct * rh_pct
    )
    return round((hi_f - 32.0) * 5.0 / 9.0, 2)


_DERIVED_COMPUTERS = {
    'dew_point':   compute_dew_point_c,
    'vpd':         compute_vpd_kpa,
    'heat_index':  compute_heat_index_c,
}


# --- SENSOR READER ----------------------------------------------------------
class SensorReader:
    def __init__(self, cfg: dict, cache: Optional[ReadingCache] = None):
        self.cfg = cfg
        self.cache = cache
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
        elif src == 'derived':
            # Derived sensors have no hardware — they consume other sensors'
            # cached readings. Validate the shape early so a misconfigured
            # entry fails loudly at daemon start, not silently at tick time.
            stype = self.cfg.get('sensor_type', '')
            if stype not in _DERIVED_COMPUTERS:
                log.warning("derived sensor %s has unsupported sensor_type=%r "
                            "(expected one of %s)",
                            self.cfg.get('sensor_id'), stype,
                            sorted(_DERIVED_COMPUTERS.keys()))
            inputs = self.cfg.get('inputs') or {}
            for key in ('temperature_c', 'humidity_pct'):
                if not isinstance(inputs.get(key), int):
                    log.warning("derived sensor %s missing required input %r "
                                "(expected an integer sensor_id)",
                                self.cfg.get('sensor_id'), key)

    def read(self) -> Optional[float]:
        src   = self.cfg.get('source', '')
        stype = self.cfg.get('sensor_type', '')
        if src == 'derived':
            return self._read_derived(stype)
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

    def _read_derived(self, stype: str) -> Optional[float]:
        """Compute a derived sensor value from other cached readings.

        Returns None (not a mock) when any input is missing or stale —
        emitting an invented dew_point when the temp sensor is dead would
        masquerade as a real reading and silently defeat the alert pipeline.
        """
        computer = _DERIVED_COMPUTERS.get(stype)
        if computer is None:
            return None
        if self.cache is None:
            log.debug("derived sensor %s has no shared cache bound — returning None",
                      self.cfg.get('sensor_id'))
            return None
        inputs = self.cfg.get('inputs') or {}
        max_age = float(self.cfg.get('input_max_age_seconds',
                                     DEFAULT_DERIVED_INPUT_MAX_AGE_SECONDS))
        t_sid = inputs.get('temperature_c')
        rh_sid = inputs.get('humidity_pct')
        if not (isinstance(t_sid, int) and isinstance(rh_sid, int)):
            return None
        t_c = self.cache.get(t_sid, max_age)
        rh = self.cache.get(rh_sid, max_age)
        if t_c is None or rh is None:
            log.debug("derived sensor %s (%s) skipped: stale/missing inputs "
                      "(temperature_c sid=%s → %s, humidity_pct sid=%s → %s, max_age=%ss)",
                      self.cfg.get('sensor_id'), stype, t_sid, t_c, rh_sid, rh, max_age)
            return None
        try:
            return computer(t_c, rh)
        except (ValueError, OverflowError) as exc:
            log.warning("derived sensor %s (%s) compute failed: %s (t=%s rh=%s)",
                        self.cfg.get('sensor_id'), stype, exc, t_c, rh)
            return None

    @staticmethod
    def _mock(stype: str) -> float:
        return {'temperature': 22.5, 'humidity': 58.0, 'soil_moisture': 42.0,
                'co2': 820.0, 'ec': 1.4, 'ph': 6.2, 'par': 380.0}.get(stype, 0.0)


# --- ACTUATOR CONTROLLER ----------------------------------------------------
class ActuatorController:
    def __init__(self, cfg: dict):
        self.actuator_id = cfg['actuator_id']
        self.device_id   = cfg.get('device_id', cfg['actuator_id'])
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
        # Shared across all readers so `source: derived` sensors can compute
        # from the freshest values physical sensors posted this tick.
        self._reading_cache = ReadingCache()
        self._readers: dict = {
            s['sensor_id']: SensorReader(s, cache=self._reading_cache)
            for s in self.cfg['sensors']
        }
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
                # Feed the shared cache so derived sensors iterated later in
                # this same tick pick up fresh inputs. Config convention:
                # list derived sensors AFTER their source sensors.
                self._reading_cache.put(sid, value)
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

    def _heartbeat_loop(self):
        device_ids = {a.device_id for a in self._actuators.values()}
        while not self._stop.is_set():
            time.sleep(30)
            if not self.api.is_reachable():
                continue
            for did in device_ids:
                self.api.patch_device_status(did, 'online')
            log.debug('Heartbeat sent for %d device(s)', len(device_ids))

    def _schedule_loop(self):
        poll_interval = self.cfg.get('schedule_poll_interval_seconds', 30)
        actuator_by_device = {a.device_id: a for a in self._actuators.values()}
        while not self._stop.is_set():
            time.sleep(poll_interval)
            if not self.api.is_reachable():
                continue
            for device in self.api.get_devices():
                did    = device.get('id')
                config = _device_config_dict(device.get('config'))
                pending = config.get('pending_command')
                if not pending:
                    continue
                rule_id = None
                prog_id = None
                if isinstance(pending, dict):
                    cmd      = pending.get('command', '')
                    sched_id = pending.get('schedule_id')
                    rule_id = pending.get('rule_id')
                    prog_id = pending.get('program_id')
                else:
                    cmd      = str(pending)
                    sched_id = config.get('pending_schedule_id')
                if not cmd:
                    continue
                actuator = actuator_by_device.get(did)
                if not actuator:
                    log.debug('No local actuator for device_id=%s', did)
                    continue
                log.info('Executing scheduled command %r for device_id=%s', cmd, did)
                actuator.execute(cmd)
                src = 'automation_rule_trigger' if rule_id is not None else 'schedule_trigger'
                self.api.post_actuator_event(
                    actuator_id=actuator.actuator_id, command=cmd,
                    source=src, schedule_id=sched_id, rule_id=rule_id,
                    program_id=prog_id,
                )
                self.api.clear_pending_command(did)
                self.api.patch_device_status(did, 'online')

    def run(self):
        log.info('gr33n Pi Client starting - farm_id=%s  api=%s',
                 self.cfg['farm']['farm_id'], self.cfg['api']['base_url'])
        threads = [
            threading.Thread(target=self._sensor_loop,    name='sensor-loop',    daemon=True),
            threading.Thread(target=self._flush_loop,     name='flush-loop',     daemon=True),
            threading.Thread(target=self._schedule_loop,  name='schedule-loop',  daemon=True),
            threading.Thread(target=self._heartbeat_loop, name='heartbeat-loop', daemon=True),
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
