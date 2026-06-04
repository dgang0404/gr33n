package fertigation

// Phase 39 WS3 — mix_batch command type and operator enqueue / preview API.
//
// Routes (registered in cmd/api/routes.go):
//
//	POST /farms/{id}/fertigation/mix-jobs        — enqueue or preview a mix_batch
//	GET  /fertigation/programs/{rid}/mix-preview — read-only plan for Zone Water tab

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/fertigation/mixplan"
	"gr33n-api/internal/httputil"
)

// POST /farms/{id}/fertigation/mix-jobs
//
// Body:
//
//	{
//	  "program_id":          42,
//	  "preview_only":        false,   // true → calculate without enqueuing
//	  "pump_flow_rate_ml_per_sec": 10 // optional override
//	}
//
// On enqueue, finds the reservoir's delivery_actuator → its device, then calls
// EnqueueDeviceCommand(mix_batch). Returns the MixPlan + command_id (if enqueued).
func (h *Handler) EnqueueMixJob(w http.ResponseWriter, r *http.Request) {
	farmID, err := farmIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}

	var body struct {
		ProgramID            int64   `json:"program_id"`
		PreviewOnly          bool    `json:"preview_only"`
		PumpFlowRateMLPerSec float64 `json:"pump_flow_rate_ml_per_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.ProgramID == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "program_id is required")
		return
	}

	ctx := r.Context()

	// ── Build MixPlan Input from DB ──────────────────────────────────────────
	prog, err := h.q.GetFertigationProgramByID(ctx, body.ProgramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load program")
		return
	}
	if prog.FarmID != farmID {
		httputil.WriteError(w, http.StatusForbidden, "program belongs to a different farm")
		return
	}

	in, err := mixplan.BuildFromProgramRow(ctx, h.q, prog, mixplan.BuildOptions{
		PumpFlowRateMlPerSec: body.PumpFlowRateMLPerSec,
	})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, mixplan.ErrProgramHasNoRecipe) {
			httputil.WriteError(w, http.StatusBadRequest,
				"program has no application_recipe_id — plain irrigation programs do not require a mix step")
			return
		}
		httputil.WriteError(w, status, err.Error())
		return
	}

	// ── Calculate ────────────────────────────────────────────────────────────
	plan, err := mixplan.Calculate(in)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if body.PreviewOnly {
		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"preview_only": true,
			"mix_plan":     plan,
		})
		return
	}

	// ── Resolve device to enqueue on ────────────────────────────────────────
	// Use the reservoir's delivery_actuator → device chain. Falls back to
	// the first actuator on the zone if delivery_actuator is unset.
	res, err := h.q.GetFertigationReservoirByID(ctx, plan.ReservoirID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load reservoir")
		return
	}
	if res.DeliveryActuatorID == nil {
		httputil.WriteError(w, http.StatusBadRequest,
			"reservoir has no delivery_actuator_id; link an actuator to the reservoir before enqueueing a mix job")
		return
	}
	actuator, err := h.q.GetActuatorByID(ctx, *res.DeliveryActuatorID)
	if err != nil || actuator.DeviceID == nil {
		httputil.WriteError(w, http.StatusBadRequest,
			"delivery actuator is not bound to a device; bind a Pi device before enqueueing")
		return
	}
	device, err := h.q.GetDeviceByID(ctx, *actuator.DeviceID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device")
		return
	}
	if device.FarmID != farmID {
		httputil.WriteError(w, http.StatusForbidden, "device belongs to a different farm")
		return
	}

	// ── Build mix_batch payload ──────────────────────────────────────────────
	payload, err := buildMixBatchPayload(plan, prog.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to build mix_batch payload")
		return
	}

	cmd, err := h.q.EnqueueDeviceCommand(ctx, db.EnqueueDeviceCommandParams{
		DeviceID:    device.ID,
		FarmID:      farmID,
		CommandType: "mix_batch",
		Payload:     payload,
		Source:      "operator",
		ProgramID:   &prog.ID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to enqueue mix_batch command")
		return
	}

	// Mirror legacy pending_command for pre-39 Pi clients.
	_ = h.q.SetDevicePendingCommand(ctx, db.SetDevicePendingCommandParams{
		ID:      device.ID,
		Column2: payload,
	})

	httputil.WriteJSON(w, http.StatusAccepted, map[string]any{
		"command_id": cmd.ID,
		"device_id":  device.ID,
		"mix_plan":   plan,
	})
}

// GET /fertigation/programs/{rid}/mix-preview
// Read-only: calculate and return a MixPlan without writing anything to the DB.
// Used by the Zone Water tab "Preview mix" button.
func (h *Handler) MixPreview(w http.ResponseWriter, r *http.Request) {
	progID, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid program id")
		return
	}

	ctx := r.Context()
	prog, err := h.q.GetFertigationProgramByID(ctx, progID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load program")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, prog.FarmID) {
		return
	}

	in, err := mixplan.BuildFromProgramRow(ctx, h.q, prog, mixplan.BuildOptions{})
	if err != nil {
		if errors.Is(err, mixplan.ErrProgramHasNoRecipe) {
			httputil.WriteJSON(w, http.StatusOK, map[string]any{
				"mix_required": false,
				"reason":       "program has no application_recipe_id (plain irrigation)",
			})
			return
		}
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	plan, err := mixplan.Calculate(in)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"mix_required": true,
		"mix_plan":     plan,
	})
}

// GET /fertigation/programs/{rid}/water-status
//
// One-shot read for the Zone Water tab (Phase 39 WS7). Returns:
//   - mix_preview:       MixPlan preview (nil if no recipe)
//   - mix_required:      bool
//   - queue_depth:       pending+in_progress command count on the delivery device
//   - last_mixing_event: most recent mixing event linked to this program
func (h *Handler) WaterStatus(w http.ResponseWriter, r *http.Request) {
	progID, err := resourceIDFromPath(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid program id")
		return
	}
	ctx := r.Context()
	prog, err := h.q.GetFertigationProgramByID(ctx, progID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "program not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load program")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, prog.FarmID) {
		return
	}

	resp := map[string]any{
		"program_id": prog.ID,
	}

	// ── Mix preview ──────────────────────────────────────────────────────────
	in, buildErr := mixplan.BuildFromProgramRow(ctx, h.q, prog, mixplan.BuildOptions{})
	if buildErr != nil {
		resp["mix_required"] = false
		resp["mix_preview"] = nil
		resp["mix_preview_error"] = buildErr.Error()
	} else {
		plan, calcErr := mixplan.Calculate(in)
		if calcErr != nil {
			resp["mix_required"] = true
			resp["mix_preview"] = nil
			resp["mix_preview_error"] = calcErr.Error()
		} else {
			resp["mix_required"] = true
			resp["mix_preview"] = plan
		}
	}

	// ── Queue depth ──────────────────────────────────────────────────────────
	var queueDepth int64
	if prog.ReservoirID != nil {
		res, rerr := h.q.GetFertigationReservoirByID(ctx, *prog.ReservoirID)
		if rerr == nil && res.DeliveryActuatorID != nil {
			act, aerr := h.q.GetActuatorByID(ctx, *res.DeliveryActuatorID)
			if aerr == nil && act.DeviceID != nil {
				queueDepth, _ = h.q.CountPendingCommandsByDevice(ctx, *act.DeviceID)
			}
		}
	}
	resp["queue_depth"] = queueDepth

	// ── Last mixing event ─────────────────────────────────────────────────────
	events, lerr := h.q.ListMixingEventsByFarm(ctx, prog.FarmID)
	if lerr == nil {
		for _, ev := range events {
			if ev.ProgramID != nil && *ev.ProgramID == prog.ID {
				resp["last_mixing_event"] = ev
				break
			}
		}
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// buildMixBatchPayload marshals the MixPlan + metadata into the device_commands.payload JSONB.
func buildMixBatchPayload(plan mixplan.MixPlan, programID int64) ([]byte, error) {
	m := map[string]any{
		"command_type":  "mix_batch",
		"program_id":    programID,
		"reservoir_id":  plan.ReservoirID,
		"mix_plan":      plan,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal mix_batch payload: %w", err)
	}
	return b, nil
}

// programIDFromPath reads the program id from an {rid} path value (reuses
// the fertigation handler's resourceIDFromPath convention).
func programIDStr(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("rid"), 10, 64)
}
