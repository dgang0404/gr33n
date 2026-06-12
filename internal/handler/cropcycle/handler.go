package cropcycle

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/cropcycle"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct{ q *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

func numericFromFloat64(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64))
	return n, err
}

func parseGrowthStage(s string) *db.Gr33nfertigationGrowthStageEnum {
	var v db.Gr33nfertigationGrowthStageEnum
	if strings.TrimSpace(s) == "" {
		v = db.Gr33nfertigationGrowthStageEnumSeedling
	} else {
		v = db.Gr33nfertigationGrowthStageEnum(strings.TrimSpace(s))
	}
	return &v
}

func parseDate(s string) (pgtype.Date, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Date{}, errors.New("empty date")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// List — GET /farms/{id}/crop-cycles
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListCropCyclesByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nfertigationCropCycle{}
	}
	if plantRaw := strings.TrimSpace(r.URL.Query().Get("plant_id")); plantRaw != "" {
		plantID, err := strconv.ParseInt(plantRaw, 10, 64)
		if err != nil || plantID <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid plant_id")
			return
		}
		filtered := make([]db.Gr33nfertigationCropCycle, 0, len(rows))
		for _, row := range rows {
			if row.PlantID != nil && *row.PlantID == plantID {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}
	writeCyclesJSON(w, http.StatusOK, rows)
}

func writeCyclesJSON(w http.ResponseWriter, status int, rows []db.Gr33nfertigationCropCycle) {
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		out[i] = cropcycle.CycleJSON(row)
	}
	httputil.WriteJSON(w, status, out)
}

func writeCycleJSON(w http.ResponseWriter, status int, row db.Gr33nfertigationCropCycle) {
	httputil.WriteJSON(w, status, cropcycle.CycleJSON(row))
}

