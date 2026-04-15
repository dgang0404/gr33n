#!/usr/bin/env python3
"""
Unit tests for gr33n Pi client — no hardware required.
Tests sensor reading, offline queue, API client, and actuator logic.
Run: python3 -m pytest test_gr33n_client.py -v
"""

import json
import sqlite3
import tempfile
import threading
import time
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer
from unittest.mock import MagicMock, patch

# ── Import the client module ──────────────────────────────────────────────────
import sys, os
sys.path.insert(0, os.path.dirname(__file__))
import gr33n_client as client


# ─────────────────────────────────────────────────────────────────────────────
# FAKE API SERVER — runs in a thread, captures all requests
# ─────────────────────────────────────────────────────────────────────────────
received_requests = []

class FakeAPIHandler(BaseHTTPRequestHandler):
    def log_message(self, *args): pass  # silence request logs

    def do_GET(self):
        received_requests.append({"method": "GET", "path": self.path})
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()

        if self.path == "/health":
            self.wfile.write(b'{"status":"ok","service":"gr33n-api"}')
        elif "/devices" in self.path:
            # Return device 1 with a pending_command to trigger actuator
            self.wfile.write(json.dumps([
                {"id": 1, "config": {"pending_command": "actuator_on", "pending_schedule_id": 1}},
                {"id": 2, "config": {}},
                {"id": 3, "config": {"pending_command": "actuator_off"}},
            ]).encode())
        else:
            self.wfile.write(b'[]')

    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = json.loads(self.rfile.read(length)) if length else {}
        received_requests.append({"method": "POST", "path": self.path, "body": body})
        self.send_response(201)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"id":1}')

    def do_PATCH(self):
        length = int(self.headers.get("Content-Length", 0))
        body = json.loads(self.rfile.read(length)) if length else {}
        received_requests.append({"method": "PATCH", "path": self.path, "body": body})
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"id":1}')

    def do_DELETE(self):
        received_requests.append({"method": "DELETE", "path": self.path})
        self.send_response(204)
        self.end_headers()


def start_fake_server(port=18080):
    server = HTTPServer(("127.0.0.1", port), FakeAPIHandler)
    t = threading.Thread(target=server.serve_forever, daemon=True)
    t.start()
    return server


