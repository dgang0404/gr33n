package fertigation

// Phase 39 WS6 — operator sets base water EC/pH on a reservoir so the mix
// calculator always has a starting point.
//
// Route: PATCH /fertigation/reservoirs/{rid}/base-water

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// PATCH /fertigation/reservoirs/{rid}/base-water
//
// Body:
//
//	{
//	  "ec_mscm": 0.2,   -- source water EC in mS/cm (required)
//	  "ph":      7.1    -- source water pH (optional; 0 = not set)
//	}
//
// Updates reservoir.last_ec_mscm, last_ph, last_reading_time.
// The mix calculator will no longer return ErrReservoirBaseECUnknown after this call.
func (h *Handler) SetReservoirBaseWater(w http.ResponseWriter, r *http.Request) {
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
		EcMscm float64 `json:"ec_mscm"`
		Ph     float64 `json:"ph"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.EcMscm <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "ec_mscm must be a positive value (e.g. 0.2 for RO water)")
		return
	}

	ecNum, err := httputil.NumericFromFloat64(body.EcMscm)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ec_mscm value")
		return
	}
	var phNum pgtype.Numeric
	if body.Ph > 0 {
		if phNum, err = httputil.NumericFromFloat64(body.Ph); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid ph value")
			return
		}
	}

	updated, err := h.q.UpdateReservoirBaseWater(ctx, db.UpdateReservoirBaseWaterParams{
		ID:         id,
		LastEcMscm: ecNum,
		LastPh:     phNum,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "reservoir not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update reservoir base water")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"reservoir": updated,
		"message":   "Base water EC set — mix calculator will use this value for future mix plans.",
	})
}
