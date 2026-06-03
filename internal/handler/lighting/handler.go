package lighting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

// Preset definitions — keyed by preset_key.
type presetDef struct {
	Name     string
	OnHours  int32
	OffHours int32
}

var presets = map[string]presetDef{
	"peas_22_2":      {Name: "Peas 22/2 (Long-day veg)", OnHours: 22, OffHours: 2},
	"veg_18_6":       {Name: "Veg 18/6 (Vegetative)", OnHours: 18, OffHours: 6},
	"flower_12_12":   {Name: "Flower 12/12 (Flowering)", OnHours: 12, OffHours: 12},
	"seedling_16_8":  {Name: "Seedling 16/8", OnHours: 16, OffHours: 8},
}

// PresetList returns the available presets for the UI.
func PresetList() []map[string]any {
	out := make([]map[string]any, 0, len(presets))
	for key, p := range presets {
		out = append(out, map[string]any{
			"key":       key,
			"name":      p.Name,
			"on_hours":  p.OnHours,
			"off_hours": p.OffHours,
		})
	}
	return out
}

type Handler struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool, q: db.New(pool)}
}

// ── helpers ─────────────────────────────────────────────────────────────────

func farmIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func programIDFromPath(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("pid"), 10, 64)
}

// parseHHMM parses a "HH:MM" string and returns the hour and minute.
func parseHHMM(s string) (hour, minute int, err error) {
	parts := strings.SplitN(strings.TrimSpace(s), ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected HH:MM format, got %q", s)
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("invalid time %q: hour must be 0-23, minute 0-59", s)
	}
	return h, m, nil
}

// buildCronExpressions returns (onCron, offCron) given a lights_on_at anchor and on_hours.
// Example: lights_on_at="06:00", on_hours=18 → "0 6 * * *", "0 0 * * *"
func buildCronExpressions(lightsOnAt string, onHours int32) (onCron, offCron string, err error) {
	h, m, err := parseHHMM(lightsOnAt)
	if err != nil {
		return "", "", err
	}
	onCron = fmt.Sprintf("%d %d * * *", m, h)

	offTotalMinutes := (h*60 + m + int(onHours)*60) % (24 * 60)
	offH := offTotalMinutes / 60
	offM := offTotalMinutes % 60
	offCron = fmt.Sprintf("%d %d * * *", offM, offH)
	return onCron, offCron, nil
}

