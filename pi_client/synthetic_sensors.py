"""Phase 125 WS3 — synthetic sensor readings posted through the normal API path."""

from __future__ import annotations

import logging
import math
import threading
import time
from datetime import datetime, timezone
from typing import Callable, Optional

log = logging.getLogger('gr33n.synthetic')


def synthetic_value(entry: dict, t: float) -> float:
    """Compute a synthetic reading for demo / loopback modes.

    Modes:
      sine          — center ± amplitude over period_seconds
      hold          — fixed value
      step          — alternates low/high each half period
      demo_moisture — 3-minute cycle: in-band → drop below threshold → recover
    """
    mode = (entry.get('mode') or 'sine').lower()
    center = float(entry.get('center', 50))
    amplitude = float(entry.get('amplitude', 10))
    period = float(entry.get('period_seconds', 120))

    if mode == 'hold':
        return float(entry.get('value', center))
    if mode == 'step':
        phase = (t % period) / period if period > 0 else 0
        low = float(entry.get('low', center - amplitude))
        high = float(entry.get('high', center + amplitude))
        return low if phase < 0.5 else high
    if mode == 'demo_moisture':
        phase = (t % 180.0) / 180.0
        if phase < 0.35:
            return 55.0
        if phase < 0.55:
            return 55.0 - (phase - 0.35) / 0.2 * 35.0
        if phase < 0.75:
            return 20.0
        return 20.0 + (phase - 0.75) / 0.25 * 35.0
    return center + amplitude * math.sin(2 * math.pi * t / period)


class SyntheticSensorLoop:
    """Background publisher — POST /sensors/{id}/readings + local ReadingCache."""

    def __init__(
        self,
        entries: list,
        reading_cache,
        post_reading: Callable[[int, float, str], bool],
        queue_push: Callable[[int, float, str], None],
        is_reachable: Callable[[], bool],
        stop_event: threading.Event,
    ):
        self._entries = [e for e in entries if e.get('sensor_id') is not None]
        self._cache = reading_cache
        self._post = post_reading
        self._queue = queue_push
        self._reachable = is_reachable
        self._stop = stop_event
        self._thread: Optional[threading.Thread] = None
        self._last_wall: dict = {}
        self._started_mono = time.monotonic()
        self.sensor_ids = set()
        for e in self._entries:
            try:
                self.sensor_ids.add(int(e['sensor_id']))
            except (TypeError, ValueError):
                pass

    def start(self) -> None:
        if not self._entries:
            return
        self._thread = threading.Thread(target=self._loop, name='synthetic-sensors', daemon=True)
        self._thread.start()
        log.info('Synthetic sensor loop started for %d sensor(s)', len(self._entries))

    def stop(self) -> None:
        if self._thread:
            self._thread.join(timeout=3)

    def _loop(self) -> None:
        while not self._stop.is_set():
            now_wall = time.time()
            t = time.monotonic() - self._started_mono
            for entry in self._entries:
                try:
                    sid = int(entry['sensor_id'])
                except (TypeError, ValueError):
                    continue
                interval = float(entry.get('interval_seconds', 5))
                if now_wall - self._last_wall.get(sid, 0) < interval:
                    continue
                value = synthetic_value(entry, t)
                self._last_wall[sid] = now_wall
                self._cache.put(sid, value)
                ts = datetime.now(timezone.utc).isoformat()
                if self._reachable():
                    if not self._post(sid, value, ts):
                        self._queue(sid, value, ts)
                else:
                    self._queue(sid, value, ts)
                log.debug(
                    'synthetic sensor_id=%s value=%.3f mode=%s',
                    sid, value, entry.get('mode', 'sine'),
                )
            self._stop.wait(0.5)
