#!/usr/bin/env python3
"""Phase 51 WS5 — import pi_client/config.yaml wiring into the platform.

PATCHes sensors/{id}/wiring and actuators/{id}/wiring (JWT), then writes a
minimal bootstrap config without sensors[]/actuators[].

Example:
  python3 import_config_to_platform.py --config config.yaml \\
    --email dev@gr33n.local --password devpassword

Wiring PATCH requires dashboard JWT (not PI_API_KEY). Use --jwt to skip login.
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

import requests
import yaml

import gr33n_client as client


def _api_session(base_url: str, jwt: str | None, email: str | None, password: str | None) -> requests.Session:
    s = requests.Session()
    s.headers['Content-Type'] = 'application/json'
    if jwt:
        s.headers['Authorization'] = f'Bearer {jwt}'
        return s
    if not email or not password:
        raise SystemExit(
            'Wiring import requires dashboard JWT: pass --jwt or --email + --password')
    r = s.post(
        f'{base_url.rstrip("/")}/auth/login',
        json={'username': email, 'password': password},
        timeout=30,
    )
    if r.status_code != 200:
        raise SystemExit(f'login failed ({r.status_code}): {r.text[:300]}')
    token = r.json().get('token')
    if not token:
        raise SystemExit('login response missing token')
    s.headers['Authorization'] = f'Bearer {token}'
    return s


def _resolve_device_id(session: requests.Session, base_url: str, farm_id: int, device_uid: str) -> int:
    r = session.get(f'{base_url.rstrip("/")}/farms/{farm_id}/devices', timeout=30)
    if r.status_code != 200:
        raise SystemExit(f'list devices failed ({r.status_code}): {r.text[:300]}')
    devices = r.json()
    if not isinstance(devices, list):
        raise SystemExit('unexpected devices list response')
    for row in devices:
        if (row.get('device_uid') or '').strip() == device_uid:
            return int(row['id'])
    raise SystemExit(f'device_uid {device_uid!r} not found on farm {farm_id}')


def _patch_sensor_wiring(session, base_url: str, sensor_id: int, wiring: dict) -> None:
    r = session.patch(
        f'{base_url.rstrip("/")}/sensors/{sensor_id}/wiring',
        json={'wiring': wiring},
        timeout=30,
    )
    if r.status_code != 200:
        raise RuntimeError(f'sensor {sensor_id}: HTTP {r.status_code} {r.text[:200]}')


def _patch_actuator_wiring(session, base_url: str, actuator_id: int, wiring: dict) -> None:
    r = session.patch(
        f'{base_url.rstrip("/")}/actuators/{actuator_id}/wiring',
        json={'wiring': wiring},
        timeout=30,
    )
    if r.status_code != 200:
        raise RuntimeError(f'actuator {actuator_id}: HTTP {r.status_code} {r.text[:200]}')


def import_wiring(
    config_path: str,
    *,
    api_url: str | None = None,
    api_key: str | None = None,
    jwt: str | None = None,
    email: str | None = None,
    password: str | None = None,
    dry_run: bool = False,
    output_path: str | None = None,
) -> dict:
    """Import local YAML wiring to platform; return summary dict."""
    bootstrap = client.load_bootstrap(config_path)
    sensors = bootstrap.get('sensors') or []
    actuators = bootstrap.get('actuators') or []
    if not sensors and not actuators:
        raise SystemExit('config has no sensors[] or actuators[] — nothing to import')

    device_uid = (bootstrap.get('device') or {}).get('uid', '').strip()
    if not device_uid:
        raise SystemExit('device.uid is required in config.yaml for platform import')

    base_url = (api_url or bootstrap.get('api', {}).get('base_url', '')).strip()
    if not base_url:
        raise SystemExit('api.base_url missing — pass --api-url')

    farm_id = int(bootstrap.get('farm', {}).get('farm_id', 0))
    if farm_id <= 0:
        raise SystemExit('farm.farm_id is required')

    session = _api_session(base_url, jwt, email, password)
    device_id = _resolve_device_id(session, base_url, farm_id, device_uid)

    summary = {
        'device_uid': device_uid,
        'device_id': device_id,
        'sensors_imported': [],
        'actuators_imported': [],
        'errors': [],
    }

    for entry in sensors:
        sid = entry.get('sensor_id')
        if sid is None:
            summary['errors'].append('sensor entry missing sensor_id')
            continue
        wiring = client.pi_sensor_entry_to_wiring(entry, device_id)
        try:
            if not dry_run:
                _patch_sensor_wiring(session, base_url, int(sid), wiring)
            summary['sensors_imported'].append(int(sid))
        except RuntimeError as exc:
            summary['errors'].append(str(exc))

    for entry in actuators:
        aid = entry.get('actuator_id')
        if aid is None:
            summary['errors'].append('actuator entry missing actuator_id')
            continue
        wiring = client.pi_actuator_entry_to_wiring(entry)
        try:
            if not dry_run:
                _patch_actuator_wiring(session, base_url, int(aid), wiring)
            summary['actuators_imported'].append(int(aid))
        except RuntimeError as exc:
            summary['errors'].append(str(exc))

    if summary['errors'] and not dry_run:
        print('Import completed with errors:', file=sys.stderr)
        for err in summary['errors']:
            print(f'  - {err}', file=sys.stderr)

    minimal = client.build_minimal_bootstrap(bootstrap)
    if api_key is not None:
        minimal.setdefault('api', {})
        minimal['api']['api_key'] = api_key
    elif bootstrap.get('api', {}).get('api_key'):
        minimal.setdefault('api', {})
        minimal['api']['api_key'] = bootstrap['api']['api_key']

    out_path = Path(output_path or config_path)
    if not dry_run:
        with open(out_path, 'w') as fh:
            fh.write('# Imported to platform — wiring removed; Pi syncs from API on restart.\n')
            yaml.safe_dump(minimal, fh, sort_keys=False, default_flow_style=False)
        summary['written_config'] = str(out_path)
    else:
        summary['minimal_config_preview'] = minimal

    return summary


def main(argv: list[str] | None = None) -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--config', default='config.yaml', help='pi_client config.yaml to import')
    p.add_argument('--api-url', help='override api.base_url from config')
    p.add_argument('--api-key', help='api_key to write into minimal output YAML')
    p.add_argument('--jwt', help='dashboard JWT (skips /auth/login)')
    p.add_argument('--email', help='login email for JWT')
    p.add_argument('--password', help='login password for JWT')
    p.add_argument('--output', help='write minimal YAML here (default: overwrite --config)')
    p.add_argument('--dry-run', action='store_true', help='preview PATCHes without writing')
    args = p.parse_args(argv)

    summary = import_wiring(
        args.config,
        api_url=args.api_url,
        api_key=args.api_key,
        jwt=args.jwt,
        email=args.email,
        password=args.password,
        dry_run=args.dry_run,
        output_path=args.output,
    )

    print(json.dumps({
        'device_uid': summary['device_uid'],
        'device_id': summary['device_id'],
        'sensors_imported': summary['sensors_imported'],
        'actuators_imported': summary['actuators_imported'],
        'errors': summary['errors'],
        'written_config': summary.get('written_config'),
        'dry_run': args.dry_run,
    }, indent=2))

    if summary['errors']:
        return 1
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
