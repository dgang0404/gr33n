#!/usr/bin/env python3
"""Tests for import_config_to_platform.py (Phase 51 WS5)."""

import json
import tempfile
import unittest
from unittest.mock import MagicMock, patch

import yaml

import gr33n_client as client
import import_config_to_platform as importer


class TestWiringMappers(unittest.TestCase):

    def test_pi_sensor_entry_to_wiring_dht22(self):
        w = client.pi_sensor_entry_to_wiring(
            {'sensor_id': 3, 'source': 'dht22', 'pin': 4}, 7)
        self.assertEqual(w['source'], 'dht22')
        self.assertEqual(w['gpio_pin'], 4)
        self.assertEqual(w['device_id'], 7)

    def test_pi_sensor_entry_to_wiring_ads1115(self):
        w = client.pi_sensor_entry_to_wiring(
            {'sensor_id': 8, 'source': 'ads1115', 'channel': 1}, 7)
        self.assertEqual(w['i2c_channel'], 1)

    def test_pi_actuator_entry_to_wiring(self):
        w = client.pi_actuator_entry_to_wiring(
            {'actuator_id': 1, 'device_id': 7, 'gpio_pin': 17})
        self.assertEqual(w['source'], 'gpio_relay')
        self.assertEqual(w['gpio_pin'], 17)
        self.assertEqual(w['device_id'], 7)

    def test_build_minimal_bootstrap_strips_wiring(self):
        cfg = {
            'api': {'base_url': 'http://x', 'api_key': 'k'},
            'farm': {'farm_id': 1},
            'device': {'uid': 'pi-1'},
            'sensors': [{'sensor_id': 1}],
            'actuators': [{'actuator_id': 2}],
            'schedule_poll_interval_seconds': 30,
        }
        minimal = client.build_minimal_bootstrap(cfg)
        self.assertNotIn('sensors', minimal)
        self.assertNotIn('actuators', minimal)
        self.assertEqual(minimal['device']['uid'], 'pi-1')


class TestImportConfigToPlatform(unittest.TestCase):

    def _write_config(self, data: dict) -> str:
        f = tempfile.NamedTemporaryFile('w', suffix='.yaml', delete=False)
        yaml.dump(data, f)
        f.close()
        return f.name

    @patch.object(importer, '_patch_actuator_wiring')
    @patch.object(importer, '_patch_sensor_wiring')
    @patch.object(importer, '_resolve_device_id', return_value=7)
    @patch.object(importer, '_api_session')
    def test_import_patches_and_writes_minimal_yaml(
        self, mock_session, mock_resolve, mock_patch_s, mock_patch_a,
    ):
        cfg_path = self._write_config({
            'api': {'base_url': 'http://127.0.0.1:8080', 'api_key': 'pi-key'},
            'farm': {'farm_id': 1},
            'device': {'uid': 'demo-pi'},
            'sensors': [
                {'sensor_id': 3, 'sensor_type': 'temperature', 'source': 'dht22',
                 'pin': 4, 'interval_seconds': 60},
            ],
            'actuators': [
                {'actuator_id': 1, 'device_id': 7, 'device_type': 'light', 'gpio_pin': 17},
            ],
        })
        out_path = tempfile.mktemp(suffix='.yaml')
        try:
            mock_session.return_value = MagicMock()
            summary = importer.import_wiring(
                cfg_path, jwt='test-token', output_path=out_path)
            self.assertEqual(summary['sensors_imported'], [3])
            self.assertEqual(summary['actuators_imported'], [1])
            mock_patch_s.assert_called_once()
            mock_patch_a.assert_called_once()
            with open(out_path) as fh:
                minimal = yaml.safe_load(fh)
            self.assertNotIn('sensors', minimal)
            self.assertNotIn('actuators', minimal)
            self.assertEqual(minimal['device']['uid'], 'demo-pi')
        finally:
            import os
            os.unlink(cfg_path)
            if os.path.exists(out_path):
                os.unlink(out_path)

    def test_dry_run_does_not_write(self):
        cfg_path = self._write_config({
            'api': {'base_url': 'http://127.0.0.1:8080'},
            'farm': {'farm_id': 1},
            'device': {'uid': 'demo-pi'},
            'sensors': [{'sensor_id': 1, 'source': 'bh1750'}],
            'actuators': [],
        })
        with patch.object(importer, '_resolve_device_id', return_value=7), \
             patch.object(importer, '_api_session', return_value=MagicMock()), \
             patch.object(importer, '_patch_sensor_wiring') as mock_patch:
            summary = importer.import_wiring(cfg_path, jwt='t', dry_run=True)
            mock_patch.assert_not_called()
            self.assertIn('minimal_config_preview', summary)
        import os
        os.unlink(cfg_path)

    def test_requires_device_uid(self):
        cfg_path = self._write_config({
            'api': {'base_url': 'http://x'},
            'farm': {'farm_id': 1},
            'sensors': [{'sensor_id': 1, 'source': 'dht22', 'pin': 4}],
        })
        with self.assertRaises(SystemExit):
            importer.import_wiring(cfg_path, jwt='t')
        import os
        os.unlink(cfg_path)


if __name__ == '__main__':
    unittest.main()