# ─────────────────────────────────────────────────────────────────────────────
# TESTS
# ─────────────────────────────────────────────────────────────────────────────
class TestOfflineQueue(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mktemp(suffix=".db")
        self.q = client.OfflineQueue(self.tmp)

    def tearDown(self):
        try: os.unlink(self.tmp)
        except: pass

    def test_push_and_pop(self):
        self.q.push(1, 22.5, "2026-03-03T10:00:00+00:00")
        self.q.push(2, 58.0, "2026-03-03T10:00:01+00:00")
        batch = self.q.pop_batch(10)
        self.assertEqual(len(batch), 2)
        self.assertEqual(batch[0]["sensor_id"], 1)
        self.assertAlmostEqual(batch[0]["value_raw"], 22.5)

    def test_ack_removes_rows(self):
        self.q.push(1, 22.5, "2026-03-03T10:00:00+00:00")
        batch = self.q.pop_batch(10)
        self.q.ack([item["_qid"] for item in batch])
        self.assertEqual(len(self.q.pop_batch(10)), 0)

    def test_partial_ack(self):
        self.q.push(1, 22.5, "2026-03-03T10:00:00+00:00")
        self.q.push(2, 58.0, "2026-03-03T10:00:01+00:00")
        batch = self.q.pop_batch(10)
        # Only ack the first one
        self.q.ack([batch[0]["_qid"]])
        remaining = self.q.pop_batch(10)
        self.assertEqual(len(remaining), 1)
        self.assertEqual(remaining[0]["sensor_id"], 2)

    def test_empty_ack_is_safe(self):
        self.q.ack([])  # should not raise

    def test_thread_safety(self):
        errors = []
        def push_many():
            try:
                for i in range(50):
                    self.q.push(i, float(i), "2026-03-03T10:00:00+00:00")
            except Exception as e:
                errors.append(e)
        threads = [threading.Thread(target=push_many) for _ in range(4)]
        for t in threads: t.start()
        for t in threads: t.join()
        self.assertEqual(errors, [])
        self.assertEqual(len(self.q.pop_batch(500)), 200)


class TestApiClient(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        received_requests.clear()
        cls.server = start_fake_server(18080)
        cls.api = client.Gr33nApiClient(
            base_url="http://127.0.0.1:18080",
            farm_id=1,
            api_key="",
            timeout=3,
        )

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()

    def test_health_check(self):
        self.assertTrue(self.api.is_reachable())

    def test_post_reading_success(self):
        ok = self.api.post_reading(1, 22.5, "2026-03-03T10:00:00+00:00")
        self.assertTrue(ok)
        posts = [r for r in received_requests if r["method"] == "POST" and "readings" in r["path"]]
        self.assertGreater(len(posts), 0)
        body = posts[-1]["body"]
        self.assertEqual(body["sensor_id"], 1)
        self.assertAlmostEqual(body["value_raw"], 22.5)
        self.assertTrue(body["is_valid"])

    def test_get_devices_returns_list(self):
        devices = self.api.get_devices()
        self.assertIsInstance(devices, list)
        self.assertEqual(len(devices), 3)

    def test_patch_device_status(self):
        ok = self.api.patch_device_status(1, "online")
        self.assertTrue(ok)
        patches = [r for r in received_requests if r["method"] == "PATCH"]
        self.assertGreater(len(patches), 0)
        self.assertEqual(patches[-1]["body"]["status"], "online")

    def test_post_actuator_event(self):
        ok = self.api.post_actuator_event(1, "actuator_on", "schedule_trigger", 1)
        self.assertTrue(ok)

    def test_clear_pending_command(self):
        self.assertTrue(self.api.clear_pending_command(device_id=1))

    def test_unreachable_api_returns_false(self):
        bad_api = client.Gr33nApiClient("http://127.0.0.1:19999", 1, timeout=1)
        self.assertFalse(bad_api.is_reachable())
        self.assertFalse(bad_api.post_reading(1, 22.5))


class TestSensorReader(unittest.TestCase):
    """Test mock/stub sensor reads — no hardware needed."""

    def _reader(self, sensor_type, source="dht22"):
        return client.SensorReader({
            "sensor_id": 1,
            "sensor_type": sensor_type,
            "source": source,
            "pin": 4,
            "channel": 0,
        })

    def test_temperature_mock(self):
        r = self._reader("temperature")
        val = r.read()
        self.assertIsNotNone(val)
        self.assertGreater(val, 0)
        self.assertLess(val, 50)

    def test_humidity_mock(self):
        r = self._reader("humidity")
        val = r.read()
        self.assertIsNotNone(val)
        self.assertGreaterEqual(val, 0)
        self.assertLessEqual(val, 100)

    def test_co2_mock(self):
        r = self._reader("co2", "mhz19")
        val = r.read()
        self.assertIsNotNone(val)
        self.assertGreater(val, 300)   # above outdoor ambient
        self.assertLess(val, 2000)     # below dangerous level

    def test_ec_mock(self):
        r = self._reader("ec", "ads1115")
        val = r.read()
        self.assertIsNotNone(val)
        self.assertGreaterEqual(val, 0)
        self.assertLessEqual(val, 5)   # mS/cm range

    def test_ph_mock(self):
        r = self._reader("ph", "ads1115")
        val = r.read()
        self.assertIsNotNone(val)
        self.assertGreaterEqual(val, 4)
        self.assertLessEqual(val, 9)

    def test_par_mock(self):
        r = self._reader("par", "bh1750")
        val = r.read()
        self.assertIsNotNone(val)
        self.assertGreaterEqual(val, 0)
        self.assertLessEqual(val, 2000)  # umol/m2/s

    def test_soil_moisture_mock(self):
        r = self._reader("soil_moisture", "ads1115")
        val = r.read()
        self.assertIsNotNone(val)
        self.assertGreaterEqual(val, 0)
        self.assertLessEqual(val, 100)


class TestActuatorController(unittest.TestCase):
    """Test actuator state machine — GPIO is stubbed automatically."""

    def setUp(self):
        self.actuator = client.ActuatorController({
            "actuator_id": 1,
            "device_type": "light",
            "gpio_pin": 17,
        })

    def test_initial_state_off(self):
        self.assertEqual(self.actuator.state, "off")

    def test_turn_on(self):
        self.actuator.turn_on()
        self.assertEqual(self.actuator.state, "on")

    def test_turn_off(self):
        self.actuator.turn_on()
        self.actuator.turn_off()
        self.assertEqual(self.actuator.state, "off")

    def test_execute_on_variants(self):
        for cmd in ["on", "actuator_on", "turn_on", "open", "start"]:
            self.actuator.turn_off()
            self.actuator.execute(cmd)
            self.assertEqual(self.actuator.state, "on", f"Failed for command: {cmd}")

    def test_execute_off_variants(self):
        for cmd in ["off", "actuator_off", "turn_off", "close", "stop"]:
            self.actuator.turn_on()
            self.actuator.execute(cmd)
            self.assertEqual(self.actuator.state, "off", f"Failed for command: {cmd}")

    def test_unknown_command_does_not_crash(self):
        self.actuator.execute("explode")  # should log warning, not raise
        self.assertEqual(self.actuator.state, "off")

    def test_device_id_defaults_to_actuator_id(self):
        a = client.ActuatorController({"actuator_id": 5, "device_type": "pump", "gpio_pin": 18})
        self.assertEqual(a.device_id, 5)

    def test_device_id_from_config(self):
        a = client.ActuatorController({"actuator_id": 5, "device_id": 42, "device_type": "pump", "gpio_pin": 18})
        self.assertEqual(a.device_id, 42)


class TestScheduleLoop(unittest.TestCase):
    """Test that pending_command on a device triggers the right actuator."""

    def test_pending_command_fires_actuator(self):
        server = start_fake_server(18081)
        try:
            api = client.Gr33nApiClient("http://127.0.0.1:18081", 1, timeout=3)
            actuator1 = client.ActuatorController({"actuator_id": 1, "device_type": "light", "gpio_pin": 17})
            actuator3 = client.ActuatorController({"actuator_id": 3, "device_type": "fan",   "gpio_pin": 22})
            devices = api.get_devices()
            actuators = {1: actuator1, 3: actuator3}
            for device in devices:
                did = device.get("id")
                cmd = (device.get("config") or {}).get("pending_command")
                if cmd and did in actuators:
                    actuators[did].execute(cmd)
            self.assertEqual(actuator1.state, "on")
            self.assertEqual(actuator3.state, "off")
        finally:
            server.shutdown()


class TestScheduleLoopDictCommand(unittest.TestCase):
    """Test _schedule_loop logic with dict-shaped pending_command."""

    def test_dict_pending_command_extracts_command_and_schedule_id(self):
        class DictCommandHandler(BaseHTTPRequestHandler):
            def log_message(self, *args): pass
            def do_GET(self):
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                if self.path == "/health":
                    self.wfile.write(b'{"status":"ok","service":"gr33n-api"}')
                elif "/devices" in self.path:
                    self.wfile.write(json.dumps([
                        {"id": 1, "config": {"pending_command": {"command": "on", "schedule_id": 1}}},
                        {"id": 2, "config": {}},
                        {"id": 3, "config": {"pending_command": {"command": "off", "schedule_id": 5}}},
                    ]).encode())
                else:
                    self.wfile.write(b'[]')
            def do_POST(self):
                length = int(self.headers.get("Content-Length", 0))
                self.rfile.read(length)
                self.send_response(201)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"id":1}')
            def do_PATCH(self):
                length = int(self.headers.get("Content-Length", 0))
                self.rfile.read(length)
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"id":1}')
            def do_DELETE(self):
                self.send_response(204)
                self.end_headers()

        server = HTTPServer(("127.0.0.1", 18083), DictCommandHandler)
        t = threading.Thread(target=server.serve_forever, daemon=True)
        t.start()
        try:
            api = client.Gr33nApiClient("http://127.0.0.1:18083", 1, timeout=3)
            actuator1 = client.ActuatorController({"actuator_id": 1, "device_id": 1, "device_type": "light", "gpio_pin": 17})
            actuator3 = client.ActuatorController({"actuator_id": 3, "device_id": 3, "device_type": "fan",   "gpio_pin": 22})
            actuator_by_device = {a.device_id: a for a in [actuator1, actuator3]}

            captured_schedule_ids = []
            for device in api.get_devices():
                did = device.get("id")
                config = device.get("config") or {}
                pending = config.get("pending_command")
                if not pending:
                    continue
                if isinstance(pending, dict):
                    cmd = pending.get("command", "")
                    sched_id = pending.get("schedule_id")
                else:
                    cmd = str(pending)
                    sched_id = config.get("pending_schedule_id")
                if not cmd:
                    continue
                actuator = actuator_by_device.get(did)
                if actuator:
                    actuator.execute(cmd)
                    captured_schedule_ids.append(sched_id)
                    api.post_actuator_event(actuator.actuator_id, cmd, "schedule_trigger", sched_id)
                    api.clear_pending_command(did)

            self.assertEqual(actuator1.state, "on")
            self.assertEqual(actuator3.state, "off")
            self.assertIn(1, captured_schedule_ids)
            self.assertIn(5, captured_schedule_ids)
        finally:
            server.shutdown()


