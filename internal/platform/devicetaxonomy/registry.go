// Package devicetaxonomy loads gr33ncore.device_type_registry (Phase 90).
package devicetaxonomy

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry is one registry row.
type Entry struct {
	TypeKey        string   `json:"type_key"`
	DeviceClass    string   `json:"device_class"`
	PlantNeed      string   `json:"plant_need"`
	DisplayLabel   string   `json:"display_label"`
	SupportsPulse  bool     `json:"supports_pulse"`
	GHRole         *string  `json:"gh_role,omitempty"`
	WiringSources  []string `json:"wiring_sources,omitempty"`
	SortOrder      int      `json:"sort_order"`
}

// WiringSourceOption is a Pi wiring driver choice for sensor setup.
type WiringSourceOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Payload is returned by GET /platform/device-taxonomy.
type Payload struct {
	Sensors             []Entry              `json:"sensors"`
	Actuators           []Entry              `json:"actuators"`
	ByPlantNeed         map[string][]Entry   `json:"by_plant_need"`
	WiringSourceOptions []WiringSourceOption `json:"wiring_source_options"`
}

// Registry is an indexed view of device_type_registry.
type Registry struct {
	byKey map[string]Entry
	all   []Entry
}

var (
	cacheMu sync.RWMutex
	cached  *Registry
)

var wiringSourceLabels = map[string]string{
	"dht22":        "DHT22 (temp / humidity)",
	"ads1115":      "ADS1115 (analog)",
	"mhz19":        "MH-Z19 (CO₂ serial)",
	"bh1750":       "BH1750 (light I2C)",
	"gpio_digital": "GPIO digital",
	"derived":      "Derived (computed)",
}

func normKey(s string) string {
	return strings.ToLower(strings.TrimSpace(strings.ReplaceAll(s, " ", "_")))
}

