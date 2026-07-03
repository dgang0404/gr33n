package devicetaxonomy

// HookupStep describes one wire connection for a Pi driver.
type HookupStep struct {
	Wire string `json:"wire"`
	To   string `json:"to"`
	Role string `json:"role"`
}

// DriverHookups returns per-driver physical wiring steps (Phase 121).
func DriverHookups() map[string][]HookupStep {
	return map[string][]HookupStep{
		"dht22": {
			{Wire: "VCC", To: "Physical pin 1 or 17 (3.3 V)", Role: "power3v3"},
			{Wire: "DATA", To: "Your chosen GPIO (BCM)", Role: "gpio"},
			{Wire: "GND", To: "Any GND pin (e.g. pin 6)", Role: "gnd"},
		},
		"ads1115": {
			{Wire: "VDD", To: "Physical pin 1 or 17 (3.3 V)", Role: "power3v3"},
			{Wire: "GND", To: "Any GND pin", Role: "gnd"},
			{Wire: "SDA", To: "Physical pin 3 (I²C SDA)", Role: "i2c_sda"},
			{Wire: "SCL", To: "Physical pin 5 (I²C SCL)", Role: "i2c_scl"},
			{Wire: "A0–A3", To: "Analog sensor signal (per channel)", Role: "analog_in"},
		},
		"bh1750": {
			{Wire: "VCC", To: "Physical pin 1 or 17 (3.3 V)", Role: "power3v3"},
			{Wire: "GND", To: "Any GND pin", Role: "gnd"},
			{Wire: "SDA", To: "Physical pin 3 (I²C SDA)", Role: "i2c_sda"},
			{Wire: "SCL", To: "Physical pin 5 (I²C SCL)", Role: "i2c_scl"},
		},
		"mhz19": {
			{Wire: "VIN", To: "Physical pin 2 or 4 (5 V)", Role: "power5v"},
			{Wire: "GND", To: "Any GND pin", Role: "gnd"},
			{Wire: "TX", To: "Physical pin 10 (Pi RX / GPIO15)", Role: "uart_rx"},
			{Wire: "RX", To: "Physical pin 8 (Pi TX / GPIO14)", Role: "uart_tx"},
		},
		"gpio_digital": {
			{Wire: "Signal", To: "Your chosen GPIO (BCM)", Role: "gpio"},
			{Wire: "VCC", To: "Sensor supply (3.3 V or 5 V per datasheet)", Role: "power3v3"},
			{Wire: "GND", To: "Any GND pin", Role: "gnd"},
		},
		"gpio_relay": {
			{Wire: "IN / coil", To: "GPIO pin or relay HAT channel output", Role: "gpio"},
			{Wire: "COM / NO", To: "Load wiring (mains or low-voltage per relay)", Role: "load"},
		},
		"relay_hat": {
			{Wire: "HAT stack", To: "40-pin header — seats on pins 1–40", Role: "hat"},
			{Wire: "I²C", To: "Uses pins 3 (SDA) and 5 (SCL) on the bus", Role: "i2c_sda"},
			{Wire: "DIP switches", To: "Set ID0–ID2 per stack level (see relay stack view)", Role: "dip"},
			{Wire: "Channel", To: "Assign relay channel 0–63 in wiring panel", Role: "relay_channel"},
		},
	}
}
