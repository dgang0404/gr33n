package fertigation

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/costing"
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

func numericFromFloat64(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64))
	return n, err
}

func farmIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func resourceIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("rid"), 10, 64)
}

func mixingEventIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("mid"), 10, 64)
}

// PATCH /fertigation/reservoirs/{rid}
func (h *Handler) UpdateReservoir(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid reservoir id")
		return
	}
	res, err := h.q.GetFertigationReservoirByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "reservoir not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, res.FarmID) {
		return
	}
	var req struct {
		Name                string  `json:"name"`
		Description         *string `json:"description"`
		CapacityLiters      float64 `json:"capacity_liters"`
		CurrentVolumeLiters float64 `json:"current_volume_liters"`
		Status              string  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	capacity, err := numericFromFloat64(req.CapacityLiters)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid capacity_liters")
		return
	}
	current, err := numericFromFloat64(req.CurrentVolumeLiters)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid current_volume_liters")
		return
	}
	row, err := h.q.UpdateReservoir(r.Context(), db.UpdateReservoirParams{
		ID:                  id,
		Name:                req.Name,
		Description:         req.Description,
		CapacityLiters:      capacity,
		CurrentVolumeLiters: current,
		Status:              db.Gr33nfertigationReservoirStatusEnum(req.Status),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// DELETE /fertigation/reservoirs/{rid}
func (h *Handler) DeleteReservoir(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid reservoir id")
		return
	}
	res, err := h.q.GetFertigationReservoirByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "reservoir not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, res.FarmID) {
		return
	}
	if err := h.q.DeleteReservoir(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /fertigation/programs/{rid}
func (h *Handler) UpdateProgram(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid program id")
		return
	}
	prog, err := h.q.GetFertigationProgramByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, prog.FarmID) {
		return
	}
	var req struct {
		Name              string  `json:"name"`
		Description       *string `json:"description"`
		ReservoirID       *int64  `json:"reservoir_id"`
		TargetZoneID      *int64  `json:"target_zone_id"`
		EcTargetID        *int64  `json:"ec_target_id"`
		TotalVolumeLiters float64 `json:"total_volume_liters"`
		IsActive          bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	totalVol, err := numericFromFloat64(req.TotalVolumeLiters)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid total_volume_liters")
		return
	}
	row, err := h.q.UpdateProgram(r.Context(), db.UpdateProgramParams{
		ID:                id,
		Name:              req.Name,
		Description:       req.Description,
		ReservoirID:       req.ReservoirID,
		TargetZoneID:      req.TargetZoneID,
		EcTargetID:        req.EcTargetID,
		TotalVolumeLiters: totalVol,
		IsActive:          req.IsActive,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// DELETE /fertigation/programs/{rid}
func (h *Handler) DeleteProgram(w http.ResponseWriter, r *http.Request) {
	id, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid program id")
		return
	}
	prog, err := h.q.GetFertigationProgramByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, prog.FarmID) {
		return
	}
	if err := h.q.DeleteProgram(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /farms/{id}/fertigation/reservoirs
func (h *Handler) ListReservoirsByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListReservoirsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nfertigationReservoir{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// POST /farms/{id}/fertigation/reservoirs
func (h *Handler) CreateReservoir(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	var req struct {
		ZoneID              *int64  `json:"zone_id"`
		Name                string  `json:"name"`
		Description         *string `json:"description"`
		CapacityLiters      float64 `json:"capacity_liters"`
		CurrentVolumeLiters float64 `json:"current_volume_liters"`
		Status              string  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	capacity, err := numericFromFloat64(req.CapacityLiters)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid capacity_liters")
		return
	}
	current, err := numericFromFloat64(req.CurrentVolumeLiters)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid current_volume_liters")
		return
	}

	res, err := h.q.CreateReservoir(r.Context(), db.CreateReservoirParams{
		FarmID:              farmID,
		ZoneID:              req.ZoneID,
		Name:                req.Name,
		Description:         req.Description,
		CapacityLiters:      capacity,
		CurrentVolumeLiters: current,
		Status:              db.Gr33nfertigationReservoirStatusEnum(req.Status),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, res)
}

// GET /farms/{id}/fertigation/ec-targets
func (h *Handler) ListEcTargetsByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListEcTargetsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nfertigationEcTarget{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// POST /farms/{id}/fertigation/ec-targets
func (h *Handler) CreateEcTarget(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	var req struct {
		ZoneID      *int64  `json:"zone_id"`
		GrowthStage string  `json:"growth_stage"`
		EcMinMscm   float64 `json:"ec_min_mscm"`
		EcMaxMscm   float64 `json:"ec_max_mscm"`
		PhMin       float64 `json:"ph_min"`
		PhMax       float64 `json:"ph_max"`
		Notes       *string `json:"notes"`
		Rationale   *string `json:"rationale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ecMin, err := numericFromFloat64(req.EcMinMscm)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ec_min_mscm")
		return
	}
	ecMax, err := numericFromFloat64(req.EcMaxMscm)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ec_max_mscm")
		return
	}
	phMin, err := numericFromFloat64(req.PhMin)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ph_min")
		return
	}
	phMax, err := numericFromFloat64(req.PhMax)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ph_max")
		return
	}

	row, err := h.q.CreateEcTarget(r.Context(), db.CreateEcTargetParams{
		FarmID:      farmID,
		ZoneID:      req.ZoneID,
		GrowthStage: db.Gr33nfertigationGrowthStageEnum(req.GrowthStage),
		EcMinMscm:   ecMin,
		EcMaxMscm:   ecMax,
		PhMin:       phMin,
		PhMax:       phMax,
		Notes:       req.Notes,
		Rationale:   req.Rationale,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// GET /farms/{id}/fertigation/programs
func (h *Handler) ListProgramsByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListProgramsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nfertigationProgram{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// POST /farms/{id}/fertigation/programs
func (h *Handler) CreateProgram(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	var req struct {
		Name                string  `json:"name"`
		Description         *string `json:"description"`
		ApplicationRecipeID *int64  `json:"application_recipe_id"`
		ReservoirID         *int64  `json:"reservoir_id"`
		TargetZoneID        *int64  `json:"target_zone_id"`
		ScheduleID          *int64  `json:"schedule_id"`
		EcTargetID          *int64  `json:"ec_target_id"`
		TotalVolumeLiters   float64 `json:"total_volume_liters"`
		RunDurationSeconds  *int32  `json:"run_duration_seconds"`
		EcTriggerLow        float64 `json:"ec_trigger_low"`
		PhTriggerLow        float64 `json:"ph_trigger_low"`
		PhTriggerHigh       float64 `json:"ph_trigger_high"`
		IsActive            bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	totalVol, err := numericFromFloat64(req.TotalVolumeLiters)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid total_volume_liters")
		return
	}
	ecLow, err := numericFromFloat64(req.EcTriggerLow)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ec_trigger_low")
		return
	}
	phLow, err := numericFromFloat64(req.PhTriggerLow)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ph_trigger_low")
		return
	}
	phHigh, err := numericFromFloat64(req.PhTriggerHigh)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ph_trigger_high")
		return
	}

	row, err := h.q.CreateProgram(r.Context(), db.CreateProgramParams{
		FarmID:              farmID,
		Name:                req.Name,
		Description:         req.Description,
		ApplicationRecipeID: req.ApplicationRecipeID,
		ReservoirID:         req.ReservoirID,
		TargetZoneID:        req.TargetZoneID,
		ScheduleID:          req.ScheduleID,
		EcTargetID:          req.EcTargetID,
		TotalVolumeLiters:   totalVol,
		RunDurationSeconds:  req.RunDurationSeconds,
		EcTriggerLow:        ecLow,
		PhTriggerLow:        phLow,
		PhTriggerHigh:       phHigh,
		IsActive:            req.IsActive,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// GET /farms/{id}/fertigation/events
func (h *Handler) ListEventsByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	var rows []db.Gr33nfertigationFertigationEvent
	if v := r.URL.Query().Get("crop_cycle_id"); v != "" {
		ccID, err := strconv.ParseInt(v, 10, 64)
		if err != nil || ccID < 1 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid crop_cycle_id")
			return
		}
		cc := ccID
		rows, err = h.q.ListFertigationEventsByFarmAndCropCycle(r.Context(), db.ListFertigationEventsByFarmAndCropCycleParams{
			FarmID:      farmID,
			CropCycleID: &cc,
		})
	} else {
		rows, err = h.q.ListFertigationEventsByFarm(r.Context(), farmID)
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nfertigationFertigationEvent{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// POST /farms/{id}/fertigation/events
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	var req struct {
		ProgramID           *int64             `json:"program_id"`
		ReservoirID         *int64             `json:"reservoir_id"`
		ZoneID              int64              `json:"zone_id"`
		CropCycleID         *int64             `json:"crop_cycle_id"`
		AppliedAt           time.Time          `json:"applied_at"`
		GrowthStage         *string            `json:"growth_stage"`
		VolumeAppliedLiters float64            `json:"volume_applied_liters"`
		RunDurationSeconds  *int32             `json:"run_duration_seconds"`
		EcBeforeMscm        float64            `json:"ec_before_mscm"`
		EcAfterMscm         float64            `json:"ec_after_mscm"`
		PhBefore            float64            `json:"ph_before"`
		PhAfter             float64            `json:"ph_after"`
		TriggerSource       *string            `json:"trigger_source"`
		Notes               *string            `json:"notes"`
		Metadata            map[string]any     `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	z, err := h.q.GetZoneByID(r.Context(), req.ZoneID)
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

	if req.CropCycleID != nil {
		cc, err := h.q.GetCropCycleByID(r.Context(), *req.CropCycleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "crop cycle not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if cc.FarmID != farmID || cc.ZoneID != req.ZoneID {
			httputil.WriteError(w, http.StatusBadRequest, "crop cycle must belong to this farm and zone")
			return
		}
	}

	volumeApplied, err := numericFromFloat64(req.VolumeAppliedLiters)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid volume_applied_liters")
		return
	}
	ecBefore, err := numericFromFloat64(req.EcBeforeMscm)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ec_before_mscm")
		return
	}
	ecAfter, err := numericFromFloat64(req.EcAfterMscm)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ec_after_mscm")
		return
	}
	phBefore, err := numericFromFloat64(req.PhBefore)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ph_before")
		return
	}
	phAfter, err := numericFromFloat64(req.PhAfter)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid ph_after")
		return
	}

	var growthStage db.NullGr33nfertigationGrowthStageEnum
	if req.GrowthStage != nil {
		growthStage = db.NullGr33nfertigationGrowthStageEnum{
			Gr33nfertigationGrowthStageEnum: db.Gr33nfertigationGrowthStageEnum(*req.GrowthStage),
			Valid:                           true,
		}
	}
	var triggerSource db.NullGr33nfertigationProgramTriggerEnum
	if req.TriggerSource != nil {
		triggerSource = db.NullGr33nfertigationProgramTriggerEnum{
			Gr33nfertigationProgramTriggerEnum: db.Gr33nfertigationProgramTriggerEnum(*req.TriggerSource),
			Valid:                              true,
		}
	}

	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid metadata")
		return
	}

	appliedAt := req.AppliedAt
	if appliedAt.IsZero() {
		appliedAt = time.Now().UTC()
	}

	row, err := h.q.CreateFertigationEvent(r.Context(), db.CreateFertigationEventParams{
		FarmID:              farmID,
		ProgramID:           req.ProgramID,
		ReservoirID:         req.ReservoirID,
		ZoneID:              req.ZoneID,
		CropCycleID:         req.CropCycleID,
		AppliedAt:           appliedAt,
		GrowthStage:         growthStage,
		VolumeAppliedLiters: volumeApplied,
		RunDurationSeconds:  req.RunDurationSeconds,
		EcBeforeMscm:        ecBefore,
		EcAfterMscm:         ecAfter,
		PhBefore:            phBefore,
		PhAfter:             phAfter,
		TriggerSource:       triggerSource,
		Notes:               req.Notes,
		Metadata:            metadataJSON,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// POST /farms/{id}/fertigation/mixing-events
func (h *Handler) CreateMixingEvent(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	var req struct {
		ReservoirID       int64    `json:"reservoir_id"`
		ProgramID         *int64   `json:"program_id"`
		WaterVolumeLiters float64  `json:"water_volume_liters"`
		WaterSource       *string  `json:"water_source"`
		WaterEcMscm       *float64 `json:"water_ec_mscm"`
		WaterPh           *float64 `json:"water_ph"`
		FinalEcMscm       *float64 `json:"final_ec_mscm"`
		FinalPh           *float64 `json:"final_ph"`
		FinalTempCelsius  *float64 `json:"final_temp_celsius"`
		EcTargetID        *int64   `json:"ec_target_id"`
		EcTargetMet       *bool    `json:"ec_target_met"`
		Notes             *string  `json:"notes"`
		Observations      *string  `json:"observations"`
		Components        []struct {
			InputDefinitionID int64   `json:"input_definition_id"`
			InputBatchID      *int64  `json:"input_batch_id"`
			VolumeAddedMl     float64 `json:"volume_added_ml"`
			DilutionRatio     *string `json:"dilution_ratio"`
			Notes             *string `json:"notes"`
		} `json:"components"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ReservoirID < 1 || req.WaterVolumeLiters <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "reservoir_id and water_volume_liters are required")
		return
	}

	res, err := h.q.GetFertigationReservoirByID(r.Context(), req.ReservoirID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "reservoir not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if res.FarmID != farmID {
		httputil.WriteError(w, http.StatusBadRequest, "reservoir does not belong to this farm")
		return
	}

	waterVol, _ := numericFromFloat64(req.WaterVolumeLiters)
	var waterEc, waterPh, finalEc, finalPh, finalTemp pgtype.Numeric
	if req.WaterEcMscm != nil {
		waterEc, _ = numericFromFloat64(*req.WaterEcMscm)
	}
	if req.WaterPh != nil {
		waterPh, _ = numericFromFloat64(*req.WaterPh)
	}
	if req.FinalEcMscm != nil {
		finalEc, _ = numericFromFloat64(*req.FinalEcMscm)
	}
	if req.FinalPh != nil {
		finalPh, _ = numericFromFloat64(*req.FinalPh)
	}
	if req.FinalTempCelsius != nil {
		finalTemp, _ = numericFromFloat64(*req.FinalTempCelsius)
	}

	var mixedBy pgtype.UUID
	if uid, ok := authctx.UserID(r.Context()); ok {
		mixedBy = pgtype.UUID{Bytes: uid, Valid: true}
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(r.Context())
	qtx := h.q.WithTx(tx)

	event, err := qtx.CreateMixingEvent(r.Context(), db.CreateMixingEventParams{
		FarmID:            farmID,
		ReservoirID:       req.ReservoirID,
		ProgramID:         req.ProgramID,
		MixedByUserID:     mixedBy,
		MixedAt:           time.Now().UTC(),
		WaterVolumeLiters: waterVol,
		WaterSource:       req.WaterSource,
		WaterEcMscm:       waterEc,
		WaterPh:           waterPh,
		FinalEcMscm:       finalEc,
		FinalPh:           finalPh,
		FinalTempCelsius:  finalTemp,
		EcTargetID:        req.EcTargetID,
		EcTargetMet:       req.EcTargetMet,
		Notes:             req.Notes,
		Observations:      req.Observations,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create mixing event")
		return
	}

	components := make([]db.Gr33nfertigationMixingEventComponent, 0, len(req.Components))
	for _, c := range req.Components {
		vol, _ := numericFromFloat64(c.VolumeAddedMl)
		comp, err := qtx.CreateMixingEventComponent(r.Context(), db.CreateMixingEventComponentParams{
			MixingEventID:     event.ID,
			InputDefinitionID: c.InputDefinitionID,
			InputBatchID:      c.InputBatchID,
			VolumeAddedMl:     vol,
			DilutionRatio:     c.DilutionRatio,
			Notes:             c.Notes,
		})
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to create mixing component")
			return
		}
		// Phase 20.7 WS2: inside the transaction so the deduct +
		// cost row commit atomically with the component insert.
		// The autologger is idempotent on `mixing_component:<id>`
		// so a failed Commit → retry at the route layer (future
		// offline-sync) won't double-bill.
		if err := costing.LogMixingComponent(r.Context(), qtx, farmID, comp, event.MixedAt); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to log mixing component cost: "+err.Error())
			return
		}
		components = append(components, comp)
	}

	if err := tx.Commit(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"event":      event,
		"components": components,
	})
}

// GET /farms/{id}/fertigation/mixing-events
func (h *Handler) ListMixingEventsByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListMixingEventsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nfertigationMixingEvent{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// GET /farms/{id}/fertigation/mixing-events/{mid}/components
func (h *Handler) ListMixingEventComponents(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	mid, err := mixingEventIDFromPath(r)
	if err != nil || mid < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mixing event id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	ev, err := h.q.GetMixingEventByID(r.Context(), mid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "mixing event not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ev.FarmID != farmID {
		httputil.WriteError(w, http.StatusNotFound, "mixing event not found")
		return
	}
	rows, err := h.q.ListMixingEventComponents(r.Context(), mid)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nfertigationMixingEventComponent{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}
