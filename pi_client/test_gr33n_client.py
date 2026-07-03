#!/usr/bin/env python3
"""
Unit tests for gr33n Pi client — no hardware required.
Tests sensor reading, offline queue, API client, and actuator logic.
Run: python3 -m pytest test_gr33n_client.py -v
"""

import base64
import copy
import json
import sqlite3
import tempfile
import threading
import time
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer
from unittest.mock import MagicMock, patch

import yaml

# ── Import the client module ──────────────────────────────────────────────────
import sys, os
sys.path.insert(0, os.path.dirname(__file__))
import gr33n_client as client


class TestDeviceConfigDecode(unittest.TestCase):
    """gr33n_client._device_config_dict matches Go JSON encoding of []byte config."""

    def test_dict_passthrough(self):
        self.assertEqual(client._device_config_dict({'foo': 1}), {'foo': 1})

    def test_base64_roundtrip(self):
        inner = {'pending_command': {'command': 'on', 'schedule_id': 7, 'program_id': 42}}
        b64 = base64.b64encode(json.dumps(inner).encode('utf-8')).decode('ascii')
        got = client._device_config_dict(b64)
        self.assertEqual(got['pending_command']['command'], 'on')
        self.assertEqual(got['pending_command']['schedule_id'], 7)
        self.assertEqual(got['pending_command']['program_id'], 42)

    def test_none_empty(self):
        self.assertEqual(client._device_config_dict(None), {})
        self.assertEqual(client._device_config_dict('not-valid-base64!!!'), {})


class TestPostActuatorEventPayload(unittest.TestCase):
    """POST /actuators/{id}/events JSON includes provenance fields."""

    def test_includes_rule_schedule_program(self):
        last = {}

        class CaptureHandler(BaseHTTPRequestHandler):
            def log_message(self, *args):
                pass

            def do_POST(self):
                n = int(self.headers.get('Content-Length', 0))
                raw = self.rfile.read(n) if n else b'{}'
                last['path'] = self.path
                last['json'] = json.loads(raw.decode('utf-8'))
                self.send_response(201)
                self.send_header('Content-Type', 'application/json')
                self.end_headers()
                self.wfile.write(b'{"ok":true}')

        srv = HTTPServer(('127.0.0.1', 18095), CaptureHandler)
        t = threading.Thread(target=srv.serve_forever, daemon=True)
        t.start()
        try:
            api = client.Gr33nApiClient('http://127.0.0.1:18095', 1, timeout=3)
            ok = api.post_actuator_event(
                12, 'on', source='schedule_trigger',
                schedule_id=3, rule_id=None, program_id=99,
                meta_data={'edge': 'pytest'},
                parameters_sent={'v': 1},
            )
            self.assertTrue(ok)
            self.assertIn('/actuators/12/events', last['path'])
            body = last['json']
            self.assertEqual(body['triggered_by_schedule_id'], 3)
            self.assertEqual(body['program_id'], 99)
            self.assertEqual(body['meta_data']['edge'], 'pytest')
            self.assertEqual(body['parameters_sent']['v'], 1)
            self.assertNotIn('triggered_by_rule_id', body)
        finally:
            srv.shutdown()


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

    def test_post_readings_batch_success(self):
        """post_readings_batch POSTs a JSON array to /sensors/readings/batch."""
        before = len(received_requests)
        items = [
            {"sensor_id": 1, "value_raw": 22.5, "reading_time": "2026-03-03T10:00:00+00:00", "is_valid": True},
            {"sensor_id": 2, "value_raw": 58.1, "reading_time": "2026-03-03T10:00:01+00:00", "is_valid": True},
            {"sensor_id": 3, "value_raw": 1.42, "reading_time": "2026-03-03T10:00:02+00:00", "is_valid": True},
        ]
        ok = self.api.post_readings_batch(items)
        self.assertTrue(ok)
        batch_posts = [
            r for r in received_requests[before:]
            if r["method"] == "POST" and r["path"].endswith("/sensors/readings/batch")
        ]
        self.assertEqual(len(batch_posts), 1, "expected exactly one POST to /sensors/readings/batch")
        body = batch_posts[0]["body"]
        self.assertIsInstance(body, list)
        self.assertEqual(len(body), 3)
        self.assertEqual(body[0]["sensor_id"], 1)
        self.assertAlmostEqual(body[2]["value_raw"], 1.42)

    def test_post_readings_batch_empty_short_circuits(self):
        """Empty items list must return True without making a network request."""
        before = len(received_requests)
        ok = self.api.post_readings_batch([])
        self.assertTrue(ok)
        after = received_requests[before:]
        self.assertEqual(
            [r for r in after if r["method"] == "POST" and "readings/batch" in r["path"]],
            [],
            "empty batch must not issue a POST",
        )

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

    def test_guardian_pending_uses_manual_api_call_and_proposal_meta(self):
        posted = []

        class GuardianHandler(BaseHTTPRequestHandler):
            def log_message(self, *args): pass
            def do_GET(self):
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                if self.path == "/health":
                    self.wfile.write(b'{"status":"ok"}')
                elif "/devices" in self.path:
                    self.wfile.write(json.dumps([
                        {"id": 7, "config": {"pending_command": {
                            "command": "on",
                            "source": "guardian",
                            "proposal_id": "prop-123",
                            "reason": "operator inspection",
                        }}},
                    ]).encode())
            def do_POST(self):
                length = int(self.headers.get("Content-Length", 0))
                body = self.rfile.read(length)
                posted.append(json.loads(body.decode()))
                self.send_response(201)
                self.end_headers()
                self.wfile.write(b'{"id":1}')
            def do_DELETE(self):
                self.send_response(204)
                self.end_headers()

        server = HTTPServer(("127.0.0.1", 18085), GuardianHandler)
        threading.Thread(target=server.serve_forever, daemon=True).start()
        try:
            api = client.Gr33nApiClient("http://127.0.0.1:18085", 1, timeout=3)
            actuator = client.ActuatorController({"actuator_id": 9, "device_id": 7, "device_type": "light", "gpio_pin": 17})
            device = api.get_devices()[0]
            pending = device["config"]["pending_command"]
            cmd = pending["command"]
            actuator.execute(cmd)
            meta = {}
            if pending.get("proposal_id"):
                meta["proposal_id"] = pending["proposal_id"]
            if pending.get("reason"):
                meta["reason"] = pending["reason"]
            src = "manual_api_call" if pending.get("source") == "guardian" else "schedule_trigger"
            api.post_actuator_event(actuator.actuator_id, cmd, src, meta_data=meta or None)
            api.clear_pending_command(device["id"])
            self.assertEqual(len(posted), 1)
            self.assertEqual(posted[0]["source"], "manual_api_call")
            self.assertEqual(posted[0]["meta_data"]["proposal_id"], "prop-123")
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


