package farmguardian

import (
	"fmt"
	"unicode/utf8"
	"strings"
)

const (
	// PromptCompletionReserve tokens reserved for the model reply when sizing the prompt.
	PromptCompletionReserve = 512
	// CharsPerTokenEstimate is a coarse heuristic for pre-flight prompt sizing.
	CharsPerTokenEstimate = 4
)

// SnapshotBudgetLimits caps snapshot verbosity for small-context models (Phase 122).
type SnapshotBudgetLimits struct {
	MaxZones           int
	MaxAlertDetails    int
	MaxProgramsPerZone int
	MaxProgramZones    int
	MaxCycles          int
	MaxPlantNames      int
}

// PromptBudget is the per-turn assembly limits after context-window trimming.
type PromptBudget struct {
	MaxHistoryTurns int
	RAGTopK         int
	Snapshot        SnapshotBudgetLimits
}

// DefaultPromptBudget returns full-size limits (no trimming).
func DefaultPromptBudget(maxHistoryTurns int) PromptBudget {
	if maxHistoryTurns <= 0 {
		maxHistoryTurns = 20
	}
	return PromptBudget{
		MaxHistoryTurns: maxHistoryTurns,
		RAGTopK:         RAGTopK,
		Snapshot: SnapshotBudgetLimits{
			MaxZones:           SnapshotMaxZones,
			MaxAlertDetails:    SnapshotMaxAlertDetails,
			MaxProgramsPerZone: SnapshotMaxProgramsPerZone,
			MaxProgramZones:    SnapshotMaxProgramZones,
			MaxCycles:          SnapshotMaxCycles,
			MaxPlantNames:      SnapshotMaxPlantNames,
		},
	}
}

// ComputePromptBudget derives trim limits from the resolved model context window.
// When contextWindow is 0 (unknown) or >= GuardianMinContextWindow, no trimming.
func ComputePromptBudget(contextWindow, maxHistoryTurns int) (PromptBudget, []string) {
	full := DefaultPromptBudget(maxHistoryTurns)
	if contextWindow <= 0 || contextWindow >= GuardianMinContextWindow {
		return full, nil
	}
	var log []string
	out := full

	switch {
	case contextWindow < 4096:
		out.MaxHistoryTurns = minInt(maxHistoryTurns, 4)
		out.RAGTopK = 3
		out.Snapshot = SnapshotBudgetLimits{
			MaxZones:           4,
			MaxAlertDetails:    1,
			MaxProgramsPerZone: 1,
			MaxProgramZones:    2,
			MaxCycles:          2,
			MaxPlantNames:      3,
		}
	case contextWindow < 8192:
		out.MaxHistoryTurns = minInt(maxHistoryTurns, 8)
		out.RAGTopK = 5
		out.Snapshot = SnapshotBudgetLimits{
			MaxZones:           8,
			MaxAlertDetails:    2,
			MaxProgramsPerZone: 2,
			MaxProgramZones:    4,
			MaxCycles:          4,
			MaxPlantNames:      5,
		}
	}

	if out.MaxHistoryTurns < full.MaxHistoryTurns {
		log = append(log, fmt.Sprintf("history turns %d→%d", full.MaxHistoryTurns, out.MaxHistoryTurns))
	}
	if out.RAGTopK < full.RAGTopK {
		log = append(log, fmt.Sprintf("RAG topK %d→%d", full.RAGTopK, out.RAGTopK))
	}
	if out.Snapshot.MaxZones < full.Snapshot.MaxZones {
		log = append(log, fmt.Sprintf("snapshot caps reduced (context_window=%d)", contextWindow))
	}
	return out, log
}

// TrimSummary is returned on chat done when prompt budget trimming occurred (Phase 133 WS2).
type TrimSummary struct {
	HistoryTurns           string `json:"history_turns,omitempty"`
	RAGTopK                string `json:"rag_top_k,omitempty"`
	SnapshotReduced        bool   `json:"snapshot_reduced"`
	EffectiveContextWindow int    `json:"effective_context_window"`
}

// BuildTrimSummary builds a client-visible trim summary from budget computation.
func BuildTrimSummary(full, applied PromptBudget, trimLog []string, effectiveWindow int) *TrimSummary {
	if len(trimLog) == 0 {
		return nil
	}
	out := &TrimSummary{EffectiveContextWindow: effectiveWindow}
	if applied.MaxHistoryTurns < full.MaxHistoryTurns {
		out.HistoryTurns = fmt.Sprintf("%d→%d", full.MaxHistoryTurns, applied.MaxHistoryTurns)
	}
	if applied.RAGTopK < full.RAGTopK {
		out.RAGTopK = fmt.Sprintf("%d→%d", full.RAGTopK, applied.RAGTopK)
	}
	for _, line := range trimLog {
		if strings.Contains(line, "snapshot caps reduced") {
			out.SnapshotReduced = true
			break
		}
	}
	if out.HistoryTurns == "" && out.RAGTopK == "" && !out.SnapshotReduced {
		return nil
	}
	return out
}

// ApplyBudgetLimits truncates snapshot fields before PromptBlock rendering.
func (s *Snapshot) ApplyBudgetLimits(l SnapshotBudgetLimits) {
	if s == nil || l.MaxZones <= 0 {
		return
	}
	if len(s.ZoneNames) > l.MaxZones {
		s.ZoneNames = s.ZoneNames[:l.MaxZones]
	}
	if len(s.PlantNames) > l.MaxPlantNames && l.MaxPlantNames > 0 {
		s.PlantNames = s.PlantNames[:l.MaxPlantNames]
	}
	if len(s.ActiveCycles) > l.MaxCycles && l.MaxCycles > 0 {
		s.ActiveCycles = s.ActiveCycles[:l.MaxCycles]
	}
	if len(s.ProgramsByZone) > l.MaxProgramZones && l.MaxProgramZones > 0 {
		s.ProgramsByZone = s.ProgramsByZone[:l.MaxProgramZones]
	}
	for i := range s.ProgramsByZone {
		if len(s.ProgramsByZone[i].Programs) > l.MaxProgramsPerZone && l.MaxProgramsPerZone > 0 {
			s.ProgramsByZone[i].Programs = s.ProgramsByZone[i].Programs[:l.MaxProgramsPerZone]
		}
	}
	if len(s.UnreadAlertDetails) > l.MaxAlertDetails && l.MaxAlertDetails > 0 {
		s.UnreadAlertDetails = s.UnreadAlertDetails[:l.MaxAlertDetails]
	}
}

// EstimatePromptTokens returns a coarse token estimate for logging and tests.
func EstimatePromptTokens(parts ...string) int {
	total := 0
	for _, p := range parts {
		total += utf8.RuneCountInString(p)
	}
	if total == 0 {
		return 0
	}
	return (total + CharsPerTokenEstimate - 1) / CharsPerTokenEstimate
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
