#!/usr/bin/env python3
"""
MQTT → gr33n API telemetry bridge (Phase 14 WS1).

Subscribes to gr33n/<farm_id>/<device_uid>/telemetry/<slug> (configurable prefix),
maps (device_uid, slug) to sensor_id via YAML, POSTs to /sensors/readings/batch.

Run beside Mosquitto or any MQTT broker; MCUs publish only to MQTT — the bridge
holds X-API-Key. See docs/mqtt-edge-operator-playbook.md.

Environment (typical):
  GR33N_API_URL, GR33N_FARM_ID, PI_API_KEY (or MQTT_BRIDGE_API_KEY),
  MQTT_HOST, MQTT_PORT, optional MQTT_USER / MQTT_PASS,
  MQTT_SENSOR_MAP_PATH (YAML), optional MQTT_TOPIC_PREFIX (default gr33n).
"""

from __future__ import annotations

import argparse
import json
import logging
import os
import signal
import sys
import threading
from typing import Any, Optional

import requests
import yaml

try:
    import paho.mqtt.client as mqtt
except ImportError:  # pragma: no cover
    mqtt = None  # type: ignore

LOG = logging.getLogger("gr33n.mqtt_bridge")

# Must stay in sync with internal/handler/sensor/handler.go maxBatchReadings
_MAX_BATCH = 64


def parse_telemetry_topic(
    topic: str, expected_farm_id: int, prefix: str = "gr33n"
) -> Optional[tuple[str, str]]:
    """
    Expect topic: <prefix>/<farm_id>/<device_uid>/telemetry/<slug>
    slug may contain slashes (extra path segments joined).
    Returns (device_uid, slug) or None if pattern / farm id mismatch.
    """
    parts = topic.split("/")
    if len(parts) < 5:
        return None
    if parts[0] != prefix or parts[3] != "telemetry":
        return None
    try:
        farm_in_topic = int(parts[1])
    except ValueError:
        return None
    if farm_in_topic != expected_farm_id:
        LOG.warning(
            "dropping message: topic farm_id=%s != GR33N_FARM_ID=%s (%s)",
            farm_in_topic,
            expected_farm_id,
            topic,
        )
        return None
    device_uid = parts[2]
    slug = "/".join(parts[4:])
    if not slug:
        return None
    return device_uid, slug


def extract_value(payload: bytes) -> Optional[float]:
    """Parse MCU payload: JSON {v|value_raw|value}, or plain float string."""
    if not payload:
        return None
    s = payload.decode("utf-8", errors="replace").strip()
    if not s:
        return None
    try:
        obj = json.loads(s)
        if isinstance(obj, dict):
            for k in ("v", "value_raw", "value"):
                if k in obj and obj[k] is not None:
                    return float(obj[k])
            return None
        if isinstance(obj, (int, float)):
            return float(obj)
        return None
    except (json.JSONDecodeError, TypeError, ValueError):
        pass
    try:
        return float(s)
    except ValueError:
        return None


def load_sensor_map(path: str) -> dict[tuple[str, str], int]:
    with open(path, encoding="utf-8") as f:
        data = yaml.safe_load(f)
    return sensor_map_from_data(data)


def sensor_map_from_data(data: Any) -> dict[tuple[str, str], int]:
    if not data or not isinstance(data, dict):
        return {}
    out: dict[tuple[str, str], int] = {}
    for row in data.get("sensor_map", []) or []:
        if not isinstance(row, dict):
            continue
        du = str(row.get("device_uid", "")).strip()
        slug = str(row.get("slug", "")).strip()
        sid = row.get("sensor_id")
        if not du or not slug or sid is None:
            continue
        out[(du, slug)] = int(sid)
    return out


def resolve_sensor_id(
    mapping: dict[tuple[str, str], int], device_uid: str, slug: str
) -> Optional[int]:
    if slug.isdigit():
        return int(slug)
    return mapping.get((device_uid, slug))


def _api_key() -> str:
    return (
        os.environ.get("MQTT_BRIDGE_API_KEY", "").strip()
        or os.environ.get("PI_API_KEY", "").strip()
    )