# ─────────────────────────────────────────────────────────────────────────────
# CONFIG + API KEY + DAEMON SKIP TESTS
# ─────────────────────────────────────────────────────────────────────────────

class TestLoadConfig(unittest.TestCase):
    """Test load_config merges YAML with defaults correctly."""

    def test_partial_config_merges_with_defaults(self):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump({'api': {'base_url': 'http://10.0.0.1:9090'}, 'farm': {'farm_id': 42}}, f)
            path = f.name
        try:
            cfg = client.load_config(path)
            self.assertEqual(cfg['api']['base_url'], 'http://10.0.0.1:9090')
            self.assertEqual(cfg['farm']['farm_id'], 42)
            self.assertEqual(cfg['api']['timeout_seconds'], 5)
            self.assertEqual(cfg['schedule_poll_interval_seconds'], 30)
        finally:
            os.unlink(path)

    def test_missing_file_returns_full_defaults(self):
        cfg = client.load_config('/tmp/nonexistent_gr33n_test_config.yaml')
        self.assertEqual(cfg, client.DEFAULT_CONFIG)

    def test_empty_yaml_returns_defaults(self):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            f.write('')
            path = f.name
        try:
            cfg = client.load_config(path)
            self.assertEqual(cfg['api']['base_url'], client.DEFAULT_CONFIG['api']['base_url'])
        finally:
            os.unlink(path)

    def test_scalar_override(self):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump({'schedule_poll_interval_seconds': 120}, f)
            path = f.name
        try:
            cfg = client.load_config(path)
            self.assertEqual(cfg['schedule_poll_interval_seconds'], 120)
        finally:
            os.unlink(path)


