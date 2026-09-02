package fertigation

// Phase 216 WS6 — dosing session HTTP surface for the Fertigation UI panel:
// start a session for a mixing_event, list active sessions for a farm, read
// one session's live state, and manually abort one.

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// POST /farms/{id}/fertigation/mixing-events/{mid}/dosing-session
//
//	Body: {
//	  "dose_increment_ml": 0.5, "max_total_dose_ml": 20,
//	  "circulate_seconds": 30, "settle_seconds": 30,
//	  "samples_per_cycle": 3, "max_no_improvement_cycles": 3
//	}
//
// Resolves the reservoir + its circulation_actuator_id from the mixing_event;
// the worker picks the new session up on its next tick (status starts 'idle').
func (h *Handler) CreateDosingSession(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	mixingEventID, err := mixingEventIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mixing_event id")
		return
	}

	ctx := r.Context()
	mixEvent, err := h.q.GetMixingEventByID(ctx, mixingEventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "mixing event not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load mixing event")
		return
	}
	if mixEvent.FarmID != farmID {
		httputil.WriteError(w, http.StatusBadRequest, "mixing event belongs to a different farm")
		return
	}
	res, err := h.q.GetFertigationReservoirByID(ctx, mixEvent.ReservoirID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load reservoir")
		return
	}

	var body struct {
		DoseIncrementMl        float64 `json:"dose_increment_ml"`
		MaxTotalDoseMl         float64 `json:"max_total_dose_ml"`
		CirculateSeconds       int32   `json:"circulate_seconds"`
		SettleSeconds          int32   `json:"settle_seconds"`
		SamplesPerCycle        int32   `json:"samples_per_cycle"`
		MaxNoImprovementCycles int32   `json:"max_no_improvement_cycles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.DoseIncrementMl <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "dose_increment_ml must be positive")
		return
	}
	if body.MaxTotalDoseMl <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "max_total_dose_ml must be positive")
		return
	}
	if body.CirculateSeconds <= 0 {
		body.CirculateSeconds = 30
	}
	if body.SettleSeconds <= 0 {
		body.SettleSeconds = 30
	}
	if body.SamplesPerCycle <= 0 {
		body.SamplesPerCycle = 3
	}
	if body.MaxNoImprovementCycles <= 0 {
		body.MaxNoImprovementCycles = 3
	}

	doseIncNumeric, err := httputil.NumericFromFloat64(body.DoseIncrementMl)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid dose_increment_ml")
		return
	}
	maxTotalNumeric, err := httputil.NumericFromFloat64(body.MaxTotalDoseMl)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid max_total_dose_ml")
		return
	}

	created, err := h.q.CreateDosingSession(ctx, db.CreateDosingSessionParams{
		MixingEventID:          mixingEventID,
		FarmID:                 farmID,
		ReservoirID:            res.ID,
		CirculationActuatorID:  res.CirculationActuatorID,
		DoseIncrementMl:        doseIncNumeric,
		MaxTotalDoseMl:         maxTotalNumeric,
		CirculateSeconds:       body.CirculateSeconds,
		SettleSeconds:          body.SettleSeconds,
		SamplesPerCycle:        body.SamplesPerCycle,
		MaxNoImprovementCycles: body.MaxNoImprovementCycles,
		NextActionAt:           pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create dosing session")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, created)
}

// GET /farms/{id}/fertigation/dosing-sessions
func (h *Handler) ListActiveDosingSessions(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	sessions, err := h.q.ListActiveDosingSessionsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list dosing sessions")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, sessions)
}

// GET /fertigation/dosing-sessions/{id}
func (h *Handler) GetDosingSession(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid dosing session id")
		return
	}
	session, err := h.q.GetDosingSessionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "dosing session not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load dosing session")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, session.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, session)
}

// POST /fertigation/dosing-sessions/{id}/abort
//
// Body: { "reason": "operator stopped the run" } — reason is optional.
func (h *Handler) AbortDosingSessionByOperator(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid dosing session id")
		return
	}
	session, err := h.q.GetDosingSessionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "dosing session not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load dosing session")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, session.FarmID) {
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	reason := body.Reason
	if reason == "" {
		reason = "aborted by operator"
	}

	updated, err := h.q.AbortDosingSession(r.Context(), db.AbortDosingSessionParams{
		ID: id, AbortedReason: &reason,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to abort dosing session")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, updated)
}