class Bridge:
    def __init__(
        self,
        *,
        api_url: str,
        farm_id: int,
        api_key: str,
        sensor_map: dict[tuple[str, str], int],
        topic_prefix: str = "gr33n",
        batch_ms: int = 0,
    ):
        self.api_url = api_url.rstrip("/")
        self.farm_id = farm_id
        self.sensor_map = sensor_map
        self.topic_prefix = topic_prefix
        self.batch_ms = max(0, batch_ms)
        self._session = requests.Session()
        self._session.headers.update(
            {"X-API-Key": api_key, "Content-Type": "application/json"}
        )
        self._lock = threading.Lock()
        self._buf: list[dict[str, Any]] = []
        self._timer: Optional[threading.Timer] = None

    def _post_batch(self, items: list[dict[str, Any]]) -> None:
        if not items:
            return
        try:
            r = self._session.post(
                f"{self.api_url}/sensors/readings/batch",
                json=items,
                timeout=30,
            )
            if r.status_code not in (200, 201):
                LOG.warning(
                    "batch ingest failed status=%s body=%s",
                    r.status_code,
                    (r.text or "")[:500],
                )
        except requests.RequestException as exc:
            LOG.warning("batch ingest error: %s", exc)

    def flush_pending(self) -> None:
        """Drain debounced readings (e.g. before process exit)."""
        self._flush()

    def _flush(self) -> None:
        with self._lock:
            if self._timer is not None:
                self._timer.cancel()
                self._timer = None
            batch = self._buf
            self._buf = []
        while batch:
            chunk = batch[:_MAX_BATCH]
            del batch[:_MAX_BATCH]
            self._post_batch(chunk)

    def _arm_debounce_timer_locked(self) -> None:
        if self._timer is not None:
            self._timer.cancel()
        t = threading.Timer(self.batch_ms / 1000.0, self._flush)
        t.daemon = True
        self._timer = t
        t.start()

    def enqueue_reading(self, sensor_id: int, value: float) -> None:
        item = {"sensor_id": sensor_id, "value_raw": value, "is_valid": True}
        if self.batch_ms <= 0:
            self._post_batch([item])
            return
        to_send: Optional[list[dict[str, Any]]] = None
        with self._lock:
            self._buf.append(item)
            if len(self._buf) >= _MAX_BATCH:
                if self._timer is not None:
                    self._timer.cancel()
                    self._timer = None
                to_send = self._buf
                self._buf = []
            else:
                self._arm_debounce_timer_locked()
        if to_send is not None:
            while to_send:
                chunk = to_send[:_MAX_BATCH]
                del to_send[:_MAX_BATCH]
                self._post_batch(chunk)

    def on_message(self, _c: Any, _u: Any, msg: Any) -> None:
        topic = getattr(msg, "topic", "") or ""
        parsed = parse_telemetry_topic(
            topic, self.farm_id, prefix=self.topic_prefix
        )
        if not parsed:
            return
        device_uid, slug = parsed
        sid = resolve_sensor_id(self.sensor_map, device_uid, slug)
        if sid is None:
            LOG.debug("no sensor mapping for device_uid=%s slug=%s", device_uid, slug)
            return
        val = extract_value(msg.payload)
        if val is None:
            LOG.warning("unparseable payload on %s", topic)
            return
        self.enqueue_reading(sid, val)


def main() -> None:
    logging.basicConfig(
        level=os.environ.get("LOG_LEVEL", "INFO"),
        format="%(asctime)s %(levelname)s %(message)s",
    )
    if mqtt is None:
        LOG.error("paho-mqtt is required: pip install paho-mqtt")
        sys.exit(1)

    parser = argparse.ArgumentParser(description="gr33n MQTT telemetry bridge")
    parser.add_argument(
        "--sensor-map",
        default=os.environ.get("MQTT_SENSOR_MAP_PATH", ""),
        help="YAML sensor map (or set MQTT_SENSOR_MAP_PATH)",
    )
    parser.add_argument(
        "--topic-prefix",
        default=os.environ.get("MQTT_TOPIC_PREFIX", "gr33n"),
        help="First segment of topics (default gr33n)",
    )
    parser.add_argument(
        "--batch-ms",
        type=int,
        default=int(os.environ.get("MQTT_BATCH_MS", "0")),
        help="If >0, coalesce readings for this many ms (max %d per POST)" % _MAX_BATCH,
    )
    args = parser.parse_args()

    api_url = os.environ.get("GR33N_API_URL", "").strip()
    farm_s = os.environ.get("GR33N_FARM_ID", "").strip()
    key = _api_key()
    map_path = args.sensor_map.strip()

    if not api_url or not farm_s or not key or not map_path:
        LOG.error(
            "Set GR33N_API_URL, GR33N_FARM_ID, PI_API_KEY (or MQTT_BRIDGE_API_KEY), "
            "and --sensor-map / MQTT_SENSOR_MAP_PATH"
        )
        sys.exit(1)

    farm_id = int(farm_s)
    sensor_map = load_sensor_map(map_path)
    if not sensor_map:
        LOG.warning("sensor_map is empty — only numeric telemetry slugs will work")

    bridge = Bridge(
        api_url=api_url,
        farm_id=farm_id,
        api_key=key,
        sensor_map=sensor_map,
        topic_prefix=args.topic_prefix.strip() or "gr33n",
        batch_ms=args.batch_ms,
    )

    host = os.environ.get("MQTT_HOST", "127.0.0.1").strip()
    port = int(os.environ.get("MQTT_PORT", "1883"))
    user = os.environ.get("MQTT_USER", "").strip() or None
    password = os.environ.get("MQTT_PASS", "").strip() or None
    use_tls = os.environ.get("MQTT_USE_TLS", "").strip() in ("1", "true", "yes")

    topic = f"{bridge.topic_prefix}/+/+/telemetry/#"
    client = mqtt.Client(mqtt.CallbackAPIVersion.VERSION2)
    if user:
        client.username_pw_set(user, password)
    if use_tls:
        ca = os.environ.get("MQTT_CA_FILE", "").strip() or None
        client.tls_set(ca_certs=ca or None)

    client.on_message = bridge.on_message

    stop = threading.Event()

    def handle_sig(*_a: Any) -> None:
        stop.set()
        bridge.flush_pending()
        try:
            client.disconnect()
        except Exception:
            pass

    signal.signal(signal.SIGINT, handle_sig)
    signal.signal(signal.SIGTERM, handle_sig)

    client.connect(host, port, keepalive=60)
    client.subscribe(topic, qos=1)
    LOG.info(
        "subscribed %s → %s farm_id=%s (batch_ms=%s)",
        topic,
        api_url,
        farm_id,
        bridge.batch_ms,
    )
    client.loop_forever()


if __name__ == "__main__":
    main()
