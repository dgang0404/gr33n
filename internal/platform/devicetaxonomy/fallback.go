package devicetaxonomy

// embeddedEntries mirrors db/migrations/20260622_phase90_device_type_registry.sql
// plus later additive registry migrations (e.g. phase183_gate_actuator_type.sql).
var embeddedEntries = []Entry{
	// sensors — water
	{TypeKey: "soil_moisture", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "Soil moisture", WiringSources: []string{"ads1115"}, SortOrder: 10},
	{TypeKey: "moisture", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "Moisture", WiringSources: []string{"ads1115"}, SortOrder: 11},
	{TypeKey: "ec", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "EC", WiringSources: []string{"ads1115"}, SortOrder: 12},
	{TypeKey: "ph", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "pH", WiringSources: []string{"ads1115"}, SortOrder: 13},
	{TypeKey: "water_level", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "Water level", WiringSources: []string{"gpio_digital"}, SortOrder: 14},
	{TypeKey: "flow_rate", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "Flow rate", WiringSources: []string{"gpio_digital"}, SortOrder: 15},
	{TypeKey: "water_temp", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "Water temperature", WiringSources: []string{"dht22"}, SortOrder: 16},
	{TypeKey: "dissolved_oxygen", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "Dissolved oxygen", WiringSources: []string{"ads1115"}, SortOrder: 17},
	{TypeKey: "tds", DeviceClass: "sensor", PlantNeed: "water", DisplayLabel: "TDS", WiringSources: []string{"ads1115"}, SortOrder: 18},
	// sensors — light
	{TypeKey: "lux", DeviceClass: "sensor", PlantNeed: "light", DisplayLabel: "Light level", WiringSources: []string{"bh1750"}, SortOrder: 30},
	{TypeKey: "par", DeviceClass: "sensor", PlantNeed: "light", DisplayLabel: "PAR", WiringSources: []string{"bh1750"}, SortOrder: 31},
	{TypeKey: "par_umol", DeviceClass: "sensor", PlantNeed: "light", DisplayLabel: "PAR", WiringSources: []string{"bh1750"}, SortOrder: 32},
	{TypeKey: "ppfd", DeviceClass: "sensor", PlantNeed: "light", DisplayLabel: "PPFD", WiringSources: []string{"bh1750"}, SortOrder: 33},
	{TypeKey: "light_level", DeviceClass: "sensor", PlantNeed: "light", DisplayLabel: "Light level", WiringSources: []string{"bh1750"}, SortOrder: 34},
	// sensors — air
	{TypeKey: "air_temp", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Air temperature", WiringSources: []string{"dht22"}, SortOrder: 50},
	{TypeKey: "temperature", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Temperature", WiringSources: []string{"dht22"}, SortOrder: 51},
	{TypeKey: "temp", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Temperature", WiringSources: []string{"dht22"}, SortOrder: 52},
	{TypeKey: "temp_f", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Temperature (°F)", WiringSources: []string{"dht22"}, SortOrder: 53},
	{TypeKey: "humidity", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Humidity", WiringSources: []string{"dht22"}, SortOrder: 54},
	{TypeKey: "rh", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Humidity", WiringSources: []string{"dht22"}, SortOrder: 55},
	{TypeKey: "co2", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "CO₂", WiringSources: []string{"mhz19"}, SortOrder: 56},
	{TypeKey: "vpd", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "VPD", WiringSources: []string{"derived"}, SortOrder: 57},
	{TypeKey: "dew_point", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Dew point", WiringSources: []string{"derived"}, SortOrder: 58},
	{TypeKey: "barometric_pressure", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Barometric pressure", WiringSources: []string{"derived"}, SortOrder: 59},
	{TypeKey: "pressure", DeviceClass: "sensor", PlantNeed: "air", DisplayLabel: "Pressure", WiringSources: []string{"derived"}, SortOrder: 60},
	// actuators — water
	{TypeKey: "pump", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Pump", SupportsPulse: true, SortOrder: 110},
	{TypeKey: "water_valve", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Water valve", SupportsPulse: true, SortOrder: 111},
	{TypeKey: "return_pump", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Return pump", SupportsPulse: true, SortOrder: 112},
	{TypeKey: "irrigation", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Irrigation", SortOrder: 113},
	{TypeKey: "drip", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Drip", SortOrder: 114},
	{TypeKey: "feeder_hopper", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Feeder hopper", SupportsPulse: true, SortOrder: 115},
	{TypeKey: "relay", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Relay", SupportsPulse: true, SortOrder: 116},
	{TypeKey: "air_pump", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Air pump", SupportsPulse: true, SortOrder: 117},
	// SupportsPulse: false — a gate is an open/shut toggle, not a timed-run
	// device (see phase210_gate_not_pulseable.sql for the backend mismatch this fixes).
	{TypeKey: "gate", DeviceClass: "actuator", PlantNeed: "water", DisplayLabel: "Gate", SupportsPulse: false, SortOrder: 118},
	// actuators — light
	{TypeKey: "light", DeviceClass: "actuator", PlantNeed: "light", DisplayLabel: "Light", SortOrder: 130},
	{TypeKey: "grow_light", DeviceClass: "actuator", PlantNeed: "light", DisplayLabel: "Grow light", SortOrder: 131},
	// actuators — air
	{TypeKey: "exhaust_fan", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Exhaust fan", GHRole: strPtr("fan"), SortOrder: 150},
	{TypeKey: "circulation_fan", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Circulation fan", GHRole: strPtr("fan"), SortOrder: 151},
	{TypeKey: "fan", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Fan", GHRole: strPtr("fan"), SortOrder: 152},
	{TypeKey: "ridge_vent", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Ridge vent", GHRole: strPtr("vent"), SortOrder: 153},
	{TypeKey: "glazing_panel", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Glazing panel", GHRole: strPtr("vent"), SortOrder: 154},
	{TypeKey: "shade_screen", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Shade screen", GHRole: strPtr("shade"), SortOrder: 155},
	{TypeKey: "shade_cloth_motor", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Shade cloth motor", GHRole: strPtr("shade"), SortOrder: 156},
	{TypeKey: "shade", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Shade", GHRole: strPtr("shade"), SortOrder: 157},
	{TypeKey: "vent", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Vent", GHRole: strPtr("vent"), SortOrder: 158},
	{TypeKey: "humidifier", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Humidifier", SortOrder: 159},
	{TypeKey: "dehumidifier", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Dehumidifier", SortOrder: 160},
	{TypeKey: "co2_injector", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "CO₂ injector", SortOrder: 161},
	{TypeKey: "heat_lamp", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Heat lamp", SortOrder: 162},
	{TypeKey: "heater", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Heater", SortOrder: 163},
	{TypeKey: "cooler", DeviceClass: "actuator", PlantNeed: "air", DisplayLabel: "Cooler", SortOrder: 164},
}

func strPtr(s string) *string { return &s }

func embeddedRegistry() *Registry {
	reg := &Registry{byKey: make(map[string]Entry, len(embeddedEntries))}
	for _, e := range embeddedEntries {
		reg.byKey[normKey(e.TypeKey)] = e
		reg.all = append(reg.all, e)
	}
	return reg
}

// Current returns the DB-backed registry when loaded, else the embedded seed.
func Current() *Registry {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	if cached != nil {
		return cached
	}
	return embeddedRegistry()
}
