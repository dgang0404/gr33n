package farm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmbootstrap"
	"gr33n-api/internal/httputil"
)

func int64PtrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

type Handler struct {
	q          db.Querier
	pool       *pgxpool.Pool
	httpClient *http.Client
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		q:    db.New(pool),
		pool: pool,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func NewHandlerWithQuerier(q db.Querier) *Handler {
	return &Handler{
		q:    q,
		pool: nil,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// GET /farms?user_id=<uuid>  (user_id is optional; omit to list all farms)
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		farms, err := h.q.ListAllFarms(ctx)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to list farms")
			return
		}
		httputil.WriteJSON(w, http.StatusOK, farms)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid user_id")
		return
	}
	farms, err := h.q.ListFarmsForUser(ctx, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list farms")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, farms)
}

// GET /farms/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, id) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	farm, err := h.q.GetFarmByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "farm not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, farm)
}

// POST /farms
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		db.CreateFarmParams
		BootstrapTemplate *string `json:"bootstrap_template"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	if req.OrganizationID != nil {
		uid, ok := authctx.UserID(r.Context())
		if !ok {
			httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		can, err := farmauthz.UserCanAdminOrg(ctx, h.q, *req.OrganizationID, uid)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to verify organization access")
			return
		}
		if !can {
			httputil.WriteError(w, http.StatusForbidden, "organization owner or admin required to link a new farm")
			return
		}
	}

	effectiveBootstrap := req.BootstrapTemplate
	if effectiveBootstrap == nil && req.OrganizationID != nil {
		orgRow, err := h.q.GetOrganizationByID(ctx, *req.OrganizationID)
		if err == nil && orgRow.DefaultBootstrapTemplate != nil {
			t := strings.TrimSpace(*orgRow.DefaultBootstrapTemplate)
			if t != "" {
				effectiveBootstrap = &t
			}
		}
	}

	farm, err := h.q.CreateFarm(ctx, req.CreateFarmParams)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create farm")
		return
	}

	tmplVal, tmplRequested := farmbootstrap.RequestedTemplate(effectiveBootstrap)
	var boot map[string]any
	if tmplRequested {
		if farmbootstrap.IsBlankChoice(tmplVal) {
			boot = map[string]any{"skipped": true}
		} else {
			uid, ok := authctx.UserID(r.Context())
			if !ok || uid != req.OwnerUserID {
				boot = map[string]any{"skipped": true, "reason": "creator_must_match_owner_user_id"}
			} else {
				boot, err = h.runFarmBootstrap(ctx, farm.ID, tmplVal)
				if err != nil {
					httputil.WriteError(w, http.StatusInternalServerError, "farm created but bootstrap failed: "+err.Error())
					return
				}
			}
		}
		httputil.WriteJSON(w, http.StatusCreated, map[string]any{"farm": farm, "bootstrap": boot})
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, farm)
}

// PUT /farms/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var params db.UpdateFarmParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	params.ID = id
	if !farmauthz.RequireFarmAdmin(w, r, h.q, id) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	existing, err := h.q.GetFarmByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "farm not found")
		return
	}
	if !int64PtrEqual(existing.OrganizationID, params.OrganizationID) {
		uid, ok := authctx.UserID(r.Context())
		if !ok {
			httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if params.OrganizationID != nil {
			can, err := farmauthz.UserCanAdminOrg(ctx, h.q, *params.OrganizationID, uid)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "failed to verify organization access")
				return
			}
			if !can {
				httputil.WriteError(w, http.StatusForbidden, "organization owner or admin required to link this farm")
				return
			}
		}
	}

	farm, err := h.q.UpdateFarm(ctx, params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update farm")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, farm)
}

// SetOrganization — PATCH /farms/{id}/organization
func (h *Handler) SetOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, id) {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	field, ok := raw["organization_id"]
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, "organization_id is required (null to unlink)")
		return
	}
	var orgID *int64
	if string(field) == "null" {
		orgID = nil
	} else {
		var n int64
		if err := json.Unmarshal(field, &n); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "organization_id must be an integer or null")
			return
		}
		orgID = &n
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	existing, err := h.q.GetFarmByID(ctx, id)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "farm not found")
		return
	}
	if !int64PtrEqual(existing.OrganizationID, orgID) {
		uid, ok := authctx.UserID(r.Context())
		if !ok {
			httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if orgID != nil {
			can, err := farmauthz.UserCanAdminOrg(ctx, h.q, *orgID, uid)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "failed to verify organization access")
				return
			}
			if !can {
				httputil.WriteError(w, http.StatusForbidden, "organization owner or admin required to link this farm")
				return
			}
		}
	}

	farm, err := h.q.SetFarmOrganization(ctx, db.SetFarmOrganizationParams{
		ID:             id,
		OrganizationID: orgID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update farm organization")
		return
	}
	if !int64PtrEqual(existing.OrganizationID, orgID) {
		mod := "gr33ncore"
		tbl := "farms"
		rid := strconv.FormatInt(id, 10)
		details := map[string]any{"kind": "farm_organization_changed"}
		if orgID != nil {
			details["organization_id"] = *orgID
		} else {
			details["organization_id"] = nil
		}
		if existing.OrganizationID != nil {
			details["previous_organization_id"] = *existing.OrganizationID
		} else {
			details["previous_organization_id"] = nil
		}
		auditlog.Submit(ctx, h.q, r, auditlog.Event{
			FarmID:         auditlog.FarmIDPtr(id),
			Action:         db.Gr33ncoreUserActionTypeEnumUpdateRecord,
			TargetSchema:   &mod,
			TargetTable:    &tbl,
			TargetRecordID: &rid,
			Details:        details,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, farm)
}

// DELETE /farms/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var body struct {
		UpdatedByUserID uuid.UUID `json:"updated_by_user_id"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if !farmauthz.RequireFarmAdmin(w, r, h.q, id) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.q.SoftDeleteFarm(ctx, db.SoftDeleteFarmParams{
		ID:              id,
		UpdatedByUserID: pgtype.UUID{Bytes: body.UpdatedByUserID, Valid: true},
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete farm")
		return
	}
	mod := "gr33ncore"
	tbl := "farms"
	rid := strconv.FormatInt(id, 10)
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(id),
		Action:         db.Gr33ncoreUserActionTypeEnumDeleteRecord,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details:        map[string]any{"kind": "farm_soft_deleted"},
	})
	w.WriteHeader(http.StatusNoContent)
}

