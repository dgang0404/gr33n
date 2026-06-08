package farmguardian

import (
	"strings"
	"testing"
	"time"

	"gr33n-api/internal/ai"
)

func TestShouldRunSummarizeDeviceHealthReadIntent(t *testing.T) {
	if !shouldRunSummarizeDeviceHealthReadIntent("why is my temp sensor stuck", nil) {
		t.Fatal("expected stuck sensor intent")
	}
	if !shouldRunSummarizeDeviceHealthReadIntent("Pi is offline", nil) {
		t.Fatal("expected offline intent")
	}
	if !shouldRunSummarizeDeviceHealthReadIntent("", &ContextRef{Type: "route", Path: "/sensors"}) {
		t.Fatal("expected sensors route context")
	}
	if !shouldRunSummarizeDeviceHealthReadIntent("fan not responding", nil) {
		t.Fatal("expected actuator intent")
	}
	if shouldRunSummarizeDeviceHealthReadIntent("hello", nil) {
		t.Fatal("expected no match for greeting")
	}
}

func TestReadToolIDs_IncludesSummarizeDeviceHealth(t *testing.T) {
	for _, id := range ReadToolIDs() {
		if id == "summarize_device_health" {
			return
		}
	}
	t.Fatal("summarize_device_health missing from ReadToolIDs")
}

func TestPlatformContextBlock_IncludesDeviceHealthRule(t *testing.T) {
	block := PlatformContextBlock(ai.Config{Enabled: true}, true, ReadToolIDs())
	if !strings.Contains(block, "summarize_device_health") || !strings.Contains(block, "GPIO") {
		t.Fatalf("platform context missing device health rule")
	}
}

func TestFormatRelayChannelLabel(t *testing.T) {
	got := formatRelayChannelLabel("5")
	if !strings.Contains(got, "ch 5") || !strings.Contains(got, "stack 0") || !strings.Contains(got, "relay 6") {
		t.Fatalf("relay label unexpected: %q", got)
	}
}

func TestSensorReadingStale(t *testing.T) {
	interval := int32(60)
	if !sensorReadingStale(4*time.Minute, &interval) {
		t.Fatal("expected stale when > 3x interval")
	}
	if sensorReadingStale(2*time.Minute, &interval) {
		t.Fatal("expected fresh within 3x interval")
	}
	if !sensorReadingStale(20*time.Minute, nil) {
		t.Fatal("expected stale fallback when interval unknown")
	}
}

func TestDeviceHealthRouteContext(t *testing.T) {
	if !deviceHealthRouteContext(&ContextRef{Type: "route", Path: "/pi-setup"}) {
		t.Fatal("expected pi-setup route")
	}
	if deviceHealthRouteContext(&ContextRef{Type: "route", Path: "/feeding"}) {
		t.Fatal("feeding route should not auto-trigger")
	}
}
