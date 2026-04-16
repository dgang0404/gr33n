#!/usr/bin/env python3
"""Tests for mqtt_telemetry_bridge — no broker required."""

import os
import sys

sys.path.insert(0, os.path.dirname(__file__))

import mqtt_telemetry_bridge as bridge


def test_parse_telemetry_topic_ok():
    assert bridge.parse_telemetry_topic(
        "gr33n/7/gw-a/telemetry/temp", 7
    ) == ("gw-a", "temp")
    assert bridge.parse_telemetry_topic(
        "gr33n/7/gw-a/telemetry/indoor/rh", 7
    ) == ("gw-a", "indoor/rh")


def test_parse_telemetry_topic_wrong_farm():
    assert bridge.parse_telemetry_topic("gr33n/99/gw-a/telemetry/temp", 7) is None


def test_parse_telemetry_topic_custom_prefix():
    assert bridge.parse_telemetry_topic("acme/3/d1/telemetry/x", 3, prefix="acme") == (
        "d1",
        "x",
    )
    assert bridge.parse_telemetry_topic("gr33n/3/d1/telemetry/x", 3, prefix="acme") is None


def test_extract_value_json():
    assert bridge.extract_value(b'{"v": 1.25}') == 1.25
    assert bridge.extract_value(b'{"value_raw": -3}') == -3.0
    assert bridge.extract_value(b'{"value": 0}') == 0.0
    assert bridge.extract_value(b"42.5") == 42.5


def test_extract_value_bad():
    assert bridge.extract_value(b"") is None
    assert bridge.extract_value(b"{}") is None
    assert bridge.extract_value(b"not-a-number") is None


def test_sensor_map_from_data():
    m = bridge.sensor_map_from_data(
        {
            "sensor_map": [
                {"device_uid": "a", "slug": "t", "sensor_id": 5},
                {"device_uid": "b", "slug": "x"},
            ]
        }
    )
    assert m == {("a", "t"): 5}


def test_resolve_sensor_id_numeric_slug():
    m = {("d", "hum"): 9}
    assert bridge.resolve_sensor_id(m, "d", "101") == 101
    assert bridge.resolve_sensor_id(m, "d", "hum") == 9
    assert bridge.resolve_sensor_id(m, "d", "unknown") is None