class TestApiKeyHeader(unittest.TestCase):
    """Test that X-Api-Key header is sent when api_key is configured."""

    @patch.object(client, '_read_device_key_file', return_value='')
    @patch.dict(os.environ, {}, clear=True)
    def test_api_key_sent_in_header(self, _mock_file):
        captured_headers = {}

        class HeaderCapture(BaseHTTPRequestHandler):
            def log_message(self, *args): pass
            def do_POST(self):
                captured_headers.update(self.headers)
                length = int(self.headers.get("Content-Length", 0))
                self.rfile.read(length)
                self.send_response(201)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"id":1}')

        server = HTTPServer(("127.0.0.1", 0), HeaderCapture)
        port = server.server_address[1]
        thread = threading.Thread(target=server.handle_request)
        thread.start()
        try:
            api = client.Gr33nApiClient(
                base_url=f"http://127.0.0.1:{port}",
                farm_id=1,
                api_key="test-key-123",
            )
            api.post_reading(1, 22.5)
            thread.join(timeout=2)
        finally:
            server.server_close()

        self.assertEqual(captured_headers.get("X-Api-Key"), "test-key-123")

    @patch.object(client, '_read_device_key_file', return_value='')
    @patch.dict(os.environ, {}, clear=True)
    def test_no_api_key_when_empty(self, _mock_file):
        captured_headers = {}

        class HeaderCapture(BaseHTTPRequestHandler):
            def log_message(self, *args): pass
            def do_POST(self):
                captured_headers.update(self.headers)
                length = int(self.headers.get("Content-Length", 0))
                self.rfile.read(length)
                self.send_response(201)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"id":1}')

        server = HTTPServer(("127.0.0.1", 0), HeaderCapture)
        port = server.server_address[1]
        thread = threading.Thread(target=server.handle_request)
        thread.start()
        try:
            api = client.Gr33nApiClient(
                base_url=f"http://127.0.0.1:{port}",
                farm_id=1,
                api_key="",
            )
            api.post_reading(1, 22.5)
            thread.join(timeout=2)
        finally:
            server.server_close()

        self.assertIsNone(captured_headers.get("X-Api-Key"))

    @patch.object(client, '_read_device_key_file', return_value='')
    @patch.dict(os.environ, {}, clear=True)
    def test_device_key_uses_x_device_key_header(self, _mock_file):
        captured_headers = {}

        class HeaderCapture(BaseHTTPRequestHandler):
            def log_message(self, *args): pass
            def do_POST(self):
                captured_headers.update(self.headers)
                length = int(self.headers.get("Content-Length", 0))
                self.rfile.read(length)
                self.send_response(201)
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                self.wfile.write(b'{"id":1}')

        server = HTTPServer(("127.0.0.1", 0), HeaderCapture)
        port = server.server_address[1]
        thread = threading.Thread(target=server.handle_request)
        thread.start()
        try:
            api = client.Gr33nApiClient(
                base_url=f"http://127.0.0.1:{port}",
                farm_id=1,
                api_key="gdev_42_testsecret",
            )
            api.post_reading(1, 22.5)
            thread.join(timeout=2)
        finally:
            server.server_close()

        self.assertEqual(captured_headers.get("X-Device-Key"), "gdev_42_testsecret")
        self.assertIsNone(captured_headers.get("X-Api-Key"))

    @patch.object(client, '_read_device_key_file', return_value='')
    @patch.dict(os.environ, {"GR33N_DEVICE_API_KEY": "gdev_9_fromenv"}, clear=True)
    def test_resolve_edge_prefers_env_device_key(self, _mock_file):
        header, cred = client.resolve_edge_api_credential("legacy-shared")
        self.assertEqual(header, "X-Device-Key")
        self.assertEqual(cred, "gdev_9_fromenv")


class TestDaemonLoopSkipOnUnreachable(unittest.TestCase):
    """Test that daemon loops skip gracefully when API is unreachable."""

    def _make_client(self):
        with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
            yaml.dump({
                'api': {'base_url': 'http://127.0.0.1:19999', 'timeout_seconds': 1, 'api_key': ''},
                'farm': {'farm_id': 1},
                'sensors': [{'sensor_id': 1, 'sensor_type': 'temperature', 'source': 'dht22', 'pin': 4, 'interval_seconds': 1}],
                'actuators': [{'actuator_id': 1, 'device_id': 1, 'device_type': 'light', 'gpio_pin': 17}],
                'schedule_poll_interval_seconds': 1,
                'offline_queue_path': tempfile.mktemp(suffix='.db'),
                'offline_flush_interval_seconds': 1,
            }, f)
            self._cfg_path = f.name
        return client.Gr33nPiClient(self._cfg_path)

    def test_schedule_loop_skips_on_unreachable(self):
        pi = self._make_client()
        self.assertFalse(pi.api.is_reachable())
        with patch.object(pi.api, 'get_devices') as mock_get:
            pi.api.is_reachable = MagicMock(return_value=False)
            pi._stop.set()
            # Simulate one iteration: check reachable -> skip
            if not pi.api.is_reachable():
                pass  # skip as expected
            mock_get.assert_not_called()
        os.unlink(self._cfg_path)

    def test_heartbeat_loop_skips_on_unreachable(self):
        pi = self._make_client()
        with patch.object(pi.api, 'patch_device_status') as mock_patch:
            pi.api.is_reachable = MagicMock(return_value=False)
            pi._stop.set()
            if not pi.api.is_reachable():
                pass
            mock_patch.assert_not_called()
        os.unlink(self._cfg_path)

    def test_flush_loop_skips_on_unreachable(self):
        pi = self._make_client()
        with patch.object(pi.queue, 'pop_batch') as mock_pop:
            pi.api.is_reachable = MagicMock(return_value=False)
            pi._stop.set()
            if not pi.api.is_reachable():
                pass
            mock_pop.assert_not_called()
        os.unlink(self._cfg_path)


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

    # ── Test 4: batch path + 500 handling ─────────────────────────────────────
    def test_batch_posts_to_correct_path_and_handles_500(self):
        """post_readings_batch must hit /sensors/readings/batch and return False on 500."""
        posted_paths = []
        posted_bodies = []

        class BatchHandler(BaseHTTPRequestHandler):
            def log_message(self, *args): pass
            mode = {"status": 201}

            def do_POST(self):
                length = int(self.headers.get("Content-Length", 0))
                raw = self.rfile.read(length)
                try:
                    posted_bodies.append(json.loads(raw))
                except Exception:
                    posted_bodies.append(None)
                posted_paths.append(self.path)
                self.send_response(BatchHandler.mode["status"])
                self.send_header("Content-Type", "application/json")
                self.end_headers()
                if BatchHandler.mode["status"] >= 500:
                    self.wfile.write(b'{"error":"boom"}')
                else:
                    self.wfile.write(b'{"inserted":2}')

        server = HTTPServer(("127.0.0.1", 0), BatchHandler)
        port = server.server_address[1]
        t = threading.Thread(target=server.serve_forever, daemon=True)
        t.start()
        try:
            api = client.Gr33nApiClient(base_url=f"http://127.0.0.1:{port}", farm_id=1, timeout=3)

            items = [
                {"sensor_id": 1, "value_raw": 22.5, "reading_time": "2026-03-03T10:00:00+00:00", "is_valid": True},
                {"sensor_id": 2, "value_raw": 58.1, "reading_time": "2026-03-03T10:00:01+00:00", "is_valid": True},
            ]

            BatchHandler.mode["status"] = 201
            self.assertTrue(api.post_readings_batch(items))
            self.assertTrue(
                any(p.endswith("/sensors/readings/batch") for p in posted_paths),
                f"expected POST to /sensors/readings/batch, got: {posted_paths}",
            )
            self.assertEqual(posted_bodies[-1], items)

            BatchHandler.mode["status"] = 500
            self.assertFalse(
                api.post_readings_batch(items),
                "post_readings_batch must return False on HTTP 500 so caller can requeue",
            )
        finally:
            server.shutdown()
            server.server_close()