// Load reads the registry from Postgres.
func Load(ctx context.Context, pool *pgxpool.Pool) (*Registry, error) {
	if pool == nil {
		return nil, fmt.Errorf("device taxonomy: nil pool")
	}
	cacheMu.RLock()
	if cached != nil {
		r := cached
		cacheMu.RUnlock()
		return r, nil
	}
	cacheMu.RUnlock()

	rows, err := pool.Query(ctx, `
SELECT type_key, device_class, plant_need, display_label, supports_pulse, gh_role, wiring_sources, sort_order
FROM gr33ncore.device_type_registry
ORDER BY sort_order, type_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reg := &Registry{byKey: make(map[string]Entry)}
	for rows.Next() {
		var e Entry
		var ghRole *string
		var wiringJSON []byte
		if err := rows.Scan(&e.TypeKey, &e.DeviceClass, &e.PlantNeed, &e.DisplayLabel, &e.SupportsPulse, &ghRole, &wiringJSON, &e.SortOrder); err != nil {
			return nil, err
		}
		e.GHRole = ghRole
		if len(wiringJSON) > 0 {
			_ = json.Unmarshal(wiringJSON, &e.WiringSources)
		}
		key := normKey(e.TypeKey)
		reg.byKey[key] = e
		reg.all = append(reg.all, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cached = reg
	cacheMu.Unlock()
	return reg, nil
}

// ResetCache clears the in-process cache (tests).
func ResetCache() {
	cacheMu.Lock()
	cached = nil
	cacheMu.Unlock()
}

func (r *Registry) entry(deviceClass, typeKey string) (Entry, bool) {
	if r == nil {
		return Entry{}, false
	}
	e, ok := r.byKey[normKey(typeKey)]
	if !ok {
		return Entry{}, false
	}
	if deviceClass != "" && e.DeviceClass != deviceClass {
		return Entry{}, false
	}
	return e, true
}

// PlantNeed returns water/light/air for a device type (fallback air).
func (r *Registry) PlantNeed(deviceClass, typeKey string) string {
	if e, ok := r.entry(deviceClass, typeKey); ok {
		return e.PlantNeed
	}
	return fallbackPlantNeed(deviceClass, typeKey)
}

// DisplayLabel returns a farmer-facing label for a type key.
func (r *Registry) DisplayLabel(deviceClass, typeKey string) string {
	if e, ok := r.entry(deviceClass, typeKey); ok {
		return e.DisplayLabel
	}
	return humanize(typeKey)
}

// SupportsPulse reports timed pulse commands for an actuator type.
func (r *Registry) SupportsPulse(actuatorType string) bool {
	if e, ok := r.entry("actuator", actuatorType); ok {
		return e.SupportsPulse
	}
	return fallbackSupportsPulse(actuatorType)
}

// GHRole returns shade|vent|fan when set on an actuator type.
func (r *Registry) GHRole(actuatorType string) string {
	if e, ok := r.entry("actuator", actuatorType); ok && e.GHRole != nil {
		return *e.GHRole
	}
	return ""
}

// Payload builds the API response shape.
func (r *Registry) Payload() Payload {
	sensors := make([]Entry, 0)
	actuators := make([]Entry, 0)
	byNeed := map[string][]Entry{
		"water": {},
		"light": {},
		"air":   {},
	}
	wiringSet := map[string]struct{}{}

	for _, e := range r.all {
		switch e.DeviceClass {
		case "sensor":
			sensors = append(sensors, e)
		case "actuator":
			actuators = append(actuators, e)
		}
		byNeed[e.PlantNeed] = append(byNeed[e.PlantNeed], e)
		for _, w := range e.WiringSources {
			wiringSet[w] = struct{}{}
		}
	}

	wiringOpts := make([]WiringSourceOption, 0, len(wiringSet))
	for k := range wiringSet {
		lbl := wiringSourceLabels[k]
		if lbl == "" {
			lbl = humanize(k)
		}
		wiringOpts = append(wiringOpts, WiringSourceOption{Value: k, Label: lbl})
	}
	sort.Slice(wiringOpts, func(i, j int) bool { return wiringOpts[i].Value < wiringOpts[j].Value })

	return Payload{
		Sensors:             sensors,
		Actuators:           actuators,
		ByPlantNeed:         byNeed,
		WiringSourceOptions: wiringOpts,
	}
}

// NeedSectionTitle maps plant_need to Guardian/UI section headers.
func NeedSectionTitle(need string) string {
	switch need {
	case "water":
		return "Feed & water sensors"
	case "light":
		return "Light sensors"
	default:
		return "Air & climate sensors"
	}
}

func humanize(s string) string {
	return strings.ReplaceAll(normKey(s), "_", " ")
}

func fallbackPlantNeed(deviceClass, typeKey string) string {
	t := normKey(typeKey)
	if deviceClass == "sensor" {
		if strings.Contains(t, "moisture") || strings.Contains(t, "ec") || strings.Contains(t, "ph") || strings.Contains(t, "water") {
			return "water"
		}
		if strings.Contains(t, "lux") || strings.Contains(t, "par") || strings.Contains(t, "light") {
			return "light"
		}
		return "air"
	}
	if strings.Contains(t, "pump") && !strings.Contains(t, "air") {
		return "water"
	}
	if strings.Contains(t, "valve") || t == "relay" {
		return "water"
	}
	if strings.Contains(t, "light") {
		return "light"
	}
	return "air"
}

func fallbackSupportsPulse(actuatorType string) bool {
	t := normKey(actuatorType)
	if t == "relay" || strings.Contains(t, "pump") {
		return true
	}
	return false
}

// LoadPayload is a convenience for handlers.
func LoadPayload(ctx context.Context, pool *pgxpool.Pool) (Payload, error) {
	reg, err := Load(ctx, pool)
	if err != nil {
		return Current().Payload(), nil
	}
	if reg == nil || len(reg.all) == 0 {
		return Current().Payload(), nil
	}
	return reg.Payload(), nil
}
