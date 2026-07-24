package farmauthz

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

// FarmCaps describes what a user may do on a single farm (after membership or ownership).
type FarmCaps struct {
	ViewCosts bool
	EditCosts bool
	Operate   bool // field / production mutations (tasks, zones, sensors, recipes, etc.)
	Admin     bool // farm record, membership management
}

func fullCaps() FarmCaps {
	return FarmCaps{ViewCosts: true, EditCosts: true, Operate: true, Admin: true}
}

func capsForRole(r commontypes.FarmMemberRoleEnum) FarmCaps {
	return scopesToLegacyCaps(roleTemplateScopes(r))
}

func capsForMembership(role commontypes.FarmMemberRoleEnum, permissions json.RawMessage) FarmCaps {
	return scopesToLegacyCaps(mergeRoleScopes(role, parsePermissionOverrides(permissions)))
}

// RequireFarmCaps runs check on resolved caps; handles farm missing vs not a member.
func RequireFarmCaps(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64, check func(FarmCaps) bool, denied string) bool {
	ctx := r.Context()
	if authctx.FarmAuthzSkip(ctx) {
		return true
	}
	uid, ok := authctx.UserID(ctx)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	farm, err := q.GetFarmByID(ctx, farmID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "farm not found")
			return false
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load farm")
		return false
	}
	var caps FarmCaps
	if farm.OwnerUserID == uid {
		caps = fullCaps()
	} else {
		m, err := q.GetFarmMembership(ctx, db.GetFarmMembershipParams{FarmID: farmID, UserID: uid})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusForbidden, "not a member of this farm")
				return false
			}
			httputil.WriteError(w, http.StatusInternalServerError, "failed to verify farm membership")
			return false
		}
		caps = capsForMembership(m.RoleInFarm, m.Permissions)
	}
	if !check(caps) {
		httputil.WriteError(w, http.StatusForbidden, denied)
		return false
	}
	return true
}

func RequireCostRead(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	return RequireFarmScope(w, r, q, farmID, ScopeMoneyCostsRead, "insufficient role to view costs")
}

func RequireCostWrite(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	return RequireFarmScope(w, r, q, farmID, ScopeMoneyCostsWrite, "insufficient role to edit costs")
}

// RequireFarmOperate is deprecated in favor of RequireFarmScope(..., ScopeFarmOperate).
func RequireFarmOperate(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	return RequireFarmScope(w, r, q, farmID, ScopeFarmOperate, "insufficient role to modify farm operations")
}

func RequireFarmAdmin(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	return RequireFarmScope(w, r, q, farmID, ScopeFarmAdmin, "insufficient role for farm administration")
}

// FarmCapsForUser resolves capabilities for a user on a farm without writing HTTP errors.
func FarmCapsForUser(ctx context.Context, q db.Querier, userID uuid.UUID, farmID int64) (FarmCaps, error) {
	if authctx.FarmAuthzSkip(ctx) {
		return fullCaps(), nil
	}
	farm, err := q.GetFarmByID(ctx, farmID)
	if err != nil {
		return FarmCaps{}, err
	}
	if farm.OwnerUserID == userID {
		return fullCaps(), nil
	}
	m, err := q.GetFarmMembership(ctx, db.GetFarmMembershipParams{FarmID: farmID, UserID: userID})
	if err != nil {
		return FarmCaps{}, err
	}
	return capsForMembership(m.RoleInFarm, m.Permissions), nil
}