// Get — GET /crop-cycles/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	row, err := h.q.GetCropCycleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, row.FarmID) {
		return
	}
	writeCycleJSON(w, http.StatusOK, row)
}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body struct {
		ZoneID           int64   `json:"zone_id"`
		Name             string  `json:"name"`
		StrainOrVariety  *string `json:"strain_or_variety"`
		BatchLabel       *string `json:"batch_label"`
		CurrentStage     string  `json:"current_stage"`
		IsActive         *bool   `json:"is_active"`
		StartedAt        string  `json:"started_at"`
		CycleNotes       *string `json:"cycle_notes"`
		PrimaryProgramID *int64  `json:"primary_program_id"`
		PlantID          *int64  `json:"plant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || body.ZoneID == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "name and zone_id required")
		return
	}
	z, err := h.q.GetZoneByID(r.Context(), body.ZoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "zone not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if z.FarmID != farmID {
		httputil.WriteError(w, http.StatusBadRequest, "zone does not belong to this farm")
		return
	}
	started, err := parseDate(body.StartedAt)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid started_at (use YYYY-MM-DD)")
		return
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	batchLabel := cropcycle.ResolveBatchLabel(body.BatchLabel, body.StrainOrVariety)
	plantID, err := h.resolvePlantIDForFarm(r.Context(), farmID, body.PlantID, active)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	stage := parseGrowthStage(body.CurrentStage)
	row, err := h.q.CreateCropCycle(r.Context(), db.CreateCropCycleParams{
		FarmID:           farmID,
		ZoneID:           body.ZoneID,
		Name:             name,
		BatchLabel:       batchLabel,
		CurrentStage:     stage,
		IsActive:         active,
		StartedAt:        started,
		CycleNotes:       body.CycleNotes,
		PrimaryProgramID: body.PrimaryProgramID,
		PlantID:          plantID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httputil.WriteError(w, http.StatusConflict, "only one active crop cycle per zone is allowed")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if stage != nil {
		enteredAt := stageEnteredAt(started, row.CreatedAt)
		if _, err := h.q.InsertCropCycleStageEvent(r.Context(), db.InsertCropCycleStageEventParams{
			CropCycleID: row.ID,
			GrowthStage: *stage,
			EnteredAt:   enteredAt,
		}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeCycleJSON(w, http.StatusCreated, row)
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	var body struct {
		Name             string   `json:"name"`
		BatchLabel       *string  `json:"batch_label"`
		StrainOrVariety  *string  `json:"strain_or_variety"`
		ZoneID           int64    `json:"zone_id"`
		IsActive         bool     `json:"is_active"`
		CycleNotes       *string  `json:"cycle_notes"`
		HarvestedAt      *string  `json:"harvested_at"`
		YieldGrams       *float64 `json:"yield_grams"`
		YieldNotes       *string  `json:"yield_notes"`
		PrimaryProgramID *int64   `json:"primary_program_id"`
		PlantID          *int64   `json:"plant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || body.ZoneID == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "name and zone_id required")
		return
	}
	existing, err := h.q.GetCropCycleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	z, err := h.q.GetZoneByID(r.Context(), body.ZoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "zone not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if z.FarmID != existing.FarmID {
		httputil.WriteError(w, http.StatusBadRequest, "zone does not belong to this farm")
		return
	}
	harvested := existing.HarvestedAt
	if body.HarvestedAt != nil {
		if strings.TrimSpace(*body.HarvestedAt) == "" {
			harvested = pgtype.Date{}
		} else {
			harvested, err = parseDate(*body.HarvestedAt)
			if err != nil {
				httputil.WriteError(w, http.StatusBadRequest, "invalid harvested_at")
				return
			}
		}
	}
	yield := existing.YieldGrams
	if body.YieldGrams != nil {
		yield, err = numericFromFloat64(*body.YieldGrams)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid yield_grams")
			return
		}
	}
	plantID := existing.PlantID
	if body.PlantID != nil {
		plantID, err = h.resolvePlantIDForFarm(r.Context(), existing.FarmID, body.PlantID, body.IsActive)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	} else if body.IsActive {
		if plantID == nil || *plantID <= 0 {
			httputil.WriteError(w, http.StatusBadRequest, "plant_id required for active crop cycle — pick a catalog plant in Zone → Plants or Start grow")
			return
		}
		if err := cropcycle.ValidatePlantForActiveGrow(r.Context(), h.q, existing.FarmID, *plantID); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	batchLabel := existing.BatchLabel
	if body.BatchLabel != nil || body.StrainOrVariety != nil {
		batchLabel = cropcycle.ResolveBatchLabel(body.BatchLabel, body.StrainOrVariety)
	}
	row, err := h.q.UpdateCropCycle(r.Context(), db.UpdateCropCycleParams{
		ID:               id,
		Name:             name,
		BatchLabel:       batchLabel,
		ZoneID:           body.ZoneID,
		IsActive:         body.IsActive,
		CycleNotes:       body.CycleNotes,
		HarvestedAt:      harvested,
		YieldGrams:       yield,
		YieldNotes:       body.YieldNotes,
		PrimaryProgramID: body.PrimaryProgramID,
		PlantID:          plantID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httputil.WriteError(w, http.StatusConflict, "only one active crop cycle per zone is allowed")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeCycleJSON(w, http.StatusOK, row)
}

// UpdateStage — PATCH /crop-cycles/{id}/stage
func (h *Handler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	var body struct {
		CurrentStage string `json:"current_stage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(body.CurrentStage) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "current_stage required")
		return
	}
	cc, err := h.q.GetCropCycleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, cc.FarmID) {
		return
	}
	newStage := parseGrowthStage(body.CurrentStage)
	row, err := h.q.UpdateCropCycleStage(r.Context(), db.UpdateCropCycleStageParams{
		ID:           id,
		CurrentStage: newStage,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if newStage != nil && stageChanged(cc.CurrentStage, newStage) {
		if _, err := h.q.InsertCropCycleStageEvent(r.Context(), db.InsertCropCycleStageEventParams{
			CropCycleID: id,
			GrowthStage: *newStage,
			EnteredAt:   time.Now().UTC(),
		}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeCycleJSON(w, http.StatusOK, row)
}

func (h *Handler) resolvePlantIDForFarm(ctx context.Context, farmID int64, plantID *int64, active bool) (*int64, error) {
	if active {
		if plantID == nil || *plantID <= 0 {
			return nil, errors.New("plant_id required for active crop cycle — pick a catalog plant in Zone → Plants or Start grow")
		}
		if err := cropcycle.ValidatePlantForActiveGrow(ctx, h.q, farmID, *plantID); err != nil {
			return nil, err
		}
		return plantID, nil
	}
	if plantID == nil || *plantID <= 0 {
		return nil, nil
	}
	p, err := h.q.GetPlant(ctx, *plantID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("plant_id not found")
		}
		return nil, err
	}
	if p.FarmID != farmID {
		return nil, errors.New("plant_id does not belong to this farm")
	}
	return plantID, nil
}

func stageEnteredAt(started pgtype.Date, fallback time.Time) time.Time {
	if started.Valid {
		return started.Time.UTC()
	}
	return fallback.UTC()
}

func stageChanged(prev, next *db.Gr33nfertigationGrowthStageEnum) bool {
	if next == nil {
		return false
	}
	if prev == nil {
		return true
	}
	return *prev != *next
}

// Delete — DELETE /crop-cycles/{id} (soft: is_active = false)
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	cc, err := h.q.GetCropCycleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, cc.FarmID) {
		return
	}
	if err := h.q.SoftDeleteCropCycle(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