// SetInsertCommonsOptIn — PATCH /farms/{id}/insert-commons/opt-in
func (h *Handler) SetInsertCommonsOptIn(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, id) {
		return
	}
	var body struct {
		InsertCommonsOptIn           bool  `json:"insert_commons_opt_in"`
		InsertCommonsRequireApproval *bool `json:"insert_commons_require_approval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	existing, err := h.q.GetFarmByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "farm not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load farm")
		return
	}
	reqApproval := existing.InsertCommonsRequireApproval
	if body.InsertCommonsRequireApproval != nil {
		reqApproval = *body.InsertCommonsRequireApproval
	}
	if !body.InsertCommonsOptIn {
		reqApproval = false
	}
	row, err := h.q.SetFarmInsertCommonsOptIn(ctx, db.SetFarmInsertCommonsOptInParams{
		ID:                           id,
		InsertCommonsOptIn:           body.InsertCommonsOptIn,
		InsertCommonsRequireApproval: reqApproval,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "farm not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update farm")
		return
	}
	mod := "gr33ncore"
	tbl := "farms"
	rid := strconv.FormatInt(id, 10)
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(id),
		Action:         db.Gr33ncoreUserActionTypeEnumChangeSetting,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details: map[string]any{
			"kind":                            "insert_commons_opt_in",
			"insert_commons_opt_in":           body.InsertCommonsOptIn,
			"insert_commons_require_approval": row.InsertCommonsRequireApproval,
		},
	})
	httputil.WriteJSON(w, http.StatusOK, row)
}