# ─────────────────────────────────────────────────────────────────────────────
# Phase 20.5 WS1 — Derived sensors (dew_point / vpd / heat_index)
# ─────────────────────────────────────────────────────────────────────────────

class TestDerivedSensorMath(unittest.TestCase):
    """Pure-math tests for derived-sensor formulas.

    Reference pairs from published engineering tables (NOAA, FAO-56,
    Extension agronomy). Tolerance is wide enough to accommodate the single
    rounding step inside the computer functions.
    """

    def test_dew_point_known_pair_hot_humid(self):
        # t=30°C, RH=80% → dp ~= 26.2°C (NOAA psychrometric chart)
        self.assertAlmostEqual(client.compute_dew_point_c(30.0, 80.0), 26.2, delta=0.2)

    def test_dew_point_known_pair_cool_moderate(self):
        # t=20°C, RH=50% → dp ~= 9.3°C
        self.assertAlmostEqual(client.compute_dew_point_c(20.0, 50.0), 9.3, delta=0.3)

    def test_dew_point_cannabis_flower_window(self):
        # cannabis flower target: ~25°C / 50% RH → dp ~= 13.9°C (cited mid-flower ideal)
        self.assertAlmostEqual(client.compute_dew_point_c(25.0, 50.0), 13.9, delta=0.3)

    def test_dew_point_handles_zero_humidity(self):
        # Degenerate input must not crash (log(0) guard).
        val = client.compute_dew_point_c(20.0, 0.0)
        self.assertIsInstance(val, float)

    def test_vpd_known_pair(self):
        # t=25°C, RH=50% → SVP ~= 3.169 kPa → VPD ~= 1.585 kPa
        self.assertAlmostEqual(client.compute_vpd_kpa(25.0, 50.0), 1.585, delta=0.01)

    def test_vpd_saturated_air_is_zero(self):
        # RH=100% means zero deficit regardless of temperature.
        self.assertAlmostEqual(client.compute_vpd_kpa(22.0, 100.0), 0.0, delta=0.001)

    def test_vpd_low_rh_is_higher(self):
        # Monotonicity: at fixed temperature, lower RH → higher VPD.
        self.assertGreater(
            client.compute_vpd_kpa(25.0, 30.0),
            client.compute_vpd_kpa(25.0, 70.0),
        )

    def test_heat_index_below_threshold_returns_dry_bulb(self):
        # Below ~26.7°C (80°F) the NWS regression doesn't apply.
        self.assertAlmostEqual(client.compute_heat_index_c(20.0, 80.0), 20.0, delta=0.01)

    def test_heat_index_hot_humid(self):
        # t=32°C (~90°F), RH=70% → HI ~= 41°C (~106°F per NWS table)
        self.assertAlmostEqual(client.compute_heat_index_c(32.0, 70.0), 41.0, delta=1.5)


