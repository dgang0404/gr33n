#!/usr/bin/env python3
"""Tests for Phase 125 WS3 synthetic sensor loopback."""

import os
import sys
import threading
import unittest

sys.path.insert(0, os.path.dirname(__file__))

import gr33n_client as client
from synthetic_sensors import SyntheticSensorLoop, synthetic_value


class TestSyntheticValue(unittest.TestCase):
    def test_hold(self):
        v = synthetic_value({'mode': 'hold', 'value': 42.0}, 0.0)
        self.assertEqual(v, 42.0)

    def test_sine_bounds(self):
        entry = {'mode': 'sine', 'center': 50, 'amplitude': 10, 'period_seconds': 100}
        vals = [synthetic_value(entry, t) for t in (0, 25, 50, 75)]
        self.assertTrue(all(40 <= v <= 60 for v in vals))

    def test_demo_moisture_drops_low(self):
        entry = {'mode': 'demo_moisture'}
        # t=100s → phase ~0.55 → transitioning/dropped
        v = synthetic_value(entry, 100.0)
        self.assertLess(v, 30.0)

    def test_demo_moisture_in_band_early(self):
        v = synthetic_value(entry := {'mode': 'demo_moisture'}, 30.0)
        self.assertGreaterEqual(v, 50.0)


class TestSyntheticSensorLoop(unittest.TestCase):
    def test_posts_and_updates_cache(self):
        cache = client.ReadingCache()
        posted = []
        stop = threading.Event()

        def post(sid, value, ts):
            posted.append((sid, value))
            return True

        loop = SyntheticSensorLoop(
            entries=[{'sensor_id': 7, 'mode': 'hold', 'value': 33.0, 'interval_seconds': 0}],
            reading_cache=cache,
            post_reading=post,
            queue_push=lambda *a: None,
            is_reachable=lambda: True,
            stop_event=stop,
        )
        loop.start()
        stop.wait(0.3)
        stop.set()
        loop.stop()

        self.assertTrue(posted)
        self.assertEqual(posted[0][0], 7)
        self.assertEqual(cache.get(7, 60), 33.0)


if __name__ == '__main__':
    unittest.main()