// materializeSchedules creates (or replaces) the ON and OFF schedules for a
// lighting program within the provided transaction. Returns the schedule IDs.
func materializeSchedules(
	ctx context.Context,
	qtx *db.Queries,
	prog db.Gr33ncoreLightingProgram,
	onCron, offCron string,
) (onID, offID int64, err error) {
	onRow, err := qtx.CreateSchedule(ctx, db.CreateScheduleParams{
		FarmID:         prog.FarmID,
		Name:           fmt.Sprintf("LP-%d ON: %s", prog.ID, prog.Name),
		Description:    ptrStr(fmt.Sprintf("Auto-generated ON schedule for lighting program %d", prog.ID)),
		ScheduleType:   "lighting",
		CronExpression: onCron,
		Timezone:       prog.Timezone,
		IsActive:       prog.IsActive,
		MetaData:       []byte(fmt.Sprintf(`{"lighting_program_id":%d,"side":"on"}`, prog.ID)),
		Preconditions:  []byte("[]"),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("create ON schedule: %w", err)
	}
	offRow, err := qtx.CreateSchedule(ctx, db.CreateScheduleParams{
		FarmID:         prog.FarmID,
		Name:           fmt.Sprintf("LP-%d OFF: %s", prog.ID, prog.Name),
		Description:    ptrStr(fmt.Sprintf("Auto-generated OFF schedule for lighting program %d", prog.ID)),
		ScheduleType:   "lighting",
		CronExpression: offCron,
		Timezone:       prog.Timezone,
		IsActive:       prog.IsActive,
		MetaData:       []byte(fmt.Sprintf(`{"lighting_program_id":%d,"side":"off"}`, prog.ID)),
		Preconditions:  []byte("[]"),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("create OFF schedule: %w", err)
	}
	return onRow.ID, offRow.ID, nil
}

// createScheduleActions creates the control_actuator ON and OFF executable actions.
func createScheduleActions(
	ctx context.Context,
	qtx *db.Queries,
	actuatorID int64,
	onScheduleID, offScheduleID int64,
) error {
	actionType := commontypes.ExecutableActionTypeEnum("control_actuator")

	_, err := qtx.CreateExecutableActionForSchedule(ctx, db.CreateExecutableActionForScheduleParams{
		ScheduleID:     &onScheduleID,
		ExecutionOrder: 0,
		ActionType:     actionType,
		TargetActuatorID: &actuatorID,
		ActionCommand:  ptrStr("on"),
		ActionParameters: json.RawMessage(`{}`),
		DelayBeforeExecutionSeconds: ptrInt32(0),
	})
	if err != nil {
		return fmt.Errorf("create ON action: %w", err)
	}

	_, err = qtx.CreateExecutableActionForSchedule(ctx, db.CreateExecutableActionForScheduleParams{
		ScheduleID:     &offScheduleID,
		ExecutionOrder: 0,
		ActionType:     actionType,
		TargetActuatorID: &actuatorID,
		ActionCommand:  ptrStr("off"),
		ActionParameters: json.RawMessage(`{}`),
		DelayBeforeExecutionSeconds: ptrInt32(0),
	})
	if err != nil {
		return fmt.Errorf("create OFF action: %w", err)
	}
	return nil
}

func ptrStr(s string) *string   { return &s }
func ptrInt32(n int32) *int32   { return &n }
func ptrInt64(n int64) *int64   { return &n }

// ── preset list ──────────────────────────────────────────────────────────────

// GET /lighting-programs/presets
func (h *Handler) ListPresets(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, PresetList())
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

type createProgramRequest struct {
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	ZoneID      int64           `json:"zone_id"`
	ActuatorID  int64           `json:"actuator_id"`
	OnHours     int32           `json:"on_hours"`
	OffHours    int32           `json:"off_hours"`
	LightsOnAt  string          `json:"lights_on_at"`
	Timezone    string          `json:"timezone"`
	CropCycleID *int64          `json:"crop_cycle_id"`
	IsActive    bool            `json:"is_active"`
	Metadata    json.RawMessage `json:"metadata"`
}

func (req *createProgramRequest) validate() error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if req.ZoneID <= 0 {
		return fmt.Errorf("zone_id is required")
	}
	if req.ActuatorID <= 0 {
		return fmt.Errorf("actuator_id is required")
	}
	if req.OnHours <= 0 || req.OnHours > 24 {
		return fmt.Errorf("on_hours must be 1-24")
	}
	if req.OffHours < 0 || req.OffHours >= 24 {
		return fmt.Errorf("off_hours must be 0-23")
	}
	if req.OnHours+req.OffHours != 24 {
		return fmt.Errorf("on_hours + off_hours must equal 24")
	}
	if req.LightsOnAt == "" {
		req.LightsOnAt = "06:00"
	}
	if _, _, err := parseHHMM(req.LightsOnAt); err != nil {
		return err
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}
	if _, err := time.LoadLocation(req.Timezone); err != nil {
		return fmt.Errorf("invalid timezone %q", req.Timezone)
	}
	return nil
}

// POST /farms/{id}/lighting-programs
func (h *Handler) CreateProgram(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var req createProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.validate(); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Metadata == nil {
		req.Metadata = json.RawMessage(`{}`)
	}

	onCron, offCron, err := buildCronExpressions(req.LightsOnAt, req.OnHours)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck

	qtx := h.q.WithTx(tx)

	prog, err := qtx.CreateLightingProgram(r.Context(), db.CreateLightingProgramParams{
		FarmID:      farmID,
		ZoneID:      req.ZoneID,
		ActuatorID:  req.ActuatorID,
		Name:        req.Name,
		Description: req.Description,
		OnHours:     req.OnHours,
		OffHours:    req.OffHours,
		LightsOnAt:  req.LightsOnAt,
		Timezone:    req.Timezone,
		CropCycleID: req.CropCycleID,
		IsActive:    req.IsActive,
		Metadata:    req.Metadata,
		// Schedule IDs set after creation
		ScheduleOnID:  nil,
		ScheduleOffID: nil,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create lighting program")
		return
	}

	onID, offID, err := materializeSchedules(r.Context(), qtx, prog, onCron, offCron)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := createScheduleActions(r.Context(), qtx, req.ActuatorID, onID, offID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	prog, err = qtx.UpdateLightingProgramSchedules(r.Context(), db.UpdateLightingProgramSchedulesParams{
		ID:            prog.ID,
		ScheduleOnID:  &onID,
		ScheduleOffID: &offID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to link schedules")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to commit")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, prog)
}

// POST /farms/{id}/lighting-programs/from-preset
func (h *Handler) CreateFromPreset(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body struct {
		PresetKey   string  `json:"preset_key"`
		Name        *string `json:"name"`
		ZoneID      int64   `json:"zone_id"`
		ActuatorID  int64   `json:"actuator_id"`
		LightsOnAt  string  `json:"lights_on_at"`
		Timezone    string  `json:"timezone"`
		CropCycleID *int64  `json:"crop_cycle_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, ok := presets[body.PresetKey]
	if !ok {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("unknown preset_key %q; available: peas_22_2, veg_18_6, flower_12_12, seedling_16_8", body.PresetKey))
		return
	}
	name := p.Name
	if body.Name != nil && strings.TrimSpace(*body.Name) != "" {
		name = *body.Name
	}
	if body.LightsOnAt == "" {
		body.LightsOnAt = "06:00"
	}
	if body.Timezone == "" {
		// Try to inherit from farm.
		farm, ferr := h.q.GetFarmByID(r.Context(), farmID)
		if ferr == nil && farm.Timezone != "" {
			body.Timezone = farm.Timezone
		} else {
			body.Timezone = "UTC"
		}
	}
	if _, err := time.LoadLocation(body.Timezone); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid timezone %q", body.Timezone))
		return
	}

	meta, _ := json.Marshal(map[string]string{"preset_key": body.PresetKey})

	req := createProgramRequest{
		Name:        name,
		ZoneID:      body.ZoneID,
		ActuatorID:  body.ActuatorID,
		OnHours:     p.OnHours,
		OffHours:    p.OffHours,
		LightsOnAt:  body.LightsOnAt,
		Timezone:    body.Timezone,
		CropCycleID: body.CropCycleID,
		IsActive:    true,
		Metadata:    json.RawMessage(meta),
	}
	if err := req.validate(); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	onCron, offCron, err := buildCronExpressions(req.LightsOnAt, req.OnHours)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck

	qtx := h.q.WithTx(tx)

	prog, err := qtx.CreateLightingProgram(r.Context(), db.CreateLightingProgramParams{
		FarmID:        farmID,
		ZoneID:        req.ZoneID,
		ActuatorID:    req.ActuatorID,
		Name:          req.Name,
		Description:   req.Description,
		OnHours:       req.OnHours,
		OffHours:      req.OffHours,
		LightsOnAt:    req.LightsOnAt,
		Timezone:      req.Timezone,
		CropCycleID:   req.CropCycleID,
		IsActive:      req.IsActive,
		Metadata:      req.Metadata,
		ScheduleOnID:  nil,
		ScheduleOffID: nil,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create lighting program")
		return
	}

	onID, offID, err := materializeSchedules(r.Context(), qtx, prog, onCron, offCron)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := createScheduleActions(r.Context(), qtx, req.ActuatorID, onID, offID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	prog, err = qtx.UpdateLightingProgramSchedules(r.Context(), db.UpdateLightingProgramSchedulesParams{
		ID:            prog.ID,
		ScheduleOnID:  &onID,
		ScheduleOffID: &offID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to link schedules")
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to commit")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, prog)
}

// GET /farms/{id}/lighting-programs
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListLightingProgramsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list lighting programs")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreLightingProgram{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// GET /lighting-programs/{pid}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := programIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid lighting program id")
		return
	}
	prog, err := h.q.GetLightingProgramByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "lighting program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load lighting program")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, prog.FarmID) {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, prog)
}

// PATCH /lighting-programs/{pid}
// Accepts a subset of fields; regenerates crons if photoperiod params change.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := programIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid lighting program id")
		return
	}
	prog, err := h.q.GetLightingProgramByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "lighting program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load lighting program")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, prog.FarmID) {
		return
	}

	var body struct {
		Name        *string         `json:"name"`
		Description *string         `json:"description"`
		OnHours     *int32          `json:"on_hours"`
		OffHours    *int32          `json:"off_hours"`
		LightsOnAt  *string         `json:"lights_on_at"`
		Timezone    *string         `json:"timezone"`
		CropCycleID *int64          `json:"crop_cycle_id"`
		IsActive    *bool           `json:"is_active"`
		Metadata    json.RawMessage `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Merge with existing values.
	if body.Name != nil {
		prog.Name = *body.Name
	}
	if body.Description != nil {
		prog.Description = body.Description
	}
	if body.OnHours != nil {
		prog.OnHours = *body.OnHours
	}
	if body.OffHours != nil {
		prog.OffHours = *body.OffHours
	}
	if body.LightsOnAt != nil {
		prog.LightsOnAt = *body.LightsOnAt
	}
	if body.Timezone != nil {
		prog.Timezone = *body.Timezone
	}
	if body.CropCycleID != nil {
		prog.CropCycleID = body.CropCycleID
	}
	if body.IsActive != nil {
		prog.IsActive = *body.IsActive
	}
	if body.Metadata != nil {
		prog.Metadata = body.Metadata
	}

	if prog.OnHours+prog.OffHours != 24 {
		httputil.WriteError(w, http.StatusBadRequest, "on_hours + off_hours must equal 24")
		return
	}
	if _, _, err := parseHHMM(prog.LightsOnAt); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := time.LoadLocation(prog.Timezone); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid timezone %q", prog.Timezone))
		return
	}

	onCron, offCron, err := buildCronExpressions(prog.LightsOnAt, prog.OnHours)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck

	qtx := h.q.WithTx(tx)

	// Delete old schedules (cascade deletes their actions).
	if prog.ScheduleOnID != nil {
		_ = qtx.DeleteSchedule(r.Context(), *prog.ScheduleOnID)
	}
	if prog.ScheduleOffID != nil {
		_ = qtx.DeleteSchedule(r.Context(), *prog.ScheduleOffID)
	}

	onID, offID, err := materializeSchedules(r.Context(), qtx, prog, onCron, offCron)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := createScheduleActions(r.Context(), qtx, prog.ActuatorID, onID, offID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, err := qtx.UpdateLightingProgram(r.Context(), db.UpdateLightingProgramParams{
		ID:            prog.ID,
		Name:          prog.Name,
		Description:   prog.Description,
		OnHours:       prog.OnHours,
		OffHours:      prog.OffHours,
		LightsOnAt:    prog.LightsOnAt,
		Timezone:      prog.Timezone,
		ScheduleOnID:  ptrInt64(onID),
		ScheduleOffID: ptrInt64(offID),
		CropCycleID:   prog.CropCycleID,
		IsActive:      prog.IsActive,
		Metadata:      prog.Metadata,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update lighting program")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to commit")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, updated)
}

// POST /lighting-programs/{pid}/activate
func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	h.setActive(w, r, true)
}

// POST /lighting-programs/{pid}/deactivate
func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	h.setActive(w, r, false)
}

func (h *Handler) setActive(w http.ResponseWriter, r *http.Request, active bool) {
	id, err := programIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid lighting program id")
		return
	}
	prog, err := h.q.GetLightingProgramByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "lighting program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load lighting program")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, prog.FarmID) {
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck

	qtx := h.q.WithTx(tx)

	if prog.ScheduleOnID != nil {
		if _, err := qtx.UpdateScheduleActive(r.Context(), db.UpdateScheduleActiveParams{ID: *prog.ScheduleOnID, IsActive: active}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to update ON schedule")
			return
		}
	}
	if prog.ScheduleOffID != nil {
		if _, err := qtx.UpdateScheduleActive(r.Context(), db.UpdateScheduleActiveParams{ID: *prog.ScheduleOffID, IsActive: active}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to update OFF schedule")
			return
		}
	}

	updated, err := qtx.UpdateLightingProgramActive(r.Context(), db.UpdateLightingProgramActiveParams{
		ID:       id,
		IsActive: active,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update lighting program")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to commit")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, updated)
}

// DELETE /lighting-programs/{pid}
// Deletes the program and its paired schedules (cascade removes actions).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := programIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid lighting program id")
		return
	}
	prog, err := h.q.GetLightingProgramByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "lighting program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load lighting program")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, prog.FarmID) {
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin transaction")
		return
	}
	defer tx.Rollback(r.Context()) //nolint:errcheck

	qtx := h.q.WithTx(tx)

	// Delete schedules first (cascade removes executable_actions).
	if prog.ScheduleOnID != nil {
		_ = qtx.DeleteSchedule(r.Context(), *prog.ScheduleOnID)
	}
	if prog.ScheduleOffID != nil {
		_ = qtx.DeleteSchedule(r.Context(), *prog.ScheduleOffID)
	}

	if err := qtx.DeleteLightingProgram(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete lighting program")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to commit")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