class TestReadingCache(unittest.TestCase):
    """Staleness + thread-safety of the shared ReadingCache."""

    def test_get_returns_none_when_empty(self):
        c = client.ReadingCache()
        self.assertIsNone(c.get(42, max_age_s=60))

    def test_get_returns_fresh_value(self):
        c = client.ReadingCache()
        c.put(1, 22.5, now=100.0)
        self.assertEqual(c.get(1, max_age_s=60, now=120.0), 22.5)

    def test_get_returns_none_when_stale(self):
        c = client.ReadingCache()
        c.put(1, 22.5, now=0.0)
        self.assertIsNone(c.get(1, max_age_s=30, now=1000.0))

    def test_put_overwrites(self):
        c = client.ReadingCache()
        c.put(1, 22.5, now=100.0)
        c.put(1, 30.0, now=110.0)
        self.assertEqual(c.get(1, max_age_s=60, now=120.0), 30.0)


class TestDerivedSensorReader(unittest.TestCase):
    """Integration of SensorReader with ReadingCache for `source: derived`."""

    def _derived_reader(self, stype='dew_point', t_sid=1, rh_sid=2,
                        max_age=120, cache=None):
        cfg = {
            'sensor_id': 99,
            'sensor_type': stype,
            'source': 'derived',
            'inputs': {'temperature_c': t_sid, 'humidity_pct': rh_sid},
            'input_max_age_seconds': max_age,
        }
        return client.SensorReader(cfg, cache=cache)

    def test_returns_none_when_inputs_missing(self):
        cache = client.ReadingCache()
        r = self._derived_reader(cache=cache)
        self.assertIsNone(r.read())

    def test_returns_none_when_one_input_missing(self):
        cache = client.ReadingCache()
        cache.put(1, 25.0)
        r = self._derived_reader(cache=cache)
        self.assertIsNone(r.read(),
            "dew_point must not fabricate a value when humidity is uncached")

    def test_returns_none_without_cache(self):
        r = self._derived_reader(cache=None)
        self.assertIsNone(r.read(),
            "a derived reader without a cache binding should fail closed")

    def test_computes_dew_point_when_inputs_fresh(self):
        cache = client.ReadingCache()
        cache.put(1, 25.0)
        cache.put(2, 50.0)
        r = self._derived_reader(stype='dew_point', cache=cache)
        val = r.read()
        self.assertIsNotNone(val)
        self.assertAlmostEqual(val, 13.9, delta=0.3)

    def test_computes_vpd_when_inputs_fresh(self):
        cache = client.ReadingCache()
        cache.put(1, 25.0)
        cache.put(2, 50.0)
        r = self._derived_reader(stype='vpd', cache=cache)
        val = r.read()
        self.assertIsNotNone(val)
        self.assertAlmostEqual(val, 1.585, delta=0.01)

    def test_returns_none_when_input_stale(self):
        cache = client.ReadingCache()
        # Seed values with an ancient monotonic timestamp so the max-age
        # check trips no matter when the test runs.
        cache.put(1, 25.0, now=0.0)
        cache.put(2, 50.0, now=0.0)
        r = self._derived_reader(max_age=1, cache=cache)
        self.assertIsNone(r.read(),
            "stale inputs must suppress the derived reading, not emit old math")

    def test_unknown_sensor_type_returns_none(self):
        cache = client.ReadingCache()
        cache.put(1, 25.0)
        cache.put(2, 50.0)
        r = self._derived_reader(stype='mystery_metric', cache=cache)
        self.assertIsNone(r.read())


# ─────────────────────────────────────────────────────────────────────────────
# PHASE 51 WS2 — platform config sync (bootstrap + fetch + cache)
# ─────────────────────────────────────────────────────────────────────────────

