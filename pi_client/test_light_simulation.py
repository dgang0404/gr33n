#!/usr/bin/env python3
"""Tests for Phase 125 light simulation driver."""

import os
import sys
import time
import unittest

sys.path.insert(0, os.path.dirname(__file__))

import gr33n_client as client
import light_simulation as ls


class TestSensorComfortStatus(unittest.TestCase):
    def test_ok_center(self):
        self.assertEqual(ls.sensor_comfort_status(50.0, 25.0, 80.0), 'ok')

    def test_warn_near_low_edge(self):
        # 15% of span (55) = 8.25; low + 8.25 = 33.25
        self.assertEqual(ls.sensor_comfort_status(33.0, 25.0, 80.0), 'warn')

    def test_alert_low(self):
        self.assertEqual(ls.sensor_comfort_status(10.0, 25.0, 80.0), 'alert_low')

    def test_alert_high(self):
        self.assertEqual(ls.sensor_comfort_status(90.0, 25.0, 80.0), 'alert_high')

    def test_no_data(self):
        self.assertEqual(ls.sensor_comfort_status(None, 0.0, 100.0), 'no_data')


class TestSimulationActuator(unittest.TestCase):
    def test_no_gpio_side_effects(self):
        act = client.SimulationActuatorController({
            'actuator_id': 1,
            'device_id': 10,
            'device_type': 'pump',
        })
        act.execute('on')
        self.assertEqual(act.state, 'on')
        act.execute('off')
        self.assertEqual(act.state, 'off')


class TestLightSimulationDriver(unittest.TestCase):
    def test_sensor_pixel_turns_green_when_reading_in_band(self):
        cache = client.ReadingCache()
        cache.put(7, 50.0, now=100.0)
        strip_pixels = {}

        class StubStrip:
            count = 8

            def set_pixel(self, index, rgb):
                strip_pixels[index] = rgb

            def show(self):
                pass

            def close(self):
                pass

        sim_cfg = {
            'poll_interval_seconds': 0.05,
            'neopixel': {'count': 8, 'brightness': 1.0},
            'sensors': [{
                'sensor_id': 7,
                'pixel': 0,
                'alert_threshold_low': 25,
                'alert_threshold_high': 80,
                'interval_seconds': 120,
            }],
            'actuators': [],
        }

        driver = ls.LightSimulationDriver(
            simulation_cfg=sim_cfg,
            reading_cache=cache,
            get_actuators=lambda: {},
            get_actuator_flags=lambda _aid: (False, False),
            is_api_reachable=lambda: True,
            get_activity_until=lambda: 0.0,
        )
        driver._strip = StubStrip()
        driver._heartbeat = ls.GpioIndicator(None)
        driver._fault = ls.GpioIndicator(None)
        driver._refresh(100.0)

        self.assertEqual(strip_pixels[0], ls._scale(ls.COLOR_OK, 0.6))

    def test_actuator_on_blinks_type_color(self):
        act = client.SimulationActuatorController({
            'actuator_id': 2,
            'device_id': 10,
            'device_type': 'pump',
        })
        act.turn_on()
        rgb_on = ls.rgb_for_actuator('pump', 'on', False, False, 0.0, 1.0)
        rgb_off_phase = ls.rgb_for_actuator('pump', 'on', False, False, 0.3, 1.0)
        self.assertNotEqual(rgb_on, (0, 0, 0))
        self.assertEqual(rgb_off_phase, (0, 0, 0))
        self.assertEqual(act.state, 'on')


class TestResolveConfigSimulation(unittest.TestCase):
    def test_simulation_block_preserved(self):
        boot = {
            'api': {'base_url': 'http://x', 'timeout_seconds': 5, 'api_key': ''},
            'farm': {'farm_id': 1},
            'simulation': {'enabled': True, 'neopixel': {'count': 8}},
        }
        cfg = client.resolve_config(boot, None)
        self.assertTrue(cfg['simulation']['enabled'])
        self.assertEqual(cfg['simulation']['neopixel']['count'], 8)


if __name__ == '__main__':
    unittest.main()
