package chat

import (
	"net/http"
	"os"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/httputil"
)

type qaLatestResponse struct {
	Summary farmguardian.QARunSummary      `json:"summary"`
	Scores  []farmguardian.EvalQuestionScore `json:"scores,omitempty"`
}

// GetLatestQARun handles GET /v1/guardian/qa/latest — latest archived smoke/regression run.
func (h *Handler) GetLatestQARun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.cfg.Enabled {
		httputil.WriteError(w, http.StatusNotFound, "ai disabled")
		return
	}
	if _, ok := authctx.UserID(r.Context()); !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	sum, scores, err := farmguardian.LatestQARunSummary()
	if err != nil {
		if os.IsNotExist(err) {
			httputil.WriteError(w, http.StatusNotFound, "no qa runs archived yet")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load qa run")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, qaLatestResponse{Summary: sum, Scores: scores})
}