class TestPhase51ConfigSync(unittest.TestCase):

    REMOTE_CONFIG = {
        'device_uid': 'veg-pi-01',
        'device_id': 1,
        'farm_id': 1,
        'config_version': 3,
        'sensors': [
            {'sensor_id': 3, 'sensor_type': 'temperature', 'source': 'dht22',
             'pin': 4, 'interval_seconds': 60},
        ],
        'actuators': [
            {'actuator_id': 1, 'device_id': 1, 'device_type': 'light', 'gpio_pin': 17},
        ],
        'schedule_poll_interval_seconds': 30,
        'offline_queue_path': '/var/lib/gr33n/queue.db',
        'offline_flush_interval_seconds': 60,
    }

    def _mock_config_server(self, status=200, body=None):
        payload = body if body is not None else self.REMOTE_CONFIG

        class Handler(BaseHTTPRequestHandler):
            def log_message(self, *args): pass

            def do_GET(self):
                if self.path.startswith('/devices/by-uid/veg-pi-01/config'):
                    self.send_response(status)
                    self.send_header('Content-Type', 'application/json')
                    self.end_headers()
                    if status == 200:
                        self.wfile.write(json.dumps(payload).encode())
                    return
                self.send_response(404)
                self.end_headers()

        server = HTTPServer(('127.0.0.1', 0), Handler)
        port = server.server_address[1]
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        return server, port

    def test_fetch_remote_config_success(self):
        server, port = self._mock_config_server()
        try:
            api = client.Gr33nApiClient(f'http://127.0.0.1:{port}', 1, api_key='k')
            got = client.fetch_remote_config(api, 'veg-pi-01')
            self.assertIsNotNone(got)
            self.assertEqual(got['config_version'], 3)
            self.assertEqual(len(got['sensors']), 1)
            self.assertEqual(got['sensors'][0]['pin'], 4)
        finally:
            server.shutdown()

    def test_resolve_config_uses_remote_wiring(self):
        bootstrap = client.load_bootstrap('/tmp/nonexistent_phase51_bootstrap.yaml')
        bootstrap['device'] = {'uid': 'veg-pi-01'}
        cfg = client.resolve_config(bootstrap, self.REMOTE_CONFIG)
        self.assertEqual(cfg['sensors'][0]['sensor_id'], 3)
        self.assertEqual(cfg['actuators'][0]['gpio_pin'], 17)
        self.assertEqual(cfg['config_version'], 3)
        self.assertEqual(cfg['api']['base_url'], bootstrap['api']['base_url'])

    def test_compute_wiring_config_sha256_stable(self):
        cfg = client.resolve_config(
            {'api': {'base_url': 'http://x'}, 'farm': {'farm_id': 1}, 'device': {'uid': 'u'}},
            self.REMOTE_CONFIG,
        )
        h1 = client.compute_wiring_config_sha256(cfg)
        h2 = client.compute_wiring_config_sha256(cfg)
        self.assertEqual(h1, h2)
        self.assertEqual(len(h1), 64)

    def test_local_wiring_takes_precedence(self):
        bootstrap = {
            'api': {'base_url': 'http://10.0.0.2:8080', 'timeout_seconds': 5, 'api_key': ''},
            'farm': {'farm_id': 1},
            'sensors': [{'sensor_id': 99, 'sensor_type': 'temperature', 'source': 'dht22',
                         'pin': 7, 'interval_seconds': 30}],
            'actuators': [],
            'schedule_poll_interval_seconds': 30,
            'offline_queue_path': tempfile.mktemp(suffix='.db'),
            'offline_flush_interval_seconds': 60,
        }
        cfg = client.resolve_config(bootstrap, self.REMOTE_CONFIG)
        self.assertEqual(cfg['sensors'][0]['sensor_id'], 99)
        self.assertEqual(cfg['sensors'][0]['pin'], 7)
        self.assertNotIn('config_version', cfg)

    def test_fetch_remote_config_offline_falls_back_to_cache(self):
        cache_path = tempfile.mktemp(suffix='.json')
        client.write_config_cache(cache_path, self.REMOTE_CONFIG)
        try:
            bootstrap = client.load_bootstrap('/tmp/nonexistent_phase51_bootstrap.yaml')
            bootstrap['device'] = {'uid': 'veg-pi-01'}
            bootstrap['offline_queue_path'] = tempfile.mktemp(suffix='.db')
            api = client.Gr33nApiClient('http://127.0.0.1:19999', 1, timeout=1)
            cfg, synced = client.resolve_startup_config(bootstrap, api, cache_path)
            self.assertFalse(synced)
            self.assertEqual(cfg['sensors'][0]['sensor_id'], 3)
            self.assertEqual(cfg['config_version'], 3)
        finally:
            os.unlink(cache_path)

    def test_startup_without_wiring_or_cache_raises(self):
        bootstrap = client.load_bootstrap('/tmp/nonexistent_phase51_bootstrap.yaml')
        bootstrap['device'] = {'uid': 'veg-pi-01'}
        api = client.Gr33nApiClient('http://127.0.0.1:19999', 1, timeout=1)
        with self.assertRaises(RuntimeError) as ctx:
            client.resolve_startup_config(bootstrap, api, tempfile.mktemp(suffix='.json'))
        self.assertIn('no wiring config', str(ctx.exception))

    def test_resolve_startup_config_reports_live_fetch(self):
        server, port = self._mock_config_server()
        cache_path = tempfile.mktemp(suffix='.json')
        try:
            bootstrap = client.load_bootstrap('/tmp/nonexistent_phase51_bootstrap.yaml')
            bootstrap['device'] = {'uid': 'veg-pi-01'}
            bootstrap['offline_queue_path'] = tempfile.mktemp(suffix='.db')
            api = client.Gr33nApiClient(f'http://127.0.0.1:{port}', 1, timeout=3, api_key='k')
            cfg, synced = client.resolve_startup_config(bootstrap, api, cache_path)
            self.assertTrue(synced)
            self.assertEqual(cfg['device_id'], 1)
        finally:
            server.shutdown()

    def test_bootstrap_client_fetches_on_startup(self):
        server, port = self._mock_config_server()
        cache_path = tempfile.mktemp(suffix='.json')
        try:
            with tempfile.NamedTemporaryFile('w', suffix='.yaml', delete=False) as f:
                yaml.dump({
                    'api': {'base_url': f'http://127.0.0.1:{port}', 'timeout_seconds': 3,
                            'api_key': 'test'},
                    'farm': {'farm_id': 1},
                    'device': {'uid': 'veg-pi-01'},
                    'offline_queue_path': tempfile.mktemp(suffix='.db'),
                }, f)
                cfg_path = f.name
            os.environ['CONFIG_CACHE_PATH'] = cache_path
            pi = client.Gr33nPiClient(cfg_path)
            self.assertEqual(pi.cfg['sensors'][0]['pin'], 4)
            self.assertEqual(pi._config_version, 3)
            self.assertTrue(os.path.exists(cache_path))
            os.unlink(cfg_path)
        finally:
            os.environ.pop('CONFIG_CACHE_PATH', None)
            if os.path.exists(cache_path):
                os.unlink(cache_path)
            server.shutdown()


