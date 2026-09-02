package fertigation

// Phase 216 WS7 — wire (or clear) the reservoir's circulation_pump actuator.
// Used both as the WS2 dosing-session default and the WS7 idle-maintenance
// program's target, so the operator sets it once per reservoir.
//
// Route: PATCH /fertigation/reservoirs/{rid}/circulation-actuator

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// PATCH /fertigation/reservoirs/{rid}/circulation-actuator
//
// Body: { "actuator_id": 42 }  -- actuator_id: 0 or omitted clears the wiring.
func (h *Handler) SetReservoirCirculationActuator(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid reservoir id")
		return
	}

	ctx := r.Context()
	res, err := h.q.GetFertigationReservoirByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "reservoir not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load reservoir")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, res.FarmID) {
		return
	}

	var body struct {
		ActuatorID int64 `json:"actuator_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var actuatorID *int64
	if body.ActuatorID > 0 {
		actuator, err := h.q.GetActuatorByID(ctx, body.ActuatorID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "actuator not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, "failed to load actuator")
			return
		}
		if actuator.FarmID != res.FarmID {
			httputil.WriteError(w, http.StatusBadRequest, "actuator belongs to a different farm")
			return
		}
		if actuator.ActuatorType != "circulation_pump" {
			httputil.WriteError(w, http.StatusBadRequest, "actuator_type must be circulation_pump")
			return
		}
		actuatorID = &body.ActuatorID
	}

	updated, err := h.q.UpdateReservoirCirculationActuator(ctx, db.UpdateReservoirCirculationActuatorParams{
		ID:                    id,
		CirculationActuatorID: actuatorID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update reservoir")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"reservoir": updated,
	})
}
