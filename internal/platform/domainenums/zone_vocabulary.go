package domainenums

// ZoneTypeOption is one zone_type value with wizard visibility and optional hint.
type ZoneTypeOption struct {
	Value          string `json:"value"`
	Label          string `json:"label"`
	WizardVisible  bool   `json:"wizard_visible"`
	Hint           string `json:"hint,omitempty"`
}

// HintOption is an enum value with an optional farmer-facing hint.
type HintOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Hint  string `json:"hint,omitempty"`
}

type zoneTypeDef struct {
	value         string
	label         string
	wizardVisible bool
	hint          string
}

var zoneTypeDefs = []zoneTypeDef{
	{"indoor", "Indoor grow zone", true, "Tent, rack, or indoor bay"},
	{"greenhouse", "Greenhouse", true, "Glazing, shade, vents, and climate profile"},
	{"outdoor", "Outdoor", true, "Garden bed, field, or patio grow"},
	{"nursery", "Nursery", false, ""},
	{"seedling", "Seedling room", false, ""},
	{"veg", "Veg room (legacy)", false, ""},
	{"flower", "Flower room (legacy)", false, ""},
	{"storage", "Storage", false, ""},
}

var greenhouseCoverTypes = []Option{
	{Value: "glass", Label: "Glass"},
	{Value: "polycarbonate", Label: "Polycarbonate"},
	{Value: "film", Label: "Film / poly"},
}

var greenhouseAutomationPolicies = []HintOption{
	{Value: "manual", Label: "Manual only", Hint: "You control shade and fans"},
	{Value: "auto", Label: "Auto (sensor rules)", Hint: "Uses lux/temp sensors when wired"},
	{Value: "schedule_only", Label: "Schedule only", Hint: "Time-based, not sensor-driven"},
}

func zoneTypeOptions() []ZoneTypeOption {
	out := make([]ZoneTypeOption, len(zoneTypeDefs))
	for i, d := range zoneTypeDefs {
		out[i] = ZoneTypeOption{
			Value:         d.value,
			Label:         d.label,
			WizardVisible: d.wizardVisible,
			Hint:          d.hint,
		}
	}
	return out
}

func optionValues(opts []Option) map[string]struct{} {
	m := make(map[string]struct{}, len(opts))
	for _, o := range opts {
		m[o.Value] = struct{}{}
	}
	return m
}

func hintOptionValues(opts []HintOption) map[string]struct{} {
	m := make(map[string]struct{}, len(opts))
	for _, o := range opts {
		m[o.Value] = struct{}{}
	}
	return m
}

// IsValidGreenhouseCoverType reports whether cover_type is allowed in greenhouse_climate meta.
func IsValidGreenhouseCoverType(value string) bool {
	_, ok := optionValues(greenhouseCoverTypes)[value]
	return ok
}

// IsValidGreenhouseAutomationPolicy reports whether automation_policy is allowed.
func IsValidGreenhouseAutomationPolicy(value string) bool {
	_, ok := hintOptionValues(greenhouseAutomationPolicies)[value]
	return ok
}

// GreenhouseCoverTypeLabel returns the farmer-facing label for a cover_type value.
func GreenhouseCoverTypeLabel(value string) string {
	return optionLabel(greenhouseCoverTypes, value)
}

// GreenhouseAutomationPolicyLabel returns the farmer-facing label for automation_policy.
func GreenhouseAutomationPolicyLabel(value string) string {
	for _, o := range greenhouseAutomationPolicies {
		if o.Value == value {
			return o.Label
		}
	}
	return humanize(value)
}

func optionLabel(opts []Option, value string) string {
	for _, o := range opts {
		if o.Value == value {
			return o.Label
		}
	}
	return humanize(value)
}
