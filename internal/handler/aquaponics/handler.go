// Package aquaponics implements Phase 20.8 WS2 CRUD for
// gr33naquaponics.loops. A loop is the fish_tank ↔ grow_bed
// coupling; the actual pumps, sensors, and dosers live on the
// zones themselves (core primitives).
package aquaponics

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct{ q *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// ListLoops — GET /farms/{id}/aquaponics-loops
func (h *Handler) ListLoops(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListAquaponicsLoopsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33naquaponicsLoop{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// GetLoop — GET /aquaponics-loops/{id}
func (h *Handler) GetLoop(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid loop id")
		return
	}
	row, err := h.q.GetAquaponicsLoopByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "loop not found")
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

type loopReq struct {
	Label          string          `json:"label"`
	FishTankZoneID *int64          `json:"fish_tank_zone_id"`
	GrowBedZoneID  *int64          `json:"grow_bed_zone_id"`
	Active         *bool           `json:"active"`
	Meta           json.RawMessage `json:"meta"`
}

// CreateLoop — POST /farms/{id}/aquaponics-loops
func (h *Handler) CreateLoop(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body loopReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	label := strings.TrimSpace(body.Label)
	if label == "" {
		httputil.WriteError(w, http.StatusBadRequest, "label required")
		return
	}
	if err := h.assertZonesInFarm(r, farmID, body.FishTankZoneID, body.GrowBedZoneID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	meta := validMetaOrNil(body.Meta, w)
	if meta == nil && len(body.Meta) > 0 {
		return
	}
	row, err := h.q.CreateAquaponicsLoop(r.Context(), db.CreateAquaponicsLoopParams{
		FarmID:         farmID,
		Label:          label,
		FishTankZoneID: body.FishTankZoneID,
		GrowBedZoneID:  body.GrowBedZoneID,
		Meta:           meta,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// UpdateLoop — PUT /aquaponics-loops/{id}
func (h *Handler) UpdateLoop(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid loop id")
		return
	}
	existing, err := h.q.GetAquaponicsLoopByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "loop not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	var body loopReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	label := strings.TrimSpace(body.Label)
	if label == "" {
		httputil.WriteError(w, http.StatusBadRequest, "label required")
		return
	}
	if err := h.assertZonesInFarm(r, existing.FarmID, body.FishTankZoneID, body.GrowBedZoneID); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	active := existing.Active
	if body.Active != nil {
		active = *body.Active
	}
	meta := validMetaOrNil(body.Meta, w)
	if meta == nil && len(body.Meta) > 0 {
		return
	}
	row, err := h.q.UpdateAquaponicsLoop(r.Context(), db.UpdateAquaponicsLoopParams{
		ID:             id,
		Label:          label,
		FishTankZoneID: body.FishTankZoneID,
		GrowBedZoneID:  body.GrowBedZoneID,
		Active:         active,
		Meta:           meta,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// DeleteLoop — DELETE /aquaponics-loops/{id}
func (h *Handler) DeleteLoop(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid loop id")
		return
	}
	existing, err := h.q.GetAquaponicsLoopByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "loop not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	if err := h.q.SoftDeleteAquaponicsLoop(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── helpers ────────────────────────────────────────────────────────────────

func validMetaOrNil(raw json.RawMessage, w http.ResponseWriter) []byte {
	if len(raw) == 0 {
		return nil
	}
	if !json.Valid(raw) {
		httputil.WriteError(w, http.StatusBadRequest, "meta must be valid JSON")
		return nil
	}
	return raw
}

func (h *Handler) assertZonesInFarm(r *http.Request, farmID int64, zoneIDs ...*int64) error {
	for _, zp := range zoneIDs {
		if zp == nil {
			continue
		}
		z, err := h.q.GetZoneByID(r.Context(), *zp)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("zone not found")
			}
			return err
		}
		if z.FarmID != farmID {
			return errors.New("zone does not belong to this farm")
		}
	}
	return nil
}
