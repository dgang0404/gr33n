package farmauthz

import (
	"errors"
	"net/http"

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
	switch r {
	case commontypes.FarmMemberOwner, commontypes.FarmMemberManager:
		return fullCaps()
	case commontypes.FarmMemberFinance:
		return FarmCaps{ViewCosts: true, EditCosts: true, Operate: false, Admin: false}
	case commontypes.FarmMemberOperator, commontypes.FarmMemberWorker, commontypes.FarmMemberAgronomist:
		return FarmCaps{ViewCosts: false, EditCosts: false, Operate: true, Admin: false}
	case commontypes.FarmMemberViewer, commontypes.FarmMemberCustomRole:
		return FarmCaps{ViewCosts: false, EditCosts: false, Operate: false, Admin: false}
	default:
		return FarmCaps{}
	}
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
		caps = capsForRole(m.RoleInFarm)
	}
	if !check(caps) {
		httputil.WriteError(w, http.StatusForbidden, denied)
		return false
	}
	return true
}

func RequireCostRead(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	return RequireFarmCaps(w, r, q, farmID, func(c FarmCaps) bool { return c.ViewCosts },
		"insufficient role to view costs")
}

func RequireCostWrite(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	return RequireFarmCaps(w, r, q, farmID, func(c FarmCaps) bool { return c.EditCosts },
		"insufficient role to edit costs")
}

func RequireFarmOperate(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	return RequireFarmCaps(w, r, q, farmID, func(c FarmCaps) bool { return c.Operate },
		"insufficient role to modify farm operations")
}

func RequireFarmAdmin(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	return RequireFarmCaps(w, r, q, farmID, func(c FarmCaps) bool { return c.Admin },
		"insufficient role for farm administration")
}
