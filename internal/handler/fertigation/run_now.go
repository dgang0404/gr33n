package fertigation

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// RunProgramNow handles POST /farms/{id}/fertigation/programs/{rid}/run-now (B1).
func (h *Handler) RunProgramNow(w http.ResponseWriter, r *http.Request) {
	if h.worker == nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "automation worker not configured")
		return
	}
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	programID, err := strconv.ParseInt(r.PathValue("rid"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid program id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	prog, err := h.q.GetFertigationProgramByID(r.Context(), programID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if prog.FarmID != farmID {
		httputil.WriteError(w, http.StatusNotFound, "program not found")
		return
	}

	userID, hasUser := authctx.UserID(r.Context())
	status, message, duplicate, runErr := h.worker.RunProgramNow(r.Context(), prog)
	if runErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, runErr.Error())
		return
	}
	if hasUser {
		slog.Info("fertigation program run-now",
			"farm_id", farmID,
			"program_id", programID,
			"user_id", userID,
			"status", status,
			"duplicate", duplicate,
		)
	}

	code := http.StatusAccepted
	if duplicate {
		code = http.StatusOK
	}
	httputil.WriteJSON(w, code, map[string]any{
		"program_id": programID,
		"status":     status,
		"message":    message,
		"duplicate":  duplicate,
	})
}
