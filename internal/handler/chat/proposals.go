package chat

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

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
	SessionID  string         `json:"session_id,omitempty"`

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

	limit, offset, err := httputil.ParseLimitOffsetStrict(r, defaultProposalListLimit, maxProposalListLimit)
	if err != nil {
		if errors.Is(err, httputil.ErrInvalidLimit) {
			httputil.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		httputil.WriteError(w, http.StatusBadRequest, "invalid offset")
		return
	}

	listParams := db.ListGuardianProposalsByUserParams{
		UserID: userID,
		FarmID: farmIDPtr,
		Status: statusPtr,
		Limit:  int32(limit),
		Offset: int32(offset),
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
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
}

// PostDismissProposal handles POST /v1/chat/proposals/{id}/dismiss (Phase 73 WS3).
func (h *Handler) PostDismissProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
	rawID := strings.TrimSpace(r.PathValue("id"))
	proposalID, err := uuid.Parse(rawID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid proposal id")
		return
	}
	ctx := r.Context()
	_ = h.q.ExpireStaleGuardianProposals(ctx)
	row, err := h.q.DismissGuardianProposal(ctx, db.DismissGuardianProposalParams{
		ProposalID: proposalID,
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "proposal not found or not dismissible")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, row.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rowToProposalListItem(row))
}

type suggestEmptyZoneBody struct {
	FarmID  int64  `json:"farm_id"`
	ZoneID  int64  `json:"zone_id"`
	CropKey string `json:"crop_key,omitempty"`
}

// PostSuggestEmptyZoneProposal handles POST /v1/chat/proposals/suggest-empty-zone (Phase 73 WS2).
func (h *Handler) PostSuggestEmptyZoneProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
	var body suggestEmptyZoneBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.FarmID <= 0 || body.ZoneID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "farm_id and zone_id required")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, body.FarmID) {
		return
	}
	ctx := r.Context()
	_ = h.q.ExpireStaleGuardianProposals(ctx)
	row, err := farmguardian.InsertEmptyZoneSetupProposal(ctx, h.q, userID, body.FarmID, body.ZoneID, body.CropKey)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "already has") || strings.Contains(msg, "already on farm") {
			httputil.WriteError(w, http.StatusConflict, msg)
			return
		}
		httputil.WriteError(w, http.StatusBadRequest, msg)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, rowToProposalListItem(row))
}

func pgUUIDString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	id, err := uuid.FromBytes(u.Bytes[:])
	if err != nil {
		return ""
	}
	return id.String()
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
		SessionID:            pgUUIDString(row.SessionID),
		Revision:             ap.Revision,
		SupersedesProposalID: ap.SupersedesProposalID,
		OperatorProvided:     ap.OperatorProvided,
		ImpactSummary:        ap.ImpactSummary,
	}
}