class TestHeartbeat(unittest.TestCase):
    """Test that heartbeat logic patches status for every configured device."""

    def test_heartbeat_patches_all_devices(self):
        server = start_fake_server(18084)
        try:
            api = client.Gr33nApiClient("http://127.0.0.1:18084", 1, timeout=3)
            a1 = client.ActuatorController({"actuator_id": 1, "device_id": 10, "device_type": "light", "gpio_pin": 17})
            a2 = client.ActuatorController({"actuator_id": 2, "device_id": 20, "device_type": "fan",   "gpio_pin": 22})
            device_ids = {a.device_id for a in [a1, a2]}

            received_requests.clear()
            for did in device_ids:
                api.patch_device_status(did, "online")

            patches = [r for r in received_requests if r["method"] == "PATCH"]
            patched_paths = {r["path"] for r in patches}
            self.assertIn("/devices/10/status", patched_paths)
            self.assertIn("/devices/20/status", patched_paths)
            self.assertEqual(len(patches), 2)
            for p in patches:
                self.assertEqual(p["body"]["status"], "online")
        finally:
            server.shutdown()


class TestOfflineQueueIntegration(unittest.TestCase):
    """Test that readings queue when API is down and flush when it comes back."""

    def test_queue_and_flush(self):
        server = start_fake_server(18082)
        try:
            tmp = tempfile.mktemp(suffix=".db")
            q = client.OfflineQueue(tmp)
            bad_api  = client.Gr33nApiClient("http://127.0.0.1:19999", 1, timeout=1)
            good_api = client.Gr33nApiClient("http://127.0.0.1:18082", 1, timeout=3)

            for i in range(3):
                if not bad_api.is_reachable():
                    q.push(i+1, float(i+1), f"2026-03-03T10:0{i}:00+00:00")

            self.assertEqual(len(q.pop_batch(10)), 3)

            batch = q.pop_batch(10)
            acked = []
            for item in batch:
                if good_api.post_reading(item["sensor_id"], item["value_raw"], item["reading_time"]):
                    acked.append(item["_qid"])
            q.ack(acked)

            self.assertEqual(len(acked), 3)
            self.assertEqual(len(q.pop_batch(10)), 0)
            try: os.unlink(tmp)
            except: pass
        finally:
            server.shutdown()


