#!/usr/bin/env python3
"""Phase 125 WS4 — scripted demo: drop Media Moisture and watch the rig react.

Posts synthetic readings through the normal API path (no bypass). Run while
pi_client is in simulation mode on a Pi with the LED strip wired per
docs/pi-light-simulation-mapping.md.

Usage:
  export PI_API_KEY=...   # or GR33N_DEVICE_API_KEY from /etc/gr33n/device.key
  python3 run_demo_moisture_loop.py --base-url http://localhost:8080 --farm-id 1
"""

from __future__ import annotations

import argparse
import os
import sys
import time
from typing import Optional

import requests

# Resolve sensor id by name (stable on demo farm 1 after seed).
DEFAULT_SENSOR_NAME = 'Media Moisture Indoor'


def resolve_api_key() -> str:
    for env in ('GR33N_DEVICE_API_KEY', 'PI_API_KEY'):
        v = os.environ.get(env, '').strip()
        if v:
            return v
    return ''


def find_sensor_id(base_url: str, farm_id: int, name: str, headers: dict) -> int:
    r = requests.get(f'{base_url.rstrip("/")}/farms/{farm_id}/sensors', headers=headers, timeout=10)
    r.raise_for_status()
    for row in r.json():
        if row.get('name') == name:
            return int(row['id'])
    raise SystemExit(f'sensor {name!r} not found on farm {farm_id}')


def post_reading(base_url: str, sensor_id: int, value: float, headers: dict) -> None:
    payload = {'sensor_id': sensor_id, 'value_raw': value, 'is_valid': True}
    r = requests.post(
        f'{base_url.rstrip("/")}/sensors/{sensor_id}/readings',
        json=payload,
        headers=headers,
        timeout=10,
    )
    if r.status_code not in (200, 201):
        raise RuntimeError(f'POST reading failed {r.status_code}: {r.text[:200]}')


def run_scenario(
    base_url: str,
    farm_id: int,
    sensor_name: str,
    interval: float,
    sensor_id: Optional[int] = None,
) -> None:
    key = resolve_api_key()
    if not key:
        print('Set PI_API_KEY or GR33N_DEVICE_API_KEY', file=sys.stderr)
        sys.exit(1)
    header = 'X-Device-Key' if key.startswith('gdev_') else 'X-API-Key'
    headers = {header: key, 'Content-Type': 'application/json'}

    if sensor_id is None:
        sid = find_sensor_id(base_url, farm_id, sensor_name, headers)
    else:
        sid = sensor_id
    print(f'Demo moisture loop — sensor_id={sid} ({sensor_name})')
    print('Phase 1: in-band (55%) …')
    for _ in range(6):
        post_reading(base_url, sid, 55.0, headers)
        time.sleep(interval)

    print('Phase 2: drop below alert threshold (25%) …')
    for v in (45.0, 35.0, 28.0, 22.0, 18.0):
        post_reading(base_url, sid, v, headers)
        print(f'  posted {v}%')
        time.sleep(interval)

    print('Phase 3: hold low — watch pump LED + UI alert …')
    for _ in range(8):
        post_reading(base_url, sid, 18.0, headers)
        time.sleep(interval)

    print('Phase 4: recover in-band …')
    for v in (30.0, 42.0, 55.0):
        post_reading(base_url, sid, v, headers)
        print(f'  posted {v}%')
        time.sleep(interval)

    print('Done — check pixel 0 (moisture) and pixel 6 (pump) on the strip.')


def main() -> None:
    p = argparse.ArgumentParser(description='Phase 125 moisture drop demo script')
    p.add_argument('--base-url', default=os.environ.get('GR33N_API_URL', 'http://localhost:8080'))
    p.add_argument('--farm-id', type=int, default=1)
    p.add_argument('--sensor-name', default=DEFAULT_SENSOR_NAME)
    p.add_argument('--sensor-id', type=int, default=None,
                   help='skip name lookup (Pi API key cannot GET /farms/{id}/sensors)')
    p.add_argument('--interval', type=float, default=5.0, help='seconds between POSTs')
    args = p.parse_args()
    run_scenario(args.base_url, args.farm_id, args.sensor_name, args.interval, args.sensor_id)


if __name__ == '__main__':
    main()