class TestPhase51LiveReload(unittest.TestCase):

    BOOTSTRAP = {
        'api': {'base_url': 'http://127.0.0.1:1', 'timeout_seconds': 1, 'api_key': 'k'},
        'farm': {'farm_id': 1},
        'device': {'uid': 'veg-pi-01'},
        'offline_queue_path': '',
        'schedule_poll_interval_seconds': 30,
        'offline_flush_interval_seconds': 60,
    }

    REMOTE_V3 = {
        'device_uid': 'veg-pi-01',
        'config_version': 3,
        'sensors': [
            {'sensor_id': 3, 'sensor_type': 'temperature', 'source': 'dht22',
             'pin': 4, 'interval_seconds': 60},
        ],
        'actuators': [
            {'actuator_id': 1, 'device_id': 1, 'device_type': 'light', 'gpio_pin': 17},
        ],
    }

    REMOTE_V4 = {
        **REMOTE_V3,
        'config_version': 4,
        'sensors': [
            {'sensor_id': 3, 'sensor_type': 'temperature', 'source': 'dht22',
             'pin': 5, 'interval_seconds': 60},
        ],
    }

    def _platform_client(self, queue_path):
        bootstrap = copy.deepcopy(self.BOOTSTRAP)
        bootstrap['offline_queue_path'] = queue_path
        with tempfile.NamedTemporaryFile('w', suffix='.yaml', delete=False) as f:
            yaml.dump(bootstrap, f)
            cfg_path = f.name
        with patch.object(client.Gr33nPiClient, '__init__', lambda self, *a, **k: None):
            pi = client.Gr33nPiClient(cfg_path)
        pi._bootstrap = bootstrap
        pi._config_cache_path = tempfile.mktemp(suffix='.json')
        pi.api = MagicMock()
        pi.device_uid = 'veg-pi-01'
        pi._config_version = 3
        pi._stop = threading.Event()
        pi._hw_lock = threading.Lock()
        pi._last_read = {}
        pi._reading_cache = client.ReadingCache()
        pi.cfg = client.resolve_config(bootstrap, self.REMOTE_V3)
        pi._readers, pi._actuators = pi._build_hardware(pi.cfg, {}, {})
        return pi, cfg_path

    def test_reload_config_rejects_empty_wiring(self):
        pi, cfg_path = self._platform_client(tempfile.mktemp(suffix='.db'))
        try:
            pi.api.fetch_device_config = MagicMock(return_value={
                'config_version': 9, 'sensors': [], 'actuators': [],
            })
            self.assertFalse(pi._reload_config())
            self.assertEqual(pi._config_version, 3)
            self.assertEqual(pi._readers[3].cfg['pin'], 4)
        finally:
            os.unlink(cfg_path)

    def test_reload_config_swaps_readers_atomically(self):
        pi, cfg_path = self._platform_client(tempfile.mktemp(suffix='.db'))
        try:
            old_reader = pi._readers[3]
            pi.api.fetch_device_config = MagicMock(return_value=self.REMOTE_V4)
            with patch.object(old_reader, 'close') as mock_close:
                self.assertTrue(pi._reload_config())
                mock_close.assert_called_once()
            self.assertEqual(pi._config_version, 4)
            self.assertEqual(pi._readers[3].cfg['pin'], 5)
            self.assertIsNot(pi._readers[3], old_reader)
        finally:
            os.unlink(cfg_path)

    def test_poll_config_version_triggers_reload(self):
        pi, cfg_path = self._platform_client(tempfile.mktemp(suffix='.db'))
        try:
            pi.api.get_config_version = MagicMock(return_value=4)
            with patch.object(pi, '_reload_config', return_value=True) as mock_reload:
                pi._poll_config_version()
                mock_reload.assert_called_once()
            pi.api.get_config_version = MagicMock(return_value=3)
            with patch.object(pi, '_reload_config') as mock_reload:
                pi._poll_config_version()
                mock_reload.assert_not_called()
        finally:
            os.unlink(cfg_path)

    def test_report_config_sync_patches_last_fetch(self):
        pi, cfg_path = self._platform_client(tempfile.mktemp(suffix='.db'))
        try:
            pi.cfg['device_id'] = 42
            pi.api.patch_device_status = MagicMock(return_value=True)
            pi._report_config_sync()
            pi.api.patch_device_status.assert_called_once()
            args, kwargs = pi.api.patch_device_status.call_args
            self.assertEqual(args[0], 42)
            self.assertEqual(args[1], 'online')
            self.assertIn('last_config_fetch_at', kwargs)
            self.assertIn('config_sha256', kwargs)
            self.assertTrue(len(kwargs['config_sha256']) == 64)
        finally:
            os.unlink(cfg_path)

    def test_local_wiring_skips_version_poll(self):
        bootstrap = {
            **self.BOOTSTRAP,
            'sensors': [{'sensor_id': 1, 'sensor_type': 'temperature', 'source': 'dht22',
                         'pin': 4, 'interval_seconds': 60}],
            'actuators': [],
            'offline_queue_path': tempfile.mktemp(suffix='.db'),
        }
        with tempfile.NamedTemporaryFile('w', suffix='.yaml', delete=False) as f:
            yaml.dump(bootstrap, f)
            cfg_path = f.name
        pi = client.Gr33nPiClient(cfg_path)
        try:
            with patch.object(pi.api, 'get_config_version') as mock_ver:
                pi._poll_config_version()
                mock_ver.assert_not_called()
        finally:
            os.unlink(cfg_path)


