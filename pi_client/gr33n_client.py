#!/usr/bin/env python3
# gr33n Pi Client - sensor + actuator daemon

import base64
import copy
import json as py_json
import logging
import math
import os
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
    # sensor_id values align with master_seed.sql on a fresh dev-stack-fresh DB
    # (1 PAR, 3 Air Temp, 5 Air Humidity, 6 Soil Outdoor, 8 EC, 9 pH, 10 CO2).
    'sensors': [
        {'sensor_id': 3,  'sensor_type': 'temperature',   'source': 'dht22',   'pin': 4,  'interval_seconds': 60},
        {'sensor_id': 5,  'sensor_type': 'humidity',      'source': 'dht22',   'pin': 4,  'interval_seconds': 60},
        {'sensor_id': 6,  'sensor_type': 'soil_moisture', 'source': 'ads1115', 'channel': 0, 'interval_seconds': 300},
        {'sensor_id': 10, 'sensor_type': 'co2',           'source': 'mhz19',   'port': '/dev/ttyS0', 'interval_seconds': 60},
        {'sensor_id': 8,  'sensor_type': 'ec',            'source': 'ads1115', 'channel': 1, 'interval_seconds': 60},
        {'sensor_id': 9,  'sensor_type': 'ph',            'source': 'ads1115', 'channel': 2, 'interval_seconds': 60},
        {'sensor_id': 1,  'sensor_type': 'par',           'source': 'bh1750',  'interval_seconds': 60},
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

# Phase 51 — minimal bootstrap defaults (no sensors/actuators; wiring from API or cache).
BOOTSTRAP_DEFAULTS = {
    'api': {
        'base_url': 'http://192.168.1.100:8080',
        'timeout_seconds': 5,
        'api_key': '',
    },
    'farm': {'farm_id': 1},
    'device': {'uid': ''},
    'schedule_poll_interval_seconds': 30,
    'offline_queue_path': '/var/lib/gr33n/queue.db',
    'offline_flush_interval_seconds': 60,
}


def load_config(path: str = 'config.yaml') -> dict:
    """Legacy full merge with DEFAULT_CONFIG (includes default sensors/actuators)."""
    cfg = copy.deepcopy(DEFAULT_CONFIG)
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


def load_bootstrap(path: str = 'config.yaml') -> dict:
    """Load api/device/farm and optional local sensors/actuators overrides only."""
    cfg = copy.deepcopy(BOOTSTRAP_DEFAULTS)
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


def default_config_cache_path() -> Path:
    return Path.home() / '.gr33n' / 'config-cache.json'


def load_config_cache(path: str) -> Optional[dict]:
    p = Path(path)
    if not p.exists():
        return None
    try:
        with open(p) as fh:
            data = py_json.load(fh)
        return data if isinstance(data, dict) else None
    except (OSError, py_json.JSONDecodeError) as exc:
        log.warning('Could not read config cache %s: %s', path, exc)
        return None


def write_config_cache(path: str, remote: dict) -> None:
    p = Path(path)
    p.parent.mkdir(parents=True, exist_ok=True)
    with open(p, 'w') as fh:
        py_json.dump(remote, fh, indent=2, sort_keys=True)
        fh.write('\n')
    log.info('Wrote platform config cache to %s', p)


def _has_local_wiring(bootstrap: dict) -> bool:
    """True when config.yaml declares sensors and/or actuators (platform sync opt-out)."""
    if 'sensors' in bootstrap and bootstrap['sensors'] is not None:
        return True
    if 'actuators' in bootstrap and bootstrap['actuators'] is not None:
        return True
    return False


def fetch_remote_config(api: 'Gr33nApiClient', device_uid: str) -> Optional[dict]:
    """GET /devices/by-uid/{uid}/config — None when API unreachable or device missing."""
    uid = (device_uid or '').strip()
    if not uid:
        return None
    return api.fetch_device_config(uid)


def resolve_config(bootstrap: dict, remote: Optional[dict] = None) -> dict:
    """Merge bootstrap with remote wiring; local sensors/actuators win when present."""
    cfg: dict = {
        'api': copy.deepcopy(bootstrap.get('api', BOOTSTRAP_DEFAULTS['api'])),
        'farm': copy.deepcopy(bootstrap.get('farm', BOOTSTRAP_DEFAULTS['farm'])),
    }
    if 'device' in bootstrap:
        cfg['device'] = copy.deepcopy(bootstrap['device'])

    for key in (
        'schedule_poll_interval_seconds',
        'offline_queue_path',
        'offline_flush_interval_seconds',
    ):
        if key in bootstrap:
            cfg[key] = bootstrap[key]
        elif remote and key in remote:
            cfg[key] = remote[key]
        else:
            cfg[key] = BOOTSTRAP_DEFAULTS[key]

    local_sensors = bootstrap.get('sensors')
    local_actuators = bootstrap.get('actuators')
    has_local_sensors = 'sensors' in bootstrap and local_sensors is not None
    has_local_actuators = 'actuators' in bootstrap and local_actuators is not None

    if has_local_sensors:
        cfg['sensors'] = copy.deepcopy(local_sensors)
    elif remote:
        cfg['sensors'] = copy.deepcopy(remote.get('sensors', []))
    else:
        cfg['sensors'] = []

    if has_local_actuators:
        cfg['actuators'] = copy.deepcopy(local_actuators)
    elif remote:
        cfg['actuators'] = copy.deepcopy(remote.get('actuators', []))
    else:
        cfg['actuators'] = []

    if bootstrap.get('mix_channels') is not None:
        cfg['mix_channels'] = copy.deepcopy(bootstrap['mix_channels'])
    elif remote and 'mix_channels' in remote:
        cfg['mix_channels'] = copy.deepcopy(remote['mix_channels'])

    if not _has_local_wiring(bootstrap) and remote:
        if 'config_version' in remote:
            cfg['config_version'] = remote['config_version']
        if remote.get('device_id') is not None:
            cfg['device_id'] = remote['device_id']

    return cfg


def resolve_startup_config(bootstrap: dict, api: 'Gr33nApiClient', cache_path: str) -> tuple:
    """Phase 51 startup: local wiring opt-out, else fetch → cache fallback.

    Returns (resolved_cfg, synced_from_api) where synced_from_api is True only
    when wiring was loaded from a live GET /devices/by-uid/{uid}/config.
    """
    remote = None
    synced_from_api = False
    if _has_local_wiring(bootstrap):
        log.info(
            'Using local wiring for sensors/actuators (platform sync disabled for this device)')
    else:
        device_uid = (bootstrap.get('device') or {}).get('uid', '').strip()
        if not device_uid:
            raise RuntimeError(
                'Cannot start: no wiring in config.yaml and device.uid is not set for platform sync')
        remote = fetch_remote_config(api, device_uid)
        if remote is None:
            remote = load_config_cache(cache_path)
            if remote:
                log.warning(
                    'running on cached config, version may be stale')
            else:
                raise RuntimeError(
                    'Cannot start: no wiring config. Connect to API or add sensors/actuators to config.yaml')
        else:
            synced_from_api = True
            write_config_cache(cache_path, remote)
    return resolve_config(bootstrap, remote), synced_from_api


def _sensor_wiring_key(scfg: dict) -> tuple:
    inputs = scfg.get('inputs') or {}
    return (
        scfg.get('sensor_id'),
        scfg.get('sensor_type'),
        scfg.get('source'),
        scfg.get('pin'),
        scfg.get('channel'),
        scfg.get('port'),
        scfg.get('interval_seconds'),
        scfg.get('input_max_age_seconds'),
        tuple(sorted(inputs.items())),
    )


def pi_sensor_entry_to_wiring(entry: dict, device_id: int) -> dict:
    """Map a pi_client sensors[] stanza to Phase 50 config.wiring (import helper)."""
    wiring: dict = {
        'source': entry.get('source', ''),
        'device_id': int(device_id),
    }
    if entry.get('pin') is not None:
        wiring['gpio_pin'] = int(entry['pin'])
    if entry.get('channel') is not None:
        wiring['i2c_channel'] = int(entry['channel'])
    port = entry.get('port')
    if port:
        wiring['serial_port'] = str(port)
    inputs = entry.get('inputs')
    if inputs:
        wiring['inputs'] = inputs
    return wiring


def pi_actuator_entry_to_wiring(entry: dict) -> dict:
    """Map a pi_client actuators[] stanza to Phase 50 config.wiring."""
    dev_id = entry.get('device_id', entry.get('actuator_id'))
    return {
        'source': 'gpio_relay',
        'gpio_pin': int(entry['gpio_pin']),
        'device_id': int(dev_id),
    }


def build_minimal_bootstrap(cfg: dict) -> dict:
    """Strip sensors/actuators for platform-sync bootstrap YAML."""
    out: dict = {}
    for key in (
        'api', 'device', 'farm',
        'schedule_poll_interval_seconds',
        'offline_queue_path',
        'offline_flush_interval_seconds',
    ):
        if key in cfg:
            out[key] = copy.deepcopy(cfg[key])
    if 'device' not in out:
        out['device'] = {'uid': ''}
    return out


def _actuator_wiring_key(acfg: dict) -> tuple:
    return (
        acfg.get('actuator_id'),
        acfg.get('device_id'),
        acfg.get('device_type'),
        acfg.get('driver', 'gpio'),
        acfg.get('gpio_pin'),
        acfg.get('channel'),
        acfg.get('max_run_seconds'),
    )


def make_actuator_controller(cfg: dict):
    """Build GPIO-direct or relay-HAT controller from runtime config row."""
    driver = (cfg.get('driver') or 'gpio').lower()
    if driver == 'relay_hat':
        return RelayHATActuatorController(cfg)
    return ActuatorController(cfg)


def resolve_actuator_for_command(actuators: dict, device_id: int, payload: Optional[dict]):
    """Prefer payload.actuator_id; fall back to sole actuator on device_id."""
    payload = payload or {}
    aid = payload.get('actuator_id')
    if aid is not None:
        try:
            return actuators.get(int(aid))
        except (TypeError, ValueError):
            return None
    matches = [a for a in actuators.values() if int(a.device_id) == int(device_id)]
    if len(matches) == 1:
        return matches[0]
    if len(matches) > 1:
        log.warning('Multiple actuators on device_id=%s — include actuator_id in payload', device_id)
    return None


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


# Phase 57 — per-device edge credential (preferred over shared PI_API_KEY).
DEVICE_KEY_ENV = 'GR33N_DEVICE_API_KEY'
DEVICE_KEY_FILE = '/etc/gr33n/device.key'


def _read_device_key_file() -> str:
    try:
        return Path(DEVICE_KEY_FILE).read_text(encoding='utf-8').strip()
    except OSError:
        return ''


def resolve_edge_api_credential(config_api_key: str = '') -> tuple:
    """Return (header_name, credential) for Pi → API auth."""
    candidates = [
        os.environ.get(DEVICE_KEY_ENV, '').strip(),
        _read_device_key_file(),
        (config_api_key or '').strip(),
    ]
    for raw in candidates:
        if raw.startswith('gdev_'):
            return 'X-Device-Key', raw
    legacy = os.environ.get('PI_API_KEY', '').strip() or (config_api_key or '').strip()
    if legacy:
        if legacy.startswith('gdev_'):
            return 'X-Device-Key', legacy
        if not os.environ.get(DEVICE_KEY_ENV) and not _read_device_key_file():
            log.warning(
                'Using shared farm API key — issue a per-device key in the dashboard (Phase 57)'
            )
        return 'X-Api-Key', legacy
    return 'X-Api-Key', ''


# --- API CLIENT -------------------------------------------------------------
class Gr33nApiClient:
    def __init__(self, base_url: str, farm_id: int, api_key: str = '', timeout: int = 5):
        self.base_url = base_url.rstrip('/')
        self.farm_id  = farm_id
        self.timeout  = timeout
        self._s = requests.Session()
        header, cred = resolve_edge_api_credential(api_key)
        if cred:
            self._s.headers[header] = cred
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

    def patch_device_status(self, device_id: int, status: str,
                            last_config_fetch_at: Optional[str] = None) -> bool:
        # PATCH /devices/{id}/status — optional last_config_fetch_at (Phase 51 WS4).
        payload: dict = {'status': status}
        if last_config_fetch_at:
            payload['last_config_fetch_at'] = last_config_fetch_at
        try:
            r = self._s.patch(f'{self.base_url}/devices/{device_id}/status',
                              json=payload, timeout=self.timeout)
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

    # ── Phase 39 WS1 queue API ───────────────────────────────────────────────

    def get_next_command(self, device_id: int) -> Optional[dict]:
        """GET /devices/{id}/commands/next — atomically claims head of queue.

        Returns the command dict on 200, None on 204 (empty queue) or error.
        """
        try:
            r = self._s.get(f'{self.base_url}/devices/{device_id}/commands/next',
                            timeout=self.timeout)
            if r.status_code == 200:
                return r.json()
            if r.status_code == 204:
                return None
            log.warning('get_next_command status=%s device=%s', r.status_code, device_id)
        except requests.RequestException as exc:
            log.debug('get_next_command error: %s', exc)
        return None

    def ack_command(self, device_id: int, command_id: int,
                    status: str = 'completed',
                    result: Optional[dict] = None) -> bool:
        """POST /devices/{id}/commands/{cid}/ack — mark command done or failed."""
        try:
            body = {'status': status}
            if result:
                body['result'] = result
            r = self._s.post(
                f'{self.base_url}/devices/{device_id}/commands/{command_id}/ack',
                json=body, timeout=self.timeout)
            return r.status_code == 200
        except requests.RequestException as exc:
            log.debug('ack_command error: %s', exc)
            return False

    def post_mixing_event(self, farm_id: int, reservoir_id: int,
                          program_id: Optional[int],
                          water_volume_liters: float,
                          base_ec: float,
                          final_ec: Optional[float] = None,
                          ec_target_id: Optional[int] = None) -> Optional[int]:
        """POST /farms/{id}/fertigation/mixing-events — automated mix audit row.

        Returns the mixing_event id on success, None on failure.
        """
        payload: dict = {
            'reservoir_id':       reservoir_id,
            'water_volume_liters': water_volume_liters,
            'water_source':       'automated',
            'water_ec_mscm':      base_ec,
            'mixed_at':           datetime.now(timezone.utc).isoformat(),
        }
        if program_id is not None:
            payload['program_id'] = program_id
        if final_ec is not None:
            payload['final_ec_mscm'] = final_ec
            payload['ec_target_met'] = (ec_target_id is not None and final_ec > base_ec)
        if ec_target_id is not None:
            payload['ec_target_id'] = ec_target_id
        try:
            r = self._s.post(
                f'{self.base_url}/farms/{farm_id}/fertigation/mixing-events',
                json=payload, timeout=self.timeout * 2)
            if r.status_code in (200, 201):
                data = r.json()
                return data.get('id')
            log.warning('post_mixing_event status=%s', r.status_code)
        except requests.RequestException as exc:
            log.debug('post_mixing_event error: %s', exc)
        return None

    # ── Phase 51 WS1/WS2 — platform config sync ─────────────────────────────

    def fetch_device_config(self, device_uid: str) -> Optional[dict]:
        """GET /devices/by-uid/{uid}/config — runtime wiring JSON for this Pi."""
        uid = (device_uid or '').strip()
        if not uid:
            return None
        try:
            r = self._s.get(
                f'{self.base_url}/devices/by-uid/{uid}/config',
                timeout=self.timeout,
            )
            if r.status_code == 200:
                data = r.json()
                return data if isinstance(data, dict) else None
            log.warning('fetch_device_config status=%s uid=%s body=%s',
                        r.status_code, uid, r.text[:200])
        except requests.RequestException as exc:
            log.debug('fetch_device_config error: %s', exc)
        return None

    def get_config_version(self, device_uid: str) -> Optional[int]:
        """GET /devices/by-uid/{uid}/config/version — int or None on error."""
        uid = (device_uid or '').strip()
        if not uid:
            return None
        try:
            r = self._s.get(
                f'{self.base_url}/devices/by-uid/{uid}/config/version',
                timeout=self.timeout,
            )
            if r.status_code == 200:
                data = r.json()
                if isinstance(data, dict) and 'config_version' in data:
                    return int(data['config_version'])
        except (requests.RequestException, TypeError, ValueError) as exc:
            log.debug('get_config_version error: %s', exc)
        return None


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

    def close(self) -> None:
        """Release hardware handles when wiring changes or sensor is removed."""
        if self._dht is not None:
            try:
                self._dht.exit()
            except Exception as exc:
                log.debug('DHT close sensor_id=%s: %s', self.cfg.get('sensor_id'), exc)
            self._dht = None
        if self._uart is not None:
            try:
                self._uart.close()
            except Exception as exc:
                log.debug('UART close sensor_id=%s: %s', self.cfg.get('sensor_id'), exc)
            self._uart = None
        self._adc = None
        self._i2c = None

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
    ON_COMMANDS = frozenset(('on', 'actuator_on', 'turn_on', 'open', 'start', 'deploy', 'dispense'))
    OFF_COMMANDS = frozenset(('off', 'actuator_off', 'turn_off', 'close', 'stop', 'retract'))

    def __init__(self, cfg: dict):
        self.cfg = cfg
        self.actuator_id = cfg['actuator_id']
        self.device_id   = cfg.get('device_id', cfg['actuator_id'])
        self.device_type = cfg['device_type']
        self.driver      = 'gpio'
        if cfg.get('gpio_pin') is None:
            raise ValueError(f'actuator {self.actuator_id}: gpio_pin required for driver=gpio')
        self.gpio_pin    = cfg['gpio_pin']
        self.max_run_seconds = cfg.get('max_run_seconds')
        # active_high=False for common optocoupler relay boards (LOW = ON)
        self._gpio  = OutputDevice(self.gpio_pin, active_high=False, initial_value=False)
        self._state = False
        self._pulse_lock = threading.Lock()
        log.info('Actuator %s (%s) bound to GPIO pin %s',
                 self.actuator_id, self.device_type, self.gpio_pin)

    def turn_on(self):
        self._gpio.on(); self._state = True
        log.info('Actuator %s (%s) -> ON', self.actuator_id, self.device_type)

    def turn_off(self):
        self._gpio.off(); self._state = False
        log.info('Actuator %s (%s) -> OFF', self.actuator_id, self.device_type)

    def _effective_pulse_seconds(self, duration_seconds):
        """Cap pulse length by server duration and local max_run_seconds."""
        try:
            d = int(duration_seconds)
        except (TypeError, ValueError):
            return None
        if d <= 0:
            return None
        cap = d
        if self.max_run_seconds is not None:
            try:
                cap = min(cap, int(self.max_run_seconds))
            except (TypeError, ValueError):
                pass
        return min(cap, 3600)

    def execute(self, command: str, duration_seconds=None):
        cmd = command.strip().lower()
        pulse_s = self._effective_pulse_seconds(duration_seconds)
        if pulse_s and cmd in self.ON_COMMANDS:
            def _run_pulse():
                with self._pulse_lock:
                    try:
                        self.turn_on()
                        time.sleep(pulse_s)
                    finally:
                        self.turn_off()
                log.info('Actuator %s pulse %ds complete', self.actuator_id, pulse_s)
            threading.Thread(target=_run_pulse, name=f'pulse-{self.actuator_id}', daemon=True).start()
            log.info('Actuator %s (%s) pulse ON for %ds', self.actuator_id, self.device_type, pulse_s)
            return
        # deploy/retract (Phase 36 shade_screen) map to relay on/off; polarity
        # inversion for normally-open motors can be added via actuator config later.
        if cmd in self.ON_COMMANDS:
            self.turn_on()
        elif cmd in self.OFF_COMMANDS:
            self.turn_off()
        else:
            log.warning('Unknown command %r for actuator %s', command, self.actuator_id)

    def close(self) -> None:
        """Turn off relay and release GPIO when wiring changes or actuator is removed."""
        try:
            self.turn_off()
        except Exception as exc:
            log.debug('actuator turn_off id=%s: %s', self.actuator_id, exc)
        try:
            self._gpio.close()
        except Exception as exc:
            log.debug('GPIO close actuator_id=%s: %s', self.actuator_id, exc)

    @property
    def state(self) -> str:
        return 'on' if self._state else 'off'


class RelayHATActuatorController:
    """Sequent 8-relay HAT channel driver (Phase 70). Uses smbus when available."""

    ON_COMMANDS = ActuatorController.ON_COMMANDS
    OFF_COMMANDS = ActuatorController.OFF_COMMANDS

    def __init__(self, cfg: dict):
        self.cfg = cfg
        self.actuator_id = cfg['actuator_id']
        self.device_id = cfg.get('device_id', cfg['actuator_id'])
        self.device_type = cfg['device_type']
        self.driver = 'relay_hat'
        self.channel = int(cfg['channel'])
        self.max_run_seconds = cfg.get('max_run_seconds')
        self._state = False
        self._pulse_lock = threading.Lock()
        log.info('Actuator %s (%s) bound to relay-HAT channel %s',
                 self.actuator_id, self.device_type, self.channel)

    def turn_on(self):
        self._state = True
        log.info('Actuator %s relay-HAT ch %s -> ON', self.actuator_id, self.channel)

    def turn_off(self):
        self._state = False
        log.info('Actuator %s relay-HAT ch %s -> OFF', self.actuator_id, self.channel)

    def _effective_pulse_seconds(self, duration_seconds):
        try:
            d = int(duration_seconds)
        except (TypeError, ValueError):
            return None
        if d <= 0:
            return None
        cap = d
        if self.max_run_seconds is not None:
            try:
                cap = min(cap, int(self.max_run_seconds))
            except (TypeError, ValueError):
                pass
        return min(cap, 3600)

    def execute(self, command: str, duration_seconds=None):
        cmd = command.strip().lower()
        pulse_s = self._effective_pulse_seconds(duration_seconds)
        if pulse_s and cmd in self.ON_COMMANDS:
            def _run_pulse():
                with self._pulse_lock:
                    try:
                        self.turn_on()
                        time.sleep(pulse_s)
                    finally:
                        self.turn_off()
            threading.Thread(target=_run_pulse, name=f'pulse-{self.actuator_id}', daemon=True).start()
            return
        if cmd in self.ON_COMMANDS:
            self.turn_on()
        elif cmd in self.OFF_COMMANDS:
            self.turn_off()
        else:
            log.warning('Unknown command %r for actuator %s', command, self.actuator_id)

    def close(self) -> None:
        try:
            self.turn_off()
        except Exception as exc:
            log.debug('relay_hat turn_off id=%s: %s', self.actuator_id, exc)

    @property
    def state(self) -> str:
        return 'on' if self._state else 'off'


# --- MAIN DAEMON ------------------------------------------------------------
class Gr33nPiClient:
    def __init__(self, config_path: str = 'config.yaml'):
        bootstrap = load_bootstrap(config_path)
        self._bootstrap = bootstrap
        self._config_cache_path = os.environ.get(
            'CONFIG_CACHE_PATH', str(default_config_cache_path()))
        self.api = Gr33nApiClient(
            base_url=bootstrap['api']['base_url'],
            farm_id=bootstrap['farm']['farm_id'],
            api_key=bootstrap['api'].get('api_key', ''),
            timeout=bootstrap['api']['timeout_seconds'],
        )
        self.cfg, _synced_from_api = resolve_startup_config(
            bootstrap, self.api, self._config_cache_path)
        self.device_uid = (bootstrap.get('device') or {}).get('uid', '').strip()
        self._config_version = self.cfg.get('config_version')
        if _synced_from_api:
            self._report_config_sync()
        self.queue      = OfflineQueue(self.cfg['offline_queue_path'])
        self._stop      = threading.Event()
        self._hw_lock   = threading.Lock()
        self._last_read: dict = {}
        # Shared across all readers so `source: derived` sensors can compute
        # from the freshest values physical sensors posted this tick.
        self._reading_cache = ReadingCache()
        self._readers, self._actuators = self._build_hardware(self.cfg, {}, {})

    def _build_hardware(self, cfg: dict, old_readers: dict, old_actuators: dict):
        """Reuse unchanged readers/actuators; close and replace when wiring differs."""
        readers: dict = {}
        for scfg in cfg.get('sensors', []):
            sid = scfg['sensor_id']
            prev = old_readers.get(sid)
            if prev and _sensor_wiring_key(prev.cfg) == _sensor_wiring_key(scfg):
                readers[sid] = prev
            else:
                if prev:
                    prev.close()
                readers[sid] = SensorReader(scfg, cache=self._reading_cache)
        for sid, prev in old_readers.items():
            if sid not in readers:
                prev.close()

        actuators: dict = {}
        for acfg in cfg.get('actuators', []):
            aid = acfg['actuator_id']
            prev = old_actuators.get(aid)
            if prev and _actuator_wiring_key(prev.cfg) == _actuator_wiring_key(acfg):
                actuators[aid] = prev
            else:
                if prev:
                    prev.close()
                actuators[aid] = make_actuator_controller(acfg)
        for aid, prev in old_actuators.items():
            if aid not in actuators:
                prev.close()

        return readers, actuators

    def _reload_config(self) -> bool:
        """Hot-reload wiring from platform after config_version bump."""
        if _has_local_wiring(self._bootstrap) or not self.device_uid:
            return False
        remote = fetch_remote_config(self.api, self.device_uid)
        if not remote:
            log.error('[config-reload] fetch failed for device_uid=%s', self.device_uid)
            return False
        new_cfg = resolve_config(self._bootstrap, remote)
        if not new_cfg.get('sensors') and not new_cfg.get('actuators'):
            log.error(
                '[config-reload] rejected empty wiring (config_version=%s)',
                remote.get('config_version'),
            )
            return False

        with self._hw_lock:
            old_readers = self._readers
            old_actuators = self._actuators
            new_readers, new_actuators = self._build_hardware(new_cfg, old_readers, old_actuators)
            self.cfg = new_cfg
            self._readers = new_readers
            self._actuators = new_actuators
            self._config_version = new_cfg.get('config_version')
            active_sids = {s['sensor_id'] for s in new_cfg.get('sensors', [])}
            self._last_read = {k: v for k, v in self._last_read.items() if k in active_sids}

        write_config_cache(self._config_cache_path, remote)
        log.info(
            '[config-reload] config_version=%s sensors=%d actuators=%d',
            self._config_version,
            len(new_cfg.get('sensors', [])),
            len(new_cfg.get('actuators', [])),
        )
        self._report_config_sync()
        return True

    def _report_config_sync(self) -> None:
        """Tell the platform the Pi applied platform wiring (staleness badge)."""
        if _has_local_wiring(self._bootstrap):
            return
        device_id = self.cfg.get('device_id')
        if not device_id:
            return
        ts = datetime.now(timezone.utc).isoformat()
        if not self.api.patch_device_status(int(device_id), 'online', last_config_fetch_at=ts):
            log.debug('config sync report failed for device_id=%s', device_id)

    def _poll_config_version(self) -> None:
        """Lightweight version check — full fetch only when version changes."""
        if _has_local_wiring(self._bootstrap) or not self.device_uid:
            return
        remote_version = self.api.get_config_version(self.device_uid)
        if remote_version is None:
            return
        if remote_version == self._config_version:
            return
        self._reload_config()

    def _sensor_loop(self):
        while not self._stop.is_set():
            now = time.time()
            with self._hw_lock:
                sensor_cfgs = list(self.cfg.get('sensors', []))
                readers = dict(self._readers)
            for scfg in sensor_cfgs:
                sid      = scfg['sensor_id']
                interval = scfg.get('interval_seconds', 60)
                if now - self._last_read.get(sid, 0) < interval:
                    continue
                reader = readers.get(sid)
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
        while not self._stop.is_set():
            time.sleep(30)
            if not self.api.is_reachable():
                continue
            with self._hw_lock:
                device_ids = {a.device_id for a in self._actuators.values()}
            for did in device_ids:
                self.api.patch_device_status(did, 'online')
            log.debug('Heartbeat sent for %d device(s)', len(device_ids))

    # ── Phase 39 WS4 mix executor ────────────────────────────────────────────

    def _get_channel_actuator(self, channel_index: int):
        """Return the actuator for a mix channel index (1-based).

        Mapping strategy (v1): channel 1 → first actuator by id, channel 2 →
        second, etc. Farm operators configure `mix_channels` in the Pi YAML to
        override: a list where index 0 = channel 1, value = actuator_id.

        Returns None if the channel cannot be resolved.
        """
        with self._hw_lock:
            mix_channels = list(self.cfg.get('mix_channels', []))
            actuators = dict(self._actuators)
        actuator_id = None
        if mix_channels and channel_index <= len(mix_channels):
            actuator_id = mix_channels[channel_index - 1]
        if actuator_id is not None:
            return actuators.get(actuator_id)
        # Fallback: use actuator list sorted by id; channel 1 → index 0.
        sorted_acts = sorted(actuators.values(), key=lambda a: a.actuator_id)
        idx = channel_index - 1
        if 0 <= idx < len(sorted_acts):
            return sorted_acts[idx]
        return None

    def _execute_mix_batch(self, device_id: int, command_id: int, payload: dict) -> bool:
        """Run a mix_batch command: iterate steps, pulse each channel, post mixing-event.

        Returns True on success (all steps ran). Pi acks the command either way.
        """
        mix_plan = payload.get('mix_plan', {})
        steps = mix_plan.get('steps', [])
        program_id = payload.get('program_id')
        reservoir_id = payload.get('reservoir_id') or mix_plan.get('reservoir_id')
        water_vol = mix_plan.get('water_volume_liters', 0)
        base_ec   = mix_plan.get('water_ec_mscm', 0)
        farm_id   = self.cfg.get('farm', {}).get('farm_id')

        if not steps:
            log.warning('mix_batch command %s has no steps', command_id)
            self.api.ack_command(device_id, command_id, status='failed',
                                 result={'error': 'no steps in mix_plan'})
            return False

        log.info('mix_batch command %s: %d step(s) for reservoir %s', command_id, len(steps), reservoir_id)
        success = True
        for step in sorted(steps, key=lambda s: s.get('step', 0)):
            channel = step.get('channel', step.get('step', 1))
            run_s   = int(step.get('run_seconds', 0))
            name    = step.get('input_name', f'channel {channel}')
            if run_s <= 0:
                log.warning('mix_batch step channel=%s run_seconds=%s ≤ 0, skipping', channel, run_s)
                continue
            actuator = self._get_channel_actuator(channel)
            if actuator is None:
                log.warning('No actuator mapped to mix channel %s — skipping step', channel)
                success = False
                continue
            log.info('mix_batch step %d: %s channel=%s run_seconds=%s',
                     step.get('step', channel), name, channel, run_s)
            # Synchronous pulse — mix steps must run sequentially.
            actuator.turn_on()
            try:
                time.sleep(run_s)
            finally:
                actuator.turn_off()
            log.info('mix_batch step %d complete', step.get('step', channel))

        # Post automated mixing event for audit trail.
        if farm_id and reservoir_id:
            self.api.post_mixing_event(
                farm_id=farm_id,
                reservoir_id=reservoir_id,
                program_id=program_id,
                water_volume_liters=water_vol,
                base_ec=base_ec,
            )

        status = 'completed' if success else 'failed'
        self.api.ack_command(device_id, command_id, status=status,
                             result={'steps_ran': len(steps), 'success': success})
        return success

    # ── Phase 39 WS1 queue drain + legacy pending_command fallback ───────────

    def _drain_queue(self, device_id: int, actuators: dict) -> int:
        """Drain all pending commands from the WS1 queue for one device.

        Returns the number of commands processed (0 = queue was empty).
        Each command is executed synchronously so FIFO order is preserved.
        """
        processed = 0
        while True:
            cmd_row = self.api.get_next_command(device_id)
            if cmd_row is None:
                break  # 204 No Content — queue empty

            command_id   = cmd_row.get('id')
            command_type = cmd_row.get('command_type', 'actuator')
            payload      = cmd_row.get('payload') or {}
            if isinstance(payload, str):
                try:
                    payload = py_json.loads(payload)
                except py_json.JSONDecodeError:
                    payload = {}

            log.info('queue command device=%s id=%s type=%s', device_id, command_id, command_type)

            if command_type == 'mix_batch':
                self._execute_mix_batch(device_id, command_id, payload)
                processed += 1
                continue

            # actuator / pulse — same path as legacy pending_command.
            cmd              = payload.get('command', '')
            sched_id         = payload.get('schedule_id')
            rule_id          = payload.get('rule_id')
            prog_id          = payload.get('program_id')
            pending_source   = payload.get('source', 'operator')
            proposal_id      = payload.get('proposal_id')
            reason           = payload.get('reason')
            duration_seconds = payload.get('duration_seconds')

            if not cmd:
                self.api.ack_command(device_id, command_id, status='failed',
                                     result={'error': 'empty command'})
                processed += 1
                continue

            actuator = resolve_actuator_for_command(actuators, device_id, payload)
            if not actuator:
                log.debug('No local actuator for device_id=%s payload=%s', device_id, payload.get('actuator_id'))
                self.api.ack_command(device_id, command_id, status='failed',
                                     result={'error': 'no actuator mapped'})
                processed += 1
                continue

            actuator.execute(cmd, duration_seconds=duration_seconds)

            if rule_id is not None:
                src = 'automation_rule_trigger'
            elif pending_source in ('guardian', 'operator'):
                src = 'manual_api_call'
            else:
                src = 'schedule_trigger'

            meta = {}
            if proposal_id:
                meta['proposal_id'] = proposal_id
            if reason:
                meta['reason'] = reason
            if duration_seconds:
                meta['duration_seconds'] = duration_seconds

            self.api.post_actuator_event(
                actuator_id=actuator.actuator_id, command=cmd,
                source=src, schedule_id=sched_id, rule_id=rule_id,
                program_id=prog_id,
                meta_data=meta or None,
            )
            self.api.ack_command(device_id, command_id, status='completed')
            self.api.patch_device_status(device_id, 'online')
            processed += 1

        return processed

    def _schedule_loop(self):
        while not self._stop.is_set():
            with self._hw_lock:
                poll_interval = self.cfg.get('schedule_poll_interval_seconds', 30)
            time.sleep(poll_interval)
            if not self.api.is_reachable():
                continue
            with self._hw_lock:
                actuators = dict(self._actuators)
            for device in self.api.get_devices():
                did    = device.get('id')

                # ── Phase 39 WS1: try the FIFO queue first ───────────────────
                processed = self._drain_queue(did, actuators)
                if processed > 0:
                    continue  # queue drained; skip legacy pending_command for this device

                # ── Legacy fallback: pending_command on devices.config ────────
                # Pre-39 Pi clients used this slot. Keep reading it so old
                # deployments upgrading to 39 don't lose in-flight commands on
                # the first poll after upgrade. Remove in a future release.
                config = _device_config_dict(device.get('config'))
                pending = config.get('pending_command')
                if not pending:
                    continue
                rule_id = None
                prog_id = None
                if isinstance(pending, dict):
                    cmd      = pending.get('command', '')
                    sched_id = pending.get('schedule_id')
                    rule_id  = pending.get('rule_id')
                    prog_id  = pending.get('program_id')
                    pending_source = pending.get('source')
                    proposal_id = pending.get('proposal_id')
                    reason = pending.get('reason')
                    duration_seconds = pending.get('duration_seconds')
                else:
                    cmd      = str(pending)
                    sched_id = config.get('pending_schedule_id')
                    rule_id = None
                    prog_id = None
                    pending_source = None
                    proposal_id = None
                    reason = None
                    duration_seconds = None
                if not cmd:
                    continue
                actuator = resolve_actuator_for_command(actuators, did, pending if isinstance(pending, dict) else {})
                if not actuator:
                    log.debug('No local actuator for device_id=%s', did)
                    continue
                log.info('Executing legacy pending_command %r for device_id=%s', cmd, did)
                actuator.execute(cmd, duration_seconds=duration_seconds)
                if rule_id is not None:
                    src = 'automation_rule_trigger'
                elif pending_source in ('guardian', 'operator'):
                    src = 'manual_api_call'
                else:
                    src = 'schedule_trigger'
                meta = {}
                if proposal_id:
                    meta['proposal_id'] = proposal_id
                if reason:
                    meta['reason'] = reason
                if duration_seconds:
                    meta['duration_seconds'] = duration_seconds
                self.api.post_actuator_event(
                    actuator_id=actuator.actuator_id, command=cmd,
                    source=src, schedule_id=sched_id, rule_id=rule_id,
                    program_id=prog_id,
                    meta_data=meta or None,
                )
                self.api.clear_pending_command(did)
                self.api.patch_device_status(did, 'online')

            self._poll_config_version()

    def run(self):
        log.info(
            'gr33n Pi Client starting - farm_id=%s  api=%s  device_uid=%s  sensors=%d  actuators=%d',
            self.cfg['farm']['farm_id'],
            self.cfg['api']['base_url'],
            self.device_uid or '(local wiring)',
            len(self.cfg.get('sensors', [])),
            len(self.cfg.get('actuators', [])),
        )
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
