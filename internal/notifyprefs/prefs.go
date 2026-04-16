package notifyprefs

import (
	"encoding/json"
	"strings"

	db "gr33n-api/internal/db"
)

// Notify is stored under profiles.preferences JSON as key "notify".
type Notify struct {
	PushEnabled bool   `json:"push_enabled"`
	MinPriority string `json:"min_priority"`
}

func Defaults() Notify {
	return Notify{PushEnabled: false, MinPriority: "medium"}
}

func FromPreferencesJSON(raw []byte) Notify {
	out := Defaults()
	if len(raw) == 0 {
		return out
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return out
	}
	nraw, ok := root["notify"]
	if !ok {
		return out
	}
	var n Notify
	if err := json.Unmarshal(nraw, &n); err != nil {
		return out
	}
	n.MinPriority = normalizePriority(n.MinPriority)
	if n.MinPriority == "" {
		n.MinPriority = out.MinPriority
	}
	return n
}

func normalizePriority(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "low", "medium", "high", "critical":
		return s
	default:
		return ""
	}
}

// SetNotify merges the notify object into raw profile preferences JSON, preserving other keys.
func SetNotify(raw []byte, n Notify) ([]byte, error) {
	var root map[string]json.RawMessage
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &root)
	}
	if root == nil {
		root = map[string]json.RawMessage{}
	}
	n.MinPriority = normalizePriority(n.MinPriority)
	if n.MinPriority == "" {
		n.MinPriority = Defaults().MinPriority
	}
	sub, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	root["notify"] = sub
	return json.Marshal(root)
}

func PriorityRankString(s string) int {
	switch normalizePriority(s) {
	case "low":
		return 0
	case "medium":
		return 1
	case "high":
		return 2
	case "critical":
		return 3
	default:
		return 1
	}
}

func PriorityRankDB(p db.Gr33ncoreNotificationPriorityEnum) int {
	switch p {
	case db.Gr33ncoreNotificationPriorityEnumLow:
		return 0
	case db.Gr33ncoreNotificationPriorityEnumMedium:
		return 1
	case db.Gr33ncoreNotificationPriorityEnumHigh:
		return 2
	case db.Gr33ncoreNotificationPriorityEnumCritical:
		return 3
	default:
		return 1
	}
}

func AlertMeetsMinPriority(alert db.Gr33ncoreAlertsNotification, minPriority string) bool {
	ar := PriorityRankDB(db.Gr33ncoreNotificationPriorityEnumMedium)
	if alert.Severity.Valid {
		ar = PriorityRankDB(alert.Severity.Gr33ncoreNotificationPriorityEnum)
	}
	return ar >= PriorityRankString(minPriority)
}
