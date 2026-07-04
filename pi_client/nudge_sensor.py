#!/usr/bin/env python3
"""Phase 125 WS3 — manually nudge one sensor reading through the normal API path.

Usage:
  python3 nudge_sensor.py --sensor-id 7 --value 18
  PI_API_KEY=... python3 nudge_sensor.py --sensor-id 7 --value 55 --base-url http://localhost:8080
"""

from __future__ import annotations

import argparse
import os
import sys
from datetime import datetime, timezone

import requests


def resolve_api_key() -> str:
    for env in ('GR33N_DEVICE_API_KEY', 'PI_API_KEY'):
        v = os.environ.get(env, '').strip()
        if v:
            return v
    return ''


def main() -> None:
    p = argparse.ArgumentParser(description='Post a single synthetic sensor reading')
    p.add_argument('--sensor-id', type=int, required=True)
    p.add_argument('--value', type=float, required=True)
    p.add_argument('--base-url', default=os.environ.get('GR33N_API_URL', 'http://localhost:8080'))
    args = p.parse_args()

    key = resolve_api_key()
    if not key:
        print('Set PI_API_KEY or GR33N_DEVICE_API_KEY', file=sys.stderr)
        sys.exit(1)
    header = 'X-Device-Key' if key.startswith('gdev_') else 'X-API-Key'
    headers = {header: key, 'Content-Type': 'application/json'}
    payload = {
        'sensor_id': args.sensor_id,
        'value_raw': args.value,
        'reading_time': datetime.now(timezone.utc).isoformat(),
        'is_valid': True,
    }
    url = f'{args.base_url.rstrip("/")}/sensors/{args.sensor_id}/readings'
    r = requests.post(url, json=payload, headers=headers, timeout=10)
    if r.status_code not in (200, 201):
        print(f'POST failed {r.status_code}: {r.text[:300]}', file=sys.stderr)
        sys.exit(1)
    print(f'Posted sensor_id={args.sensor_id} value={args.value}')


if __name__ == '__main__':
    main()
