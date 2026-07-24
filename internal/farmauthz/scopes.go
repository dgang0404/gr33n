package farmauthz

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

// Stable farm-scoped capability ids (Phase 211.03 WS1).
const (
	ScopeFarmMember       = "farm.member"
	ScopeFarmAdmin        = "farm.admin"
	ScopeFarmOperate      = "farm.operate"
	ScopeMoneyCostsRead   = "money.costs.read"
	ScopeMoneyCostsWrite  = "money.costs.write"
	ScopeNFRead           = "nf.read"
	ScopeNFInputsWrite    = "nf.inputs.write"
	ScopeNFInputsDelete   = "nf.inputs.delete"
	ScopeNFBatchesWrite   = "nf.batches.write"
	ScopeNFBatchesDelete  = "nf.batches.delete"
	ScopeNFRecipesWrite   = "nf.recipes.write"
	ScopeNFRecipesDelete  = "nf.recipes.delete"
	ScopeNFPackApply      = "nf.pack.apply"
)

// AllFarmScopes is the catalog for Settings scope pickers and docs.
var AllFarmScopes = []string{
	ScopeFarmAdmin,
	ScopeFarmOperate,
	ScopeMoneyCostsRead,
	ScopeMoneyCostsWrite,
	ScopeNFRead,
	ScopeNFInputsWrite,
	ScopeNFInputsDelete,
	ScopeNFBatchesWrite,
	ScopeNFBatchesDelete,
	ScopeNFRecipesWrite,
	ScopeNFRecipesDelete,
	ScopeNFPackApply,
}

type permissionOverrides struct {
	Scopes []string `json:"scopes"`
	Deny   []string `json:"deny"`
}

func parsePermissionOverrides(raw json.RawMessage) permissionOverrides {
	if len(raw) == 0 {
		return permissionOverrides{}
	}
	trimmed := string(raw)
	if trimmed == "{}" || trimmed == "null" {
		return permissionOverrides{}
	}
	var o permissionOverrides
	if err := json.Unmarshal(raw, &o); err != nil {
		return permissionOverrides{}
	}
	return o
}

func allScopesSet() map[string]bool {
	out := map[string]bool{ScopeFarmMember: true}
	for _, s := range AllFarmScopes {
		out[s] = true
	}
	return out
}

func roleTemplateScopes(role commontypes.FarmMemberRoleEnum) map[string]bool {
	switch role {
	case commontypes.FarmMemberOwner, commontypes.FarmMemberManager:
		return allScopesSet()
	case commontypes.FarmMemberOperator, commontypes.FarmMemberWorker, commontypes.FarmMemberAgronomist:
		return map[string]bool{
			ScopeFarmMember:      true,
			ScopeFarmOperate:     true,
			ScopeNFRead:          true,
			ScopeNFInputsWrite:   true,
			ScopeNFBatchesWrite:  true,
			ScopeNFRecipesWrite:  true,
		}
	case commontypes.FarmMemberFinance:
		return map[string]bool{
			ScopeFarmMember:      true,
			ScopeMoneyCostsRead:  true,
			ScopeMoneyCostsWrite: true,
			ScopeNFRead:          true,
			ScopeNFBatchesWrite:  true,
		}
	case commontypes.FarmMemberViewer:
		return map[string]bool{
			ScopeFarmMember: true,
			ScopeNFRead:     true,
		}
	case commontypes.FarmMemberCustomRole:
		return map[string]bool{ScopeFarmMember: true}
	default:
		return map[string]bool{ScopeFarmMember: true}
	}
}

func mergeRoleScopes(role commontypes.FarmMemberRoleEnum, overrides permissionOverrides) map[string]bool {
	if role == commontypes.FarmMemberOwner || role == commontypes.FarmMemberManager {
		return allScopesSet()
	}
	scopes := roleTemplateScopes(role)
	for _, s := range overrides.Scopes {
		if s != "" {
			scopes[s] = true
		}
	}
	for _, s := range overrides.Deny {
		delete(scopes, s)
	}
	return scopes
}

func scopesToLegacyCaps(scopes map[string]bool) FarmCaps {
	return FarmCaps{
		ViewCosts: scopes[ScopeMoneyCostsRead],
		EditCosts: scopes[ScopeMoneyCostsWrite],
		Operate:   scopes[ScopeFarmOperate],
		Admin:     scopes[ScopeFarmAdmin],
	}
}

func scopeList(scopes map[string]bool) []string {
	out := make([]string, 0, len(scopes))
	for s, ok := range scopes {
		if ok && s != ScopeFarmMember {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// ResolveFarmScopes merges role template + permissions JSONB overrides.
func ResolveFarmScopes(ctx context.Context, q db.Querier, userID uuid.UUID, farmID int64) (map[string]bool, commontypes.FarmMemberRoleEnum, error) {
	if authctx.FarmAuthzSkip(ctx) {
		return allScopesSet(), commontypes.FarmMemberOwner, nil
	}
	farm, err := q.GetFarmByID(ctx, farmID)
	if err != nil {
		return nil, "", err
	}
	if farm.OwnerUserID == userID {
		return allScopesSet(), commontypes.FarmMemberOwner, nil
	}
	m, err := q.GetFarmMembership(ctx, db.GetFarmMembershipParams{FarmID: farmID, UserID: userID})
	if err != nil {
		return nil, "", err
	}
	return mergeRoleScopes(m.RoleInFarm, parsePermissionOverrides(m.Permissions)), m.RoleInFarm, nil
}

// FarmScopesForUser resolves scopes without writing HTTP errors.
func FarmScopesForUser(ctx context.Context, q db.Querier, userID uuid.UUID, farmID int64) (map[string]bool, commontypes.FarmMemberRoleEnum, error) {
	return ResolveFarmScopes(ctx, q, userID, farmID)
}

// HasFarmScope reports whether the user holds a scope on the farm.
func HasFarmScope(ctx context.Context, q db.Querier, userID uuid.UUID, farmID int64, scope string) (bool, error) {
	scopes, _, err := ResolveFarmScopes(ctx, q, userID, farmID)
	if err != nil {
		return false, err
	}
	return scopes[scope], nil
}

// RequireFarmScope enforces a single scope on the request context user.
func RequireFarmScope(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64, scope, denied string) bool {
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
	var scopes map[string]bool
	if farm.OwnerUserID == uid {
		scopes = allScopesSet()
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
		scopes = mergeRoleScopes(m.RoleInFarm, parsePermissionOverrides(m.Permissions))
	}
	if !scopes[scope] {
		httputil.WriteError(w, http.StatusForbidden, denied)
		return false
	}
	return true
}

// RequireFarmScopes enforces that the user holds every listed scope.
func RequireFarmScopes(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64, denied string, need ...string) bool {
	for _, scope := range need {
		if !RequireFarmScope(w, r, q, farmID, scope, denied) {
			return false
		}
	}
	return true
}

// MeCapsResponse is returned by GET /farms/{id}/me/caps.
type MeCapsResponse struct {
	RoleInFarm string   `json:"role_in_farm"`
	Scopes     []string `json:"scopes"`
}

// MeCapsForUser builds the caps API payload.
func MeCapsForUser(ctx context.Context, q db.Querier, userID uuid.UUID, farmID int64) (MeCapsResponse, error) {
	scopes, role, err := ResolveFarmScopes(ctx, q, userID, farmID)
	if err != nil {
		return MeCapsResponse{}, err
	}
	return MeCapsResponse{
		RoleInFarm: string(role),
		Scopes:     scopeList(scopes),
	}, nil
}
