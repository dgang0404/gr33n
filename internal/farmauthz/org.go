package farmauthz

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

func orgRoleIsAdmin(role string) bool {
	return role == "owner" || role == "admin"
}

// RequireOrgMember ensures the user belongs to the organization.
func RequireOrgMember(w http.ResponseWriter, r *http.Request, q db.Querier, orgID int64) bool {
	ctx := r.Context()
	if authctx.FarmAuthzSkip(ctx) {
		return true
	}
	uid, ok := authctx.UserID(ctx)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	_, err := q.GetOrganizationMembership(ctx, db.GetOrganizationMembershipParams{
		OrganizationID: orgID,
		UserID:         uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
			return false
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to verify organization membership")
		return false
	}
	return true
}

// RequireOrgAdmin ensures the user is an organization owner or admin.
func RequireOrgAdmin(w http.ResponseWriter, r *http.Request, q db.Querier, orgID int64) bool {
	ctx := r.Context()
	if authctx.FarmAuthzSkip(ctx) {
		return true
	}
	uid, ok := authctx.UserID(ctx)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	m, err := q.GetOrganizationMembership(ctx, db.GetOrganizationMembershipParams{
		OrganizationID: orgID,
		UserID:         uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
			return false
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to verify organization membership")
		return false
	}
	if !orgRoleIsAdmin(m.RoleInOrg) {
		httputil.WriteError(w, http.StatusForbidden, "organization admin role required")
		return false
	}
	return true
}

// UserCanAdminOrg reports whether the user is an org owner/admin (no HTTP write).
func UserCanAdminOrg(ctx context.Context, q db.Querier, orgID int64, uid uuid.UUID) (bool, error) {
	m, err := q.GetOrganizationMembership(ctx, db.GetOrganizationMembershipParams{
		OrganizationID: orgID,
		UserID:         uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return orgRoleIsAdmin(m.RoleInOrg), nil
}
