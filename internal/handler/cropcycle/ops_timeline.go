// Phase 211.02 WS5 — crop cycle ops timeline API.

package cropcycle

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/cropcycle/opstimeline"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// OpsTimeline — GET /farms/{id}/crop-cycles/{cid}/ops-timeline?from=&to=
func (h *Handler) OpsTimeline(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	cycleID, err := strconv.ParseInt(r.PathValue("cid"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}

	cycle, err := h.q.GetCropCycleByID(r.Context(), cycleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cycle.FarmID != farmID {
		httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
		return
	}

	from, to, err := parseOpsTimelineRange(r, cycle)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if to.Before(from) {
		httputil.WriteError(w, http.StatusBadRequest, "to must be on or after from")
		return
	}

	out, err := opstimeline.Build(r.Context(), h.q, cycle, from, to)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func parseOpsTimelineRange(r *http.Request, cycle db.Gr33nfertigationCropCycle) (time.Time, time.Time, error) {
	from, to := opstimeline.DefaultRange(cycle, time.Now().UTC())
	q := r.URL.Query()
	if raw := q.Get("from"); raw != "" {
		t, err := opstimeline.ParseTimeQuery(raw)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("invalid from (use RFC3339 or YYYY-MM-DD)")
		}
		from = t
	}
	if raw := q.Get("to"); raw != "" {
		t, err := opstimeline.ParseTimeQuery(raw)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("invalid to (use RFC3339 or YYYY-MM-DD)")
		}
		to = t
		if len(raw) == len("2006-01-02") {
			to = to.Add(24*time.Hour - time.Nanosecond)
		}
	}
	return from, to, nil
}