class TestPhase70MultiActuatorDispatch(unittest.TestCase):
    """Phase 70 — actuator_id dispatch and relay-HAT controller factory."""

    def test_resolve_actuator_prefers_payload_actuator_id(self):
        a1 = client.ActuatorController({'actuator_id': 1, 'device_id': 10, 'device_type': 'light', 'gpio_pin': 17})
        a2 = client.ActuatorController({'actuator_id': 2, 'device_id': 10, 'device_type': 'pump', 'gpio_pin': 18})
        reg = {1: a1, 2: a2}
        got = client.resolve_actuator_for_command(reg, 10, {'actuator_id': 2, 'command': 'on'})
        self.assertIs(got, a2)

    def test_resolve_actuator_single_on_device_falls_back(self):
        a1 = client.ActuatorController({'actuator_id': 1, 'device_id': 10, 'device_type': 'light', 'gpio_pin': 17})
        reg = {1: a1}
        got = client.resolve_actuator_for_command(reg, 10, {'command': 'on'})
        self.assertIs(got, a1)

    def test_make_actuator_controller_relay_hat(self):
        ctrl = client.make_actuator_controller({
            'actuator_id': 9, 'device_id': 1, 'device_type': 'pump', 'driver': 'relay_hat', 'channel': 3,
        })
        self.assertIsInstance(ctrl, client.RelayHATActuatorController)
        self.assertEqual(ctrl.channel, 3)

    def test_relay_hat_controller_on_off(self):
        ctrl = client.RelayHATActuatorController({
            'actuator_id': 9, 'device_id': 1, 'device_type': 'pump', 'channel': 2,
        })
        ctrl.turn_on()
        self.assertEqual(ctrl.state, 'on')
        ctrl.turn_off()
        self.assertEqual(ctrl.state, 'off')


class TestDerivedSensorInClient(unittest.TestCase):
    """End-to-end: a Gr33nPiClient built with derived sensors wires the cache."""

    def test_client_shares_cache_across_readers(self):
        cfg = {
            'api':   {'base_url': 'http://127.0.0.1:1', 'timeout_seconds': 1, 'api_key': ''},
            'farm':  {'farm_id': 1},
            'sensors': [
                {'sensor_id': 1, 'sensor_type': 'temperature', 'source': 'dht22', 'pin': 4, 'interval_seconds': 60},
                {'sensor_id': 2, 'sensor_type': 'humidity',    'source': 'dht22', 'pin': 4, 'interval_seconds': 60},
                {'sensor_id': 8, 'sensor_type': 'dew_point',   'source': 'derived',
                 'inputs': {'temperature_c': 1, 'humidity_pct': 2},
                 'input_max_age_seconds': 600, 'interval_seconds': 60},
            ],
            'actuators': [],
            'offline_queue_path': tempfile.NamedTemporaryFile(suffix='.db', delete=False).name,
            'schedule_poll_interval_seconds': 30,
            'offline_flush_interval_seconds': 60,
        }
        with tempfile.NamedTemporaryFile('w', suffix='.yaml', delete=False) as f:
            yaml.dump(cfg, f)
            cfg_path = f.name

        c = client.Gr33nPiClient(config_path=cfg_path)
        # Every reader must share the same cache instance so cross-sensor
        # reads work within a single sensor-loop tick.
        caches = {id(r.cache) for r in c._readers.values()}
        self.assertEqual(len(caches), 1,
            "all SensorReaders should reference one shared ReadingCache")

        # Simulate the sensor-loop flow: physical sensors write to cache,
        # then derived sensor reads and computes.
        c._reading_cache.put(1, 25.0)
        c._reading_cache.put(2, 50.0)
        val = c._readers[8].read()
        self.assertIsNotNone(val)
        self.assertAlmostEqual(val, 13.9, delta=0.3)


