"""Phase 125 WS2 — NeoPixel/GPIO light driver for pre-plant simulation rig.

Maps sensor comfort bands and actuator states to LED colors per
docs/pi-light-simulation-mapping.md. Off-Pi runs use an in-memory strip stub.
"""

from __future__ import annotations

import logging
import math
import threading
import time
from typing import Callable, Optional

log = logging.getLogger('gr33n.light_sim')

# Sensor band colors (RGB 0-255) — match SensorTile + mapping spec.
COLOR_OK = (0, 180, 0)
COLOR_WARN = (255, 160, 0)
COLOR_ALERT_LOW = (0, 120, 255)
COLOR_ALERT_HIGH = (255, 0, 0)
COLOR_NO_DATA = (40, 40, 40)
COLOR_ACTUATOR_IDLE = (80, 80, 80)
COLOR_QUEUED = (255, 160, 0)
COLOR_FAULT = (255, 0, 255)
COLOR_ACTIVITY = (255, 255, 255)

ACTUATOR_TYPE_COLORS = {
    'light': (255, 200, 80),
    'pump': (60, 120, 255),
    'fan': (80, 220, 255),
    'valve': (180, 80, 255),
    'heater': (255, 80, 40),
}


def sensor_comfort_status(value: Optional[float], low: float, high: float) -> str:
    """Mirror ui/src/components/SensorTile.vue band logic."""
    if value is None:
        return 'no_data'
    if value < low:
        return 'alert_low'
    if value > high:
        return 'alert_high'
    span = high - low
    if span > 0 and (value < low + span * 0.15 or value > high - span * 0.15):
        return 'warn'
    return 'ok'


def _scale(rgb: tuple, brightness: float) -> tuple:
    b = max(0.0, min(1.0, float(brightness)))
    return tuple(int(c * b) for c in rgb)


def _blink_on(hz: float, now: float, duty: float = 0.5) -> bool:
    if hz <= 0:
        return True
    period = 1.0 / hz
    return (now % period) / period < duty


def rgb_for_sensor_status(status: str, now: float, brightness: float) -> tuple:
    if status == 'ok':
        return _scale(COLOR_OK, brightness * 0.6)
    if status == 'warn':
        return _scale(COLOR_WARN, brightness * 0.7)
    if status == 'alert_low':
        return _scale(COLOR_ALERT_LOW, brightness) if _blink_on(1.0, now) else (0, 0, 0)
    if status == 'alert_high':
        return _scale(COLOR_ALERT_HIGH, brightness) if _blink_on(1.0, now) else (0, 0, 0)
    # no_data — slow pulse
    pulse = 0.35 + 0.25 * (0.5 + 0.5 * math.sin(now * math.pi))
    return _scale(COLOR_NO_DATA, brightness * pulse)


def rgb_for_actuator(
    actuator_type: str,
    state: str,
    queued: bool,
    failed: bool,
    now: float,
    brightness: float,
) -> tuple:
    if failed:
        return _scale(COLOR_FAULT, brightness) if _blink_on(4.0, now, 0.5) else (0, 0, 0)
    if queued:
        return _scale(COLOR_QUEUED, brightness * 0.7) if _blink_on(1.0, now, 0.4) else (0, 0, 0)
    if state == 'on':
        base = ACTUATOR_TYPE_COLORS.get((actuator_type or '').lower(), (200, 200, 200))
        return _scale(base, brightness) if _blink_on(2.0, now) else (0, 0, 0)
    return _scale(COLOR_ACTUATOR_IDLE, brightness * 0.15)


class NeoPixelStrip:
    """WS2812 strip backend — stubs off-Pi, optional adafruit_neopixel on hardware."""

    def __init__(self, count: int, pin: int, brightness: float, pixel_order: str = 'GRB'):
        self.count = max(0, int(count))
        self.pin = int(pin)
        self.brightness = max(0.0, min(1.0, float(brightness)))
        self.pixel_order = pixel_order
        self._pixels = [(0, 0, 0)] * self.count
        self._hw = None
        try:
            import board  # noqa: F401
            import neopixel  # noqa: F401
            order = neopixel.GRB if pixel_order.upper() == 'GRB' else neopixel.RGB
            self._hw = neopixel.NeoPixel(getattr(board, f'D{pin}'), self.count, brightness=self.brightness, pixel_order=order, auto_write=False)
            log.info('NeoPixel strip: %d pixels on GPIO %s', self.count, pin)
        except Exception as exc:
            log.info('NeoPixel stub mode (%d pixels GPIO %s): %s', self.count, pin, exc)

    def set_pixel(self, index: int, rgb: tuple):
        if index < 0 or index >= self.count:
            return
        self._pixels[index] = rgb
        if self._hw is not None:
            self._hw[index] = rgb

    def show(self):
        if self._hw is not None:
            self._hw.show()

    def close(self):
        if self._hw is not None:
            for i in range(self.count):
                self._hw[i] = (0, 0, 0)
            self._hw.show()


