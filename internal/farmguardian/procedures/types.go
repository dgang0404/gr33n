package procedures

import (
	"encoding/json"
	"strings"

	"gr33n-api/internal/farmguardian"
)

// Procedure is an authored step-by-step field walkthrough (Phase 37 WS3).
type Procedure struct {
	ID        string `yaml:"id" json:"id"`
	Title     string `yaml:"title" json:"title"`
	Domain    string `yaml:"domain" json:"domain"`
	OfflineOK bool   `yaml:"offline_ok" json:"offline_ok"`
	Steps     []Step `yaml:"steps" json:"steps"`
}

// Step is one confirmable instruction in a procedure.
type Step struct {
	N                   int    `yaml:"n" json:"n"`
	SafetyTier          string `yaml:"safety_tier" json:"safety_tier"`
	Say                 string `yaml:"say" json:"say"`
	Confirm             string `yaml:"confirm" json:"confirm"`
	Ref                 string `yaml:"ref,omitempty" json:"ref,omitempty"`
	StopUnlessQualified bool   `yaml:"stop_unless_qualified,omitempty" json:"stop_unless_qualified,omitempty"`
}

// ActiveState is persisted under conversation_sessions.meta.active_procedure.
type ActiveState struct {
	ID     string `json:"id"`
	StepN  int    `json:"step_n"`
	Status string `json:"status"` // active | safety_stopped | completed | stopped
}

const (
	StatusActive        = "active"
	StatusSafetyStopped = "safety_stopped"
	StatusCompleted     = "completed"
	StatusStopped       = "stopped"
)

// SessionMeta is the JSON shape stored on conversation_sessions.meta.
type SessionMeta struct {
	Active *ActiveState `json:"active_procedure,omitempty"`
}

// TurnPayload is returned on /v1/chat when a procedure turn is handled without the LLM.
type TurnPayload struct {
	ProcedureID   string `json:"procedure_id"`
	Title         string `json:"title"`
	StepN         int    `json:"step_n"`
	StepTotal     int    `json:"step_total"`
	SafetyTier    string `json:"safety_tier,omitempty"`
	Say           string `json:"say,omitempty"`
	Confirm       string `json:"confirm,omitempty"`
	Ref           string `json:"ref,omitempty"`
	Status        string `json:"status"`
	SafetyStopped bool   `json:"safety_stopped,omitempty"`
	PrintPath     string `json:"print_path,omitempty"`
}

func ParseSessionMeta(raw []byte) SessionMeta {
	if len(raw) == 0 {
		return SessionMeta{}
	}
	var m SessionMeta
	_ = json.Unmarshal(raw, &m)
	return m
}

func (m SessionMeta) Marshal() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return []byte("{}")
	}
	return b
}

// NormalizeStepTier fills empty tiers as safe.
func (s *Step) NormalizeStepTier() string {
	t := farmguardian.SafetyTierSafe
	if s.SafetyTier != "" {
		t = strings.TrimSpace(strings.ToLower(s.SafetyTier))
	}
	if s.StopUnlessQualified {
		t = farmguardian.SafetyTierQualifiedPersonRequired
	}
	return t
}
