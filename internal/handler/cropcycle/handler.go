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
	"gr33n-api/internal/fertigation/programfit"
	"gr33n-api/internal/httputil"
)

type Handler struct{ q *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
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
	cropKeyFilter := strings.TrimSpace(r.URL.Query().Get("crop_key"))
	out := h.enrichCyclesJSON(r.Context(), farmID, rows)
	if cropKeyFilter != "" {
		filtered := make([]map[string]any, 0, len(out))
		for _, m := range out {
			if ck, _ := m["crop_key"].(string); ck == cropKeyFilter {
				filtered = append(filtered, m)
			}
		}
		out = filtered
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) enrichCyclesJSON(ctx context.Context, farmID int64, rows []db.Gr33nfertigationCropCycle) []map[string]any {
	plantByID := map[int64]db.Gr33ncropsPlant{}
	if plants, err := h.q.ListPlantsByFarm(ctx, farmID); err == nil {
		for _, p := range plants {
			plantByID[p.ID] = p
		}
	}
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		m := cropcycle.CycleJSON(row)
		if row.PlantID != nil {
			if p, ok := plantByID[*row.PlantID]; ok {
				id := cropcycle.ResolveCycleCropIdentity(row, &p)
				if id.CropKey != nil {
					m["crop_key"] = *id.CropKey
				}
				if id.CatalogDisplayName != nil {
					m["catalog_display_name"] = *id.CatalogDisplayName
				}
			}
		}
		out[i] = m
	}
	return out
}

func (h *Handler) writeCycleJSON(ctx context.Context, w http.ResponseWriter, status int, row db.Gr33nfertigationCropCycle, warnings []string) {
	out := cropcycle.CycleJSON(row)
	if row.PlantID != nil {
		if plant, err := h.q.GetPlant(ctx, *row.PlantID); err == nil {
			id := cropcycle.ResolveCycleCropIdentity(row, &plant)
			if id.CropKey != nil {
				out["crop_key"] = *id.CropKey
			}
			if id.CatalogDisplayName != nil {
				out["catalog_display_name"] = *id.CatalogDisplayName
			}
		}
	}
	if len(warnings) > 0 {
		out["program_fit_warnings"] = warnings
	}
	httputil.WriteJSON(w, status, out)
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
	h.writeCycleJSON(r.Context(), w, http.StatusOK, row, nil)
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
	started, err := httputil.ParseDate(body.StartedAt)
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
	var fitWarnings []string
	if body.PrimaryProgramID != nil && *body.PrimaryProgramID > 0 {
		fitWarnings, err = h.programFitWarnings(r.Context(), body.PrimaryProgramID, plantID, body.CurrentStage)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "program not found")
			return
		}
		if programfit.StrictMode() && len(fitWarnings) > 0 {
			httputil.WriteError(w, http.StatusUnprocessableEntity, strings.Join(fitWarnings, "; "))
			return
		}
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
	h.writeCycleJSON(r.Context(), w, http.StatusCreated, row, fitWarnings)
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
			harvested, err = httputil.ParseDate(*body.HarvestedAt)
			if err != nil {
				httputil.WriteError(w, http.StatusBadRequest, "invalid harvested_at")
				return
			}
		}
	}
	yield := existing.YieldGrams
	if body.YieldGrams != nil {
		yield, err = httputil.NumericFromFloat64(*body.YieldGrams)
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
	var fitWarnings []string
	if body.PrimaryProgramID != nil && *body.PrimaryProgramID > 0 {
		stageStr := ""
		if existing.CurrentStage != nil {
			stageStr = string(*existing.CurrentStage)
		}
		fitWarnings, err = h.programFitWarnings(r.Context(), body.PrimaryProgramID, plantID, stageStr)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "program not found")
			return
		}
		if programfit.StrictMode() && len(fitWarnings) > 0 {
			httputil.WriteError(w, http.StatusUnprocessableEntity, strings.Join(fitWarnings, "; "))
			return
		}
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
	h.writeCycleJSON(r.Context(), w, http.StatusOK, row, fitWarnings)
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
	h.writeCycleJSON(r.Context(), w, http.StatusOK, row, nil)
}

func (h *Handler) programFitWarnings(ctx context.Context, programID *int64, plantID *int64, stage string) ([]string, error) {
	if programID == nil || *programID <= 0 {
		return nil, nil
	}
	cropKey := ""
	if plantID != nil && *plantID > 0 {
		p, err := h.q.GetPlant(ctx, *plantID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New("plant not found")
			}
			return nil, err
		}
		if p.CropKey != nil {
			cropKey = *p.CropKey
		}
	}
	return programfit.ValidateProgramForGrow(ctx, h.q, *programID, cropKey, stage)
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
