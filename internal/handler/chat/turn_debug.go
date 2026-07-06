// Phase 139 WS3/WS4 — dev turn inspector cache + GET debug endpoint.

package chat

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

const turnDebugCacheMax = 256

var (
	turnDebugMu    sync.Mutex
	turnDebugCache = map[string]farmguardian.TurnDebug{}
	turnDebugOrder []string
)

func debugModeEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_MODE"))) {
	case "dev", "auth_test":
		return true
	default:
		return false
	}
}

func turnDebugCacheKey(sessionID uuid.UUID, turnIndex int32) string {
	return fmt.Sprintf("%s:%d", sessionID.String(), turnIndex)
}

func storeTurnDebug(sessionID uuid.UUID, turnIndex int32, dbg farmguardian.TurnDebug) {
	if !debugModeEnabled() || turnIndex < 0 {
		return
	}
	key := turnDebugCacheKey(sessionID, turnIndex)
	turnDebugMu.Lock()
	defer turnDebugMu.Unlock()
	if _, ok := turnDebugCache[key]; !ok {
		turnDebugOrder = append(turnDebugOrder, key)
	}
	turnDebugCache[key] = dbg
	for len(turnDebugOrder) > turnDebugCacheMax {
		old := turnDebugOrder[0]
		turnDebugOrder = turnDebugOrder[1:]
		delete(turnDebugCache, old)
	}
}

func lookupTurnDebug(sessionID uuid.UUID, turnIndex int32) (farmguardian.TurnDebug, bool) {
	key := turnDebugCacheKey(sessionID, turnIndex)
	turnDebugMu.Lock()
	defer turnDebugMu.Unlock()
	dbg, ok := turnDebugCache[key]
	return dbg, ok
}

type turnDebugBuildInput struct {
	toolPlan         farmguardian.ToolPlan
	chunks           []db.SearchRagNearestNeighborsFilteredRow
	trimSummary      *farmguardian.TrimSummary
	model            string
	effectiveWindow  int
	advertisedWindow int
	promptBudget     farmguardian.PromptBudget
}

func buildTurnDebug(ctx context.Context, in turnDebugBuildInput) *farmguardian.TurnDebug {
	if !debugModeEnabled() {
		return nil
	}
	return farmguardian.BuildTurnDebug(
		authctx.RequestID(ctx),
		in.toolPlan,
		in.chunks,
		in.trimSummary,
		in.model,
		in.effectiveWindow,
		in.advertisedWindow,
		in.promptBudget,
	)
}

func attachTurnDebug(resp *postResponse, dbg *farmguardian.TurnDebug) {
	if dbg != nil {
		resp.Debug = dbg
	}
}

// GetTurnDebug handles GET /v1/chat/sessions/{session_id}/turns/{turn_index}/debug.
func (h *Handler) GetTurnDebug(w http.ResponseWriter, r *http.Request) {
	if !debugModeEnabled() {
		httputil.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	sessionID, err := uuid.Parse(r.PathValue("session_id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid session_id")
		return
	}
	turnIndex, err := strconv.ParseInt(r.PathValue("turn_index"), 10, 32)
	if err != nil || turnIndex < 0 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid turn_index")
		return
	}
	if h.q == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	rows, err := h.q.ListConversationTurnsBySession(r.Context(), db.ListConversationTurnsBySessionParams{
		SessionID: sessionID,
		UserID:    userID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load session")
		return
	}
	var found *db.ListConversationTurnsBySessionRow
	for i := range rows {
		if rows[i].TurnIndex == int32(turnIndex) {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		httputil.WriteError(w, http.StatusNotFound, "turn not found")
		return
	}
	if dbg, ok := lookupTurnDebug(sessionID, int32(turnIndex)); ok {
		httputil.WriteJSON(w, http.StatusOK, dbg)
		return
	}
	// Fallback when cache evicted — partial reconstruction from persisted turn.
	farmID := int64(0)
	if found.FarmID != nil {
		farmID = *found.FarmID
	}
	var snap farmguardian.Snapshot
	if farmID > 0 {
		if s, serr := farmguardian.BuildSnapshot(r.Context(), h.q, farmID); serr == nil {
			snap = s
		}
	}
	plan := farmguardian.PlanReadTools(found.UserMessage, nil, snap)
	dbg := farmguardian.BuildTurnDebug(
		"",
		plan,
		nil,
		nil,
		found.LlmModel,
		0,
		0,
		farmguardian.DefaultPromptBudget(MaxHistoryTurns),
	)
	dbg.RAGChunkTotal = int(found.ContextCount)
	httputil.WriteJSON(w, http.StatusOK, dbg)
}
