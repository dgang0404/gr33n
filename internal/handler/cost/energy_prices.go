package cost

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

// ---------------------------------------------------------------------------
// Phase 20.95 WS2 — farm_energy_prices CRUD.
// Phase 20.7 WS4 will later multiply actuator runtime * watts * price_per_kwh
// into cost_transactions. This phase just stores the price history.
// ---------------------------------------------------------------------------

type energyPriceBody struct {
	EffectiveFrom string   `json:"effective_from"`
	EffectiveTo   *string  `json:"effective_to"`
	PricePerKWh   float64  `json:"price_per_kwh"`
	Currency      string   `json:"currency"`
	Notes         *string  `json:"notes"`
}

// ListEnergyPrices — GET /farms/{id}/energy-prices
func (h *Handler) ListEnergyPrices(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListFarmEnergyPrices(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreFarmEnergyPrice{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// CreateEnergyPrice — POST /farms/{id}/energy-prices
func (h *Handler) CreateEnergyPrice(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, farmID) {
		return
	}
	var body energyPriceBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	from, to, price, currency, err := parseEnergyPrice(&body)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	row, err := h.q.CreateFarmEnergyPrice(r.Context(), db.CreateFarmEnergyPriceParams{
		FarmID:        farmID,
		EffectiveFrom: from,
		EffectiveTo:   to,
		PricePerKwh:   price,
		Currency:      currency,
		Notes:         body.Notes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// UpdateEnergyPrice — PUT /energy-prices/{id}
func (h *Handler) UpdateEnergyPrice(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid energy price id")
		return
	}
	existing, err := h.q.GetFarmEnergyPriceByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "energy price not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, existing.FarmID) {
		return
	}
	var body energyPriceBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	from, to, price, currency, err := parseEnergyPrice(&body)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	row, err := h.q.UpdateFarmEnergyPrice(r.Context(), db.UpdateFarmEnergyPriceParams{
		ID:            id,
		EffectiveFrom: from,
		EffectiveTo:   to,
		PricePerKwh:   price,
		Currency:      currency,
		Notes:         body.Notes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// DeleteEnergyPrice — DELETE /energy-prices/{id}
func (h *Handler) DeleteEnergyPrice(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid energy price id")
		return
	}
	existing, err := h.q.GetFarmEnergyPriceByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "energy price not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, existing.FarmID) {
		return
	}
	if err := h.q.DeleteFarmEnergyPrice(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseEnergyPrice(body *energyPriceBody) (pgtype.Date, pgtype.Date, pgtype.Numeric, string, error) {
	var from, to pgtype.Date
	f, err := time.Parse("2006-01-02", strings.TrimSpace(body.EffectiveFrom))
	if err != nil {
		return from, to, pgtype.Numeric{}, "", errors.New("invalid effective_from (YYYY-MM-DD)")
	}
	from = pgtype.Date{Time: f, Valid: true}
	if body.EffectiveTo != nil && strings.TrimSpace(*body.EffectiveTo) != "" {
		tt, err := time.Parse("2006-01-02", strings.TrimSpace(*body.EffectiveTo))
		if err != nil {
			return from, to, pgtype.Numeric{}, "", errors.New("invalid effective_to (YYYY-MM-DD)")
		}
		to = pgtype.Date{Time: tt, Valid: true}
	}
	if body.PricePerKWh < 0 {
		return from, to, pgtype.Numeric{}, "", errors.New("price_per_kwh must be >= 0")
	}
	var price pgtype.Numeric
	if err := price.Scan(strconv.FormatFloat(body.PricePerKWh, 'f', -1, 64)); err != nil {
		return from, to, pgtype.Numeric{}, "", errors.New("invalid price_per_kwh")
	}
	cur := strings.ToUpper(strings.TrimSpace(body.Currency))
	if len(cur) != 3 {
		return from, to, pgtype.Numeric{}, "", errors.New("currency must be ISO 4217 (3 uppercase letters)")
	}
	return from, to, price, cur, nil
}
