// Phase 139 WS3/WS4 — dev turn inspector payload.

package farmguardian

import (
	db "gr33n-api/internal/db"
)

// TurnDebug is attached to chat done events and the turn debug API (dev/auth_test only).
type TurnDebug struct {
	RequestID               string         `json:"request_id,omitempty"`
	ToolsPlanned            []string       `json:"tools_planned,omitempty"`
	RAGChunks               map[string]int `json:"rag_chunks,omitempty"`
	RAGChunkTotal           int            `json:"rag_chunk_total,omitempty"`
	TrimSummary             *TrimSummary   `json:"trim_summary,omitempty"`
	Model                   string         `json:"model,omitempty"`
	EffectiveContextWindow  int            `json:"effective_context_window,omitempty"`
	AdvertisedContextWindow int            `json:"advertised_context_window,omitempty"`
	PromptBudget            *PromptBudget  `json:"prompt_budget,omitempty"`
}

// CountRAGChunksBySource groups retrieved chunks by source_type for the turn inspector.
func CountRAGChunksBySource(chunks []db.SearchRagNearestNeighborsFilteredRow) map[string]int {
	if len(chunks) == 0 {
		return nil
	}
	out := make(map[string]int)
	for _, c := range chunks {
		key := c.SourceType
		if key == "" {
			key = "unknown"
		}
		out[key]++
	}
	return out
}

// BuildTurnDebug assembles the dev-only debug block for one chat turn.
func BuildTurnDebug(
	requestID string,
	plan ToolPlan,
	chunks []db.SearchRagNearestNeighborsFilteredRow,
	trimSummary *TrimSummary,
	model string,
	effectiveWindow, advertisedWindow int,
	promptBudget PromptBudget,
) *TurnDebug {
	ragCounts := CountRAGChunksBySource(chunks)
	total := len(chunks)
	if total == 0 && len(plan.ToolIDs) == 0 && trimSummary == nil {
		return &TurnDebug{
			RequestID:               requestID,
			Model:                   model,
			EffectiveContextWindow:  effectiveWindow,
			AdvertisedContextWindow: advertisedWindow,
			PromptBudget:            &promptBudget,
		}
	}
	return &TurnDebug{
		RequestID:               requestID,
		ToolsPlanned:            append([]string(nil), plan.ToolIDs...),
		RAGChunks:               ragCounts,
		RAGChunkTotal:           total,
		TrimSummary:             trimSummary,
		Model:                   model,
		EffectiveContextWindow:  effectiveWindow,
		AdvertisedContextWindow: advertisedWindow,
		PromptBudget:            &promptBudget,
	}
}
