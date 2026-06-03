package chat

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

const (
	defaultProposalListLimit = 50
	maxProposalListLimit     = 100
)

// proposalListItem is the inbox/card shape for GET /v1/chat/proposals (Phase 30 WS1;
// revision lineage + blind-spot facts + impact added Phase 34).
type proposalListItem struct {
	ProposalID string         `json:"proposal_id"`
	Tool       string         `json:"tool"`
	Args       map[string]any `json:"args"`
	Summary    string         `json:"summary"`
	RiskTier   string         `json:"risk_tier"`
	ExpiresAt  time.Time      `json:"expires_at"`
	CreatedAt  time.Time      `json:"created_at"`
	FarmID     int64          `json:"farm_id"`
	Status     string         `json:"status"`

	Revision             int                         `json:"revision,omitempty"`
	SupersedesProposalID string                      `json:"supersedes_proposal_id,omitempty"`
	OperatorProvided     []farmguardian.OperatorFact `json:"operator_provided,omitempty"`
	ImpactSummary        []string                    `json:"impact_summary,omitempty"`
}

type proposalListResponse struct {
	Proposals []proposalListItem `json:"proposals"`
	Total     int64              `json:"total"`
	Limit     int32              `json:"limit"`
	Offset    int32              `json:"offset"`
}

// ListProposals handles GET /v1/chat/proposals — caller's frozen proposals (Phase 30 WS1).
func (h *Handler) ListProposals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.q == nil {
		httputil.WriteError(w, http.StatusInternalServerError, "database unavailable")
		return
	}
	userID, hasUser := authctx.UserID(r.Context())
	if !hasUser {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	ctx := r.Context()
	_ = h.q.ExpireStaleGuardianProposals(ctx)

	var farmIDPtr *int64
	if raw := strings.TrimSpace(r.URL.Query().Get("farm_id")); raw != "" {
		farmID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || farmID <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid farm_id")
			return
		}
		if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
			return
		}
		farmIDPtr = &farmID
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "pending"
	}
	switch status {
	case "pending", "expired", "confirmed", "dismissed", "superseded":
	default:
		httputil.WriteError(w, http.StatusBadRequest, "invalid status")
		return
	}
	statusPtr := &status

	limit := int32(defaultProposalListLimit)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 1 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		if n > maxProposalListLimit {
			n = maxProposalListLimit
		}
		limit = int32(n)
	}
	offset := int32(0)
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = int32(n)
	}

	listParams := db.ListGuardianProposalsByUserParams{
		UserID: userID,
		FarmID: farmIDPtr,
		Status: statusPtr,
		Limit:  limit,
		Offset: offset,
	}
	rows, err := h.q.ListGuardianProposalsByUser(ctx, listParams)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	total, err := h.q.CountGuardianProposalsByUser(ctx, db.CountGuardianProposalsByUserParams{
		UserID: userID,
		FarmID: farmIDPtr,
		Status: statusPtr,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]proposalListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, rowToProposalListItem(row))
	}
	httputil.WriteJSON(w, http.StatusOK, proposalListResponse{
		Proposals: items,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	})
}

func rowToProposalListItem(row db.Gr33ncoreGuardianActionProposal) proposalListItem {
	ap := farmguardian.ActionProposalFromRow(row)
	return proposalListItem{
		ProposalID:           ap.ProposalID,
		Tool:                 ap.Tool,
		Args:                 ap.Args,
		Summary:              ap.Summary,
		RiskTier:             ap.RiskTier,
		ExpiresAt:            ap.ExpiresAt,
		CreatedAt:            row.CreatedAt,
		FarmID:               row.FarmID,
		Status:               row.Status,
		Revision:             ap.Revision,
		SupersedesProposalID: ap.SupersedesProposalID,
		OperatorProvided:     ap.OperatorProvided,
		ImpactSummary:        ap.ImpactSummary,
	}
}