if __name__ == "__main__":
    unittest.main(verbosity=2)


# ─────────────────────────────────────────────────────────────────────────────
# EDGE CASE TESTS — invalid readings, correct sensor routing, server errors
# Added 2026-03-05. Matches gr33n_client.py actual class signatures.
# ─────────────────────────────────────────────────────────────────────────────

class TestEdgeCases(unittest.TestCase):

    # ── Test 1: SensorReader.read() returns None for bad hardware ─────────────
    def test_invalid_sensor_value_raises_exception(self):
        """SensorReader.read() must raise when hardware faults — sensor loop catches it."""
        cfg = {
            "sensor_id": 1,
            "sensor_type": "temperature",
            "hardware_identifier": "GPIO4",
            "interval_seconds": 60,
        }
        reader = client.SensorReader(cfg)
        with patch.object(client.SensorReader, '_mock', side_effect=Exception("hardware fault")):
            with patch.object(client.SensorReader, '_init_hardware', return_value=None):
                with self.assertRaises(Exception) as ctx:
                    reader.read()
        self.assertIn(
            "hardware fault", str(ctx.exception),
            "Exception message should propagate so the sensor loop can log and queue offline"
        )

    # ── Test 2: POST hits the correct sensor ID URL ───────────────────────────
    def test_reading_posts_to_correct_sensor_id(self):
        """Gr33nApiClient.post_reading(sensor_id=3) must POST to /sensors/3/readings."""
        posted_paths = []

        class CapturingHandler(BaseHTTPRequestHandler):
            def log_message(self, *args): pass
            def do_POST(self):
                length = int(self.headers.get("Content-Length", 0))
                self.rfile.read(length)
                posted_paths.append(self.path)
                self.send_response(201)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"id":1}')

        server = HTTPServer(("127.0.0.1", 0), CapturingHandler)
        port = server.server_address[1]
        thread = threading.Thread(target=server.handle_request)
        thread.start()

        api = client.Gr33nApiClient(
            base_url=f"http://127.0.0.1:{port}",
            farm_id=1
        )
        api.post_reading(sensor_id=3, value_raw=21.0)
        thread.join(timeout=2)
        server.server_close()

        self.assertTrue(
            any("/sensors/3/readings" in p for p in posted_paths),
            f"Expected POST to /sensors/3/readings, got: {posted_paths}"
        )

    # ── Test 3: API 500 → post_reading returns False ──────────────────────────
    def test_api_500_returns_false(self):
        """Gr33nApiClient.post_reading must return False on a 500 so the
        caller (Gr33nPiClient sensor loop) can queue the reading offline."""

        class ServerErrorHandler(BaseHTTPRequestHandler):
            def log_message(self, *args): pass
            def do_POST(self):
                length = int(self.headers.get("Content-Length", 0))
                self.rfile.read(length)
                self.send_response(500)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"error":"internal server error"}')

        server = HTTPServer(("127.0.0.1", 0), ServerErrorHandler)
        port = server.server_address[1]
        thread = threading.Thread(target=server.handle_request)
        thread.daemon = True
        thread.start()

        api = client.Gr33nApiClient(
            base_url=f"http://127.0.0.1:{port}",
            farm_id=1
        )
        result = api.post_reading(sensor_id=1, value_raw=22.5)
        thread.join(timeout=2)
        server.server_close()

        self.assertFalse(
            result,
            "post_reading must return False on HTTP 500 so the sensor loop queues offline"
        )


if __name__ == "__main__":
    unittest.main()