class GpioIndicator:
    """Simple GPIO LED — uses gpiozero OutputDevice when available."""

    def __init__(self, pin: Optional[int]):
        self.pin = pin
        self._dev = None
        if pin is None:
            return
        try:
            from gpiozero import OutputDevice
            self._dev = OutputDevice(pin)
        except Exception as exc:
            log.debug('GPIO indicator stub pin=%s: %s', pin, exc)

    def set_on(self, on: bool):
        if self._dev is None:
            return
        if on:
            self._dev.on()
        else:
            self._dev.off()

    def close(self):
        if self._dev is not None:
            try:
                self._dev.off()
                self._dev.close()
            except Exception:
                pass


class LightSimulationDriver:
    """Polls local sensor cache + actuator state and drives LEDs."""

    def __init__(
        self,
        simulation_cfg: dict,
        reading_cache,
        get_actuators: Callable[[], dict],
        get_actuator_flags: Callable[[int], tuple],
        is_api_reachable: Callable[[], bool],
        get_activity_until: Callable[[], float],
    ):
        self.cfg = simulation_cfg or {}
        self._cache = reading_cache
        self._get_actuators = get_actuators
        self._get_actuator_flags = get_actuator_flags
        self._is_api_reachable = is_api_reachable
        self._get_activity_until = get_activity_until
        self._stop = threading.Event()
        self._thread: Optional[threading.Thread] = None

        neo = self.cfg.get('neopixel') or {}
        self._strip = NeoPixelStrip(
            count=neo.get('count', 8),
            pin=neo.get('pin', 18),
            brightness=neo.get('brightness', 0.4),
            pixel_order=neo.get('pixel_order', 'GRB'),
        )
        gpio = self.cfg.get('gpio_leds') or {}
        self._heartbeat = GpioIndicator(gpio.get('heartbeat_pin', 17))
        self._fault = GpioIndicator(gpio.get('fault_pin', 27))
        self._poll_s = float(self.cfg.get('poll_interval_seconds', 2))
        self._sensor_maps = list(self.cfg.get('sensors') or [])
        self._actuator_maps = list(self.cfg.get('actuators') or [])
        self._activity_pixel = self.cfg.get('activity_pixel')

    def start(self):
        if self._thread and self._thread.is_alive():
            return
        self._stop.clear()
        self._thread = threading.Thread(target=self._loop, name='light-simulation', daemon=True)
        self._thread.start()
        log.info(
            'Light simulation started — %d sensor LEDs, %d actuator LEDs',
            len(self._sensor_maps),
            len(self._actuator_maps),
        )

    def stop(self):
        self._stop.set()
        if self._thread:
            self._thread.join(timeout=5)
        self._strip.close()
        self._heartbeat.close()
        self._fault.close()

    def _loop(self):
        while not self._stop.is_set():
            try:
                self._refresh(time.monotonic())
            except Exception as exc:
                log.warning('light simulation refresh failed: %s', exc)
            self._stop.wait(self._poll_s)

    def _refresh(self, now: float):
        neo = self.cfg.get('neopixel') or {}
        brightness = float(neo.get('brightness', 0.4))

        for entry in self._sensor_maps:
            pixel = int(entry.get('pixel', -1))
            sid = entry.get('sensor_id')
            if sid is None:
                continue
            try:
                sid = int(sid)
            except (TypeError, ValueError):
                continue
            interval = float(entry.get('interval_seconds', 60))
            max_age = interval * 3
            value = self._cache.get(sid, max_age, now=now)
            low = float(entry.get('alert_threshold_low', 0))
            high = float(entry.get('alert_threshold_high', 100))
            status = sensor_comfort_status(value, low, high)
            rgb = rgb_for_sensor_status(status, now, brightness)
            self._strip.set_pixel(pixel, rgb)

        actuators = self._get_actuators()
        for entry in self._actuator_maps:
            pixel = int(entry.get('pixel', -1))
            aid = entry.get('actuator_id')
            if aid is None:
                continue
            try:
                aid = int(aid)
            except (TypeError, ValueError):
                continue
            act = actuators.get(aid)
            atype = entry.get('actuator_type') or (act.device_type if act else '')
            state = act.state if act else 'off'
            queued, failed = self._get_actuator_flags(aid)
            rgb = rgb_for_actuator(atype, state, queued, failed, now, brightness)
            self._strip.set_pixel(pixel, rgb)

        activity_until = self._get_activity_until()
        if self._activity_pixel is not None and now < activity_until:
            self._strip.set_pixel(int(self._activity_pixel), _scale(COLOR_ACTIVITY, brightness))

        self._strip.show()

        # Heartbeat: 1 Hz toggle when healthy
        self._heartbeat.set_on(_blink_on(1.0, now, 0.5))
        self._fault.set_on(not self._is_api_reachable())
