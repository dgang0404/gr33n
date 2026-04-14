package fertigation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	q *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

func numericFromFloat64(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64))
	return n, err
}

func farmIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// GET /farms/{id}/fertigation/reservoirs
func (h *Handler) ListReservoirsByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
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
	rows, err := h.q.ListFertigationEventsByFarm(r.Context(), farmID)
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

	var req struct {
		ProgramID           *int64             `json:"program_id"`
		ReservoirID         *int64             `json:"reservoir_id"`
		ZoneID              int64              `json:"zone_id"`
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
