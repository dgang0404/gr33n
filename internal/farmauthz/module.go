package farmauthz

import (
	"net/http"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

// RequireFarmModule returns false after 403 when the module is disabled for the farm.
func RequireFarmModule(w http.ResponseWriter, r *http.Request, q db.Querier, farmID int64, moduleSchema string) bool {
	if authctx.FarmAuthzSkip(r.Context()) {
		return true
	}
	enabled, err := q.FarmModuleIsEnabled(r.Context(), db.FarmModuleIsEnabledParams{
		FarmID:           farmID,
		ModuleSchemaName: moduleSchema,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to verify farm module")
		return false
	}
	if !enabled {
		httputil.WriteError(w, http.StatusForbidden, "module disabled for this farm")
		return false
	}
	return true
}
