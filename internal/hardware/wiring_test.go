package hardware

import (
	"encoding/json"
	"testing"
)

func TestMergeAndExtractWiring(t *testing.T) {
	pin := 4
	dev := int64(1)
	w := &Wiring{Source: "dht22", GPIOPin: &pin, DeviceID: &dev, Notes: "Air temp"}
	cfg, err := MergeWiring(json.RawMessage(`{"notes":"grow"}`), w)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ExtractWiring(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != "dht22" || got.GPIOPin == nil || *got.GPIOPin != 4 {
		t.Fatalf("got %+v", got)
	}
	cleared, err := MergeWiring(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ex, _ := ExtractWiring(cleared); ex != nil {
		t.Fatalf("expected wiring cleared, got %+v", ex)
	}
}

func TestValidateSensorWiring(t *testing.T) {
	pin := 4
	if err := ValidateSensorWiring(&Wiring{Source: "dht22", GPIOPin: &pin}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSensorWiring(&Wiring{Source: "dht22"}); err == nil {
		t.Fatal("expected gpio required")
	}
	ch := 1
	if err := ValidateSensorWiring(&Wiring{Source: "ads1115", I2CChannel: &ch}); err != nil {
		t.Fatal(err)
	}
}

func TestFormatLabel(t *testing.T) {
	pin := 17
	label := FormatLabel(&Wiring{Source: "gpio_relay", GPIOPin: &pin})
	if label != "GPIO_RELAY · BCM GPIO 17" {
		t.Fatalf("got %q", label)
	}
}
