package organization

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool, q: db.New(pool)}
}

// Create — POST /organizations
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var body struct {
		Name          string `json:"name"`
		PlanTier      string `json:"plan_tier"`
		BillingStatus string `json:"billing_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name required")
		return
	}
	plan := strings.TrimSpace(body.PlanTier)
	if plan == "" {
		plan = "pilot"
	}
	billing := strings.TrimSpace(body.BillingStatus)
	if billing == "" {
		billing = "none"
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback(ctx)

	qtx := h.q.WithTx(tx)
	org, err := qtx.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name:          name,
		PlanTier:      plan,
		BillingStatus: billing,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create organization")
		return
	}
	if _, err := qtx.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
		OrganizationID: org.ID,
		UserID:         uid,
		RoleInOrg:      "owner",
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to add organization owner")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to commit")
		return
	}
	mod := "gr33ncore"
	tbl := "organizations"
	rid := strconv.FormatInt(org.ID, 10)
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         0,
		Action:         db.Gr33ncoreUserActionTypeEnumChangeSetting,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details: map[string]any{
			"kind": "organization_created",
			"name": org.Name,
		},
	})
	httputil.WriteJSON(w, http.StatusCreated, org)
}

// ListMine — GET /organizations
func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	rows, err := h.q.ListOrganizationsForUser(ctx, uid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list organizations")
		return
	}
	if rows == nil {
		rows = []db.ListOrganizationsForUserRow{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Get — GET /organizations/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	if !farmauthz.RequireOrgMember(w, r, h.q, orgID) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	org, err := h.q.GetOrganizationByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "organization not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load organization")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, org)
}

// Update — PATCH /organizations/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	if !farmauthz.RequireOrgAdmin(w, r, h.q, orgID) {
		return
	}
	var body struct {
		Name          *string `json:"name"`
		PlanTier      *string `json:"plan_tier"`
		BillingStatus *string `json:"billing_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	existing, err := h.q.GetOrganizationByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "organization not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load organization")
		return
	}
	name := existing.Name
	if body.Name != nil {
		t := strings.TrimSpace(*body.Name)
		if t == "" {
			httputil.WriteError(w, http.StatusBadRequest, "name cannot be empty")
			return
		}
		name = t
	}
	plan := existing.PlanTier
	if body.PlanTier != nil {
		t := strings.TrimSpace(*body.PlanTier)
		if t != "" {
			plan = t
		}
	}
	billing := existing.BillingStatus
	if body.BillingStatus != nil {
		t := strings.TrimSpace(*body.BillingStatus)
		if t != "" {
			billing = t
		}
	}
	org, err := h.q.UpdateOrganization(ctx, db.UpdateOrganizationParams{
		ID:            orgID,
		Name:          name,
		PlanTier:      plan,
		BillingStatus: billing,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update organization")
		return
	}
	mod := "gr33ncore"
	tbl := "organizations"
	rid := strconv.FormatInt(orgID, 10)
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         0,
		Action:         db.Gr33ncoreUserActionTypeEnumChangeSetting,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details:        map[string]any{"kind": "organization_updated"},
	})
	httputil.WriteJSON(w, http.StatusOK, org)
}

// UsageSummary — GET /organizations/{id}/usage-summary
func (h *Handler) UsageSummary(w http.ResponseWriter, r *http.Request) {
	orgID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	if !farmauthz.RequireOrgMember(w, r, h.q, orgID) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	oid := orgID
	summary, err := h.q.GetOrganizationUsageSummary(ctx, &oid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load usage summary")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, summary)
}

// AddMember — POST /organizations/{id}/members
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	orgID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	if !farmauthz.RequireOrgAdmin(w, r, h.q, orgID) {
		return
	}
	var body struct {
		Email     string `json:"email"`
		RoleInOrg string `json:"role_in_org"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	if email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email required")
		return
	}
	role := strings.TrimSpace(body.RoleInOrg)
	if role == "" {
		role = "member"
	}
	if role != "member" && role != "admin" {
		httputil.WriteError(w, http.StatusBadRequest, "role_in_org must be member or admin")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	prof, err := h.q.GetProfileByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "no user with this email")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to look up user")
		return
	}
	if _, err := h.q.GetOrganizationMembership(ctx, db.GetOrganizationMembershipParams{
		OrganizationID: orgID,
		UserID:         prof.UserID,
	}); err == nil {
		httputil.WriteError(w, http.StatusConflict, "user is already a member")
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to check membership")
		return
	}
	m, err := h.q.CreateOrganizationMembership(ctx, db.CreateOrganizationMembershipParams{
		OrganizationID: orgID,
		UserID:         prof.UserID,
		RoleInOrg:      role,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to add member")
		return
	}
	mod := "gr33ncore"
	tbl := "organization_memberships"
	rid := strconv.FormatInt(orgID, 10) + ":" + prof.UserID.String()
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         0,
		Action:         db.Gr33ncoreUserActionTypeEnumChangeSetting,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details: map[string]any{
			"kind": "organization_member_added",
			"role": role,
		},
	})
	httputil.WriteJSON(w, http.StatusCreated, m)
}
