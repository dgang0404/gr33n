package actuator

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/hardware"
	"gr33n-api/internal/httputil"
)

// PatchWiring — PATCH /actuators/{id}/wiring
func (h *Handler) PatchWiring(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid actuator id")
		return
	}
	existing, err := h.q.GetActuatorByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "actuator not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}

	var body struct {
		Wiring *hardware.Wiring `json:"wiring"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Wiring != nil {
		if err := hardware.ValidateActuatorWiring(body.Wiring); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if h.pool != nil {
			if err := hardware.CheckActuatorWiringConflict(r.Context(), h.pool, existing.FarmID, id, body.Wiring); err != nil {
				httputil.WriteError(w, http.StatusConflict, err.Error())
				return
			}
		}
	}
	cfg, err := hardware.MergeWiring(existing.Config, body.Wiring)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.q.UpdateActuatorConfig(r.Context(), id, cfg)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, wrapActuatorWithCommands(updated))
}
