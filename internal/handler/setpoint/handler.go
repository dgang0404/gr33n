// Package setpoint owns the CRUD handlers for gr33ncore.zone_setpoints
// (Phase 20.6). Setpoints express the "ideal environment" for a zone or
// crop cycle at a given growth stage as first-class data; the rule
// engine reads them at eval time (see internal/automation/predicates.go)
// so one rule "dew_point out of ideal" auto-adjusts as cycles advance
// through stages.
package setpoint

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool, q: db.New(pool)}
}

// setpointBody is the shared shape for create + update. Scope
// (zone_id, crop_cycle_id) and numeric fields are all nullable so
// operators can represent "zone-wide default for every stage" (both
// stage and crop_cycle_id NULL) or a cycle override (crop_cycle_id +
// stage set). `meta` is an opaque passthrough — the rule engine
// doesn't read it; it's a place for UI notes, provenance, etc.
type setpointBody struct {
	ZoneID      *int64          `json:"zone_id"`
	CropCycleID *int64          `json:"crop_cycle_id"`
	Stage       *string         `json:"stage"`
	SensorType  string          `json:"sensor_type"`
	MinValue    *float64        `json:"min_value"`
	MaxValue    *float64        `json:"max_value"`
	IdealValue  *float64        `json:"ideal_value"`
	Meta        json.RawMessage `json:"meta,omitempty"`
}

// validateAndResolveScope mirrors the chk_setpoint_scope and
// chk_setpoint_numeric_coherent CHECK constraints client-side so
// operators get a readable 400 instead of a generic 500 from Postgres.
// It also enforces the cross-farm rule: any zone_id or crop_cycle_id
// supplied must belong to `farmID` (same pattern as the Phase 19 WS4
// schedule-precondition write-path).
func (h *Handler) validateAndResolveScope(r *http.Request, farmID int64, body *setpointBody) error {
	body.SensorType = strings.TrimSpace(body.SensorType)
	if body.SensorType == "" {
		return errors.New("sensor_type is required")
	}
	if body.ZoneID == nil && body.CropCycleID == nil {
		return errors.New("zone_id or crop_cycle_id (or both) must be set")
	}
	if body.Stage != nil {
		s := strings.TrimSpace(*body.Stage)
		if s == "" {
			body.Stage = nil
		} else {
			body.Stage = &s
		}
	}
	if body.MinValue != nil && body.MaxValue != nil && *body.MinValue > *body.MaxValue {
		return errors.New("min_value must be <= max_value")
	}
	if body.IdealValue != nil {
		if body.MinValue != nil && *body.IdealValue < *body.MinValue {
			return errors.New("ideal_value must be >= min_value")
		}
		if body.MaxValue != nil && *body.IdealValue > *body.MaxValue {
			return errors.New("ideal_value must be <= max_value")
		}
	}
	if body.ZoneID != nil {
		z, err := h.q.GetZoneByID(r.Context(), *body.ZoneID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("zone_id not found")
			}
			return err
		}
		if z.FarmID != farmID {
			return errors.New("zone_id does not belong to this farm")
		}
	}
	if body.CropCycleID != nil {
		cc, err := h.q.GetCropCycleByID(r.Context(), *body.CropCycleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("crop_cycle_id not found")
			}
			return err
		}
		if cc.FarmID != farmID {
			return errors.New("crop_cycle_id does not belong to this farm")
		}
	}
	return nil
}

func numericFromOptFloat(p *float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if p == nil {
		return n, nil
	}
	if err := n.Scan(strconv.FormatFloat(*p, 'f', -1, 64)); err != nil {
		return n, err
	}
	return n, nil
}

// List — GET /farms/{id}/setpoints
// Query params: zone_id, crop_cycle_id, sensor_type. Empty/absent = no filter.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	q := r.URL.Query()
	params := db.ListSetpointsByFarmFilteredParams{FarmID: farmID}
	if s := strings.TrimSpace(q.Get("zone_id")); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid zone_id")
			return
		}
		params.ZoneID = &v
	}
	if s := strings.TrimSpace(q.Get("crop_cycle_id")); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid crop_cycle_id")
			return
		}
		params.CropCycleID = &v
	}
	if s := strings.TrimSpace(q.Get("sensor_type")); s != "" {
		params.SensorType = &s
	}
	rows, err := h.q.ListSetpointsByFarmFiltered(r.Context(), params)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreZoneSetpoint{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Create — POST /farms/{id}/setpoints
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body setpointBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.validateAndResolveScope(r, farmID, &body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	minV, err := numericFromOptFloat(body.MinValue)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid min_value")
		return
	}
	maxV, err := numericFromOptFloat(body.MaxValue)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid max_value")
		return
	}
	idealV, err := numericFromOptFloat(body.IdealValue)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ideal_value")
		return
	}
	var metaBytes []byte
	if len(body.Meta) > 0 {
		metaBytes = []byte(body.Meta)
	}
	row, err := h.q.CreateSetpoint(r.Context(), db.CreateSetpointParams{
		FarmID:      farmID,
		ZoneID:      body.ZoneID,
		CropCycleID: body.CropCycleID,
		Stage:       body.Stage,
		SensorType:  body.SensorType,
		MinValue:    minV,
		MaxValue:    maxV,
		IdealValue:  idealV,
		Meta:        metaBytes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// Get — GET /setpoints/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid setpoint id")
		return
	}
	row, err := h.q.GetSetpointByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "setpoint not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, row.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Update — PUT /setpoints/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid setpoint id")
		return
	}
	existing, err := h.q.GetSetpointByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "setpoint not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	var body setpointBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.validateAndResolveScope(r, existing.FarmID, &body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	minV, err := numericFromOptFloat(body.MinValue)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid min_value")
		return
	}
	maxV, err := numericFromOptFloat(body.MaxValue)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid max_value")
		return
	}
	idealV, err := numericFromOptFloat(body.IdealValue)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ideal_value")
		return
	}
	var metaBytes []byte
	if len(body.Meta) > 0 {
		metaBytes = []byte(body.Meta)
	}
	row, err := h.q.UpdateSetpoint(r.Context(), db.UpdateSetpointParams{
		ID:          id,
		ZoneID:      body.ZoneID,
		CropCycleID: body.CropCycleID,
		Stage:       body.Stage,
		SensorType:  body.SensorType,
		MinValue:    minV,
		MaxValue:    maxV,
		IdealValue:  idealV,
		Meta:        metaBytes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Delete — DELETE /setpoints/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid setpoint id")
		return
	}
	existing, err := h.q.GetSetpointByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "setpoint not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	if err := h.q.DeleteSetpoint(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
