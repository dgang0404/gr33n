package farmauthz

import (
	"net/http"

	"github.com/google/uuid"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

// RequireFarmMember returns false after writing an error response when the user
// is not a member (or owner) of the farm. Skipped when AUTH_MODE=dev auth bypass is active.
func RequireFarmMember(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64) bool {
	ctx := r.Context()
	if authctx.FarmAuthzSkip(ctx) {
		return true
	}
	uid, ok := authctx.UserID(ctx)
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return false
	}
	return requireFarmAccess(w, r, q, farmID, uid)
}

func requireFarmAccess(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64, uid uuid.UUID) bool {
	hasPtr, err := q.UserHasFarmAccess(r.Context(), db.UserHasFarmAccessParams{
		FarmID: farmID,
		UserID: uid,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to verify farm access")
		return false
	}
	if hasPtr == nil || !*hasPtr {
		httputil.WriteError(w, http.StatusForbidden, "not a member of this farm")
		return false
	}
	return true
}
