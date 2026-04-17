package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/platform/commontypes"
)

// ResolveProgramActionSource tags where a resolved program-action list came
// from. Operators and smoke tests use this to distinguish fully-migrated
// programs (source=executable_actions) from legacy ones still carrying a
// metadata.steps array (source=metadata_steps_fallback).
type ResolveProgramActionSource string

const (
	// ProgramActionsFromExecutableActions means we returned the rows
	// stored in gr33ncore.executable_actions verbatim. This is the
	// post-backfill happy path.
	ProgramActionsFromExecutableActions ResolveProgramActionSource = "executable_actions"

	// ProgramActionsFromMetadataStepsFallback means we synthesized
	// []executable_actions from programs.metadata.steps because no DB
	// rows exist yet. This keeps the worker backwards-compatible with
	// programs that haven't been touched by the backfill.
	ProgramActionsFromMetadataStepsFallback ResolveProgramActionSource = "metadata_steps_fallback"

	// ProgramActionsEmpty means neither source yielded any rows.
	ProgramActionsEmpty ResolveProgramActionSource = "empty"
)

// ResolveProgramActions returns the executable actions bound to `program`,
// preferring rows in gr33ncore.executable_actions and falling back to any
// `metadata.steps` array still attached to the program (Phase 20.9 WS4
// transitional support — the 20260515 backfill copies these into the table,
// but a pre-migration program row could still show up in a dev DB).
//
// The returned rows are not persisted in the fallback path; they only exist in
// memory so the caller can dispatch them.
func ResolveProgramActions(ctx context.Context, q *db.Queries, program db.Gr33nfertigationProgram) ([]db.Gr33ncoreExecutableAction, ResolveProgramActionSource, error) {
	rows, err := q.ListExecutableActionsByProgram(ctx, &program.ID)
	if err != nil {
		return nil, ProgramActionsEmpty, fmt.Errorf("list executable_actions for program %d: %w", program.ID, err)
	}
	if len(rows) > 0 {
		return rows, ProgramActionsFromExecutableActions, nil
	}

	// Fallback: peek at programs.metadata.steps.
	synth, err := synthesizeActionsFromMetadata(program)
	if err != nil {
		// A malformed metadata.steps should never kill the worker —
		// it just means "no actions". The backfill logs a NOTICE for
		// the same row so the operator already has a breadcrumb.
		log.Printf("program %d: metadata.steps unusable: %v", program.ID, err)
		return nil, ProgramActionsEmpty, nil
	}
	if len(synth) == 0 {
		return nil, ProgramActionsEmpty, nil
	}
	return synth, ProgramActionsFromMetadataStepsFallback, nil
}

// ResolveProgramActionsByID looks up the program and forwards to
// ResolveProgramActions. Convenience wrapper so callers that only have an ID
// don't need to import the db package.
func ResolveProgramActionsByID(ctx context.Context, pool *pgxpool.Pool, programID int64) ([]db.Gr33ncoreExecutableAction, ResolveProgramActionSource, error) {
	q := db.New(pool)
	prog, err := q.GetFertigationProgramByID(ctx, programID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ProgramActionsEmpty, nil
		}
		return nil, ProgramActionsEmpty, err
	}
	return ResolveProgramActions(ctx, q, prog)
}

// synthesizeActionsFromMetadata parses programs.metadata.steps (if present)
// and returns in-memory Gr33ncoreExecutableAction rows. The shape of
// metadata.steps is taken from the 20260515 backfill migration, which accepts:
//
//	steps: [
//	  {
//	    "action_type": "control_actuator" | "create_task" | "send_notification",
//	    "target_actuator_id": <int>,
//	    "target_notification_template_id": <int>,
//	    "action_command": "<string>",
//	    "action_parameters": {...},
//	    "delay_before_execution_seconds": <int>
//	  },
//	  ...
//	]
//
// Anything that fails to parse is skipped silently (same posture as the
// backfill's NOTICE-and-continue behaviour) so a single malformed step never
// hides the rest.
func synthesizeActionsFromMetadata(program db.Gr33nfertigationProgram) ([]db.Gr33ncoreExecutableAction, error) {
	if len(program.Metadata) == 0 {
		return nil, nil
	}
	var meta struct {
		Steps []json.RawMessage `json:"steps"`
	}
	if err := json.Unmarshal(program.Metadata, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	if len(meta.Steps) == 0 {
		return nil, nil
	}
	out := make([]db.Gr33ncoreExecutableAction, 0, len(meta.Steps))
	for i, raw := range meta.Steps {
		var step struct {
			ActionType                   string          `json:"action_type"`
			TargetActuatorID             *int64          `json:"target_actuator_id"`
			TargetNotificationTemplateID *int64          `json:"target_notification_template_id"`
			ActionCommand                *string         `json:"action_command"`
			ActionParameters             json.RawMessage `json:"action_parameters"`
			DelayBeforeExecutionSeconds  *int32          `json:"delay_before_execution_seconds"`
		}
		if err := json.Unmarshal(raw, &step); err != nil {
			continue
		}
		if step.ActionType == "" {
			continue
		}
		progID := program.ID
		out = append(out, db.Gr33ncoreExecutableAction{
			ID:                           0, // synthetic; not persisted
			ProgramID:                    &progID,
			ExecutionOrder:               int32(i + 1),
			ActionType:                   commontypes.ExecutableActionTypeEnum(step.ActionType),
			TargetActuatorID:             step.TargetActuatorID,
			TargetAutomationRuleID:       nil,
			TargetNotificationTemplateID: step.TargetNotificationTemplateID,
			ActionCommand:                step.ActionCommand,
			ActionParameters:             []byte(step.ActionParameters),
			DelayBeforeExecutionSeconds:  step.DelayBeforeExecutionSeconds,
		})
	}
	return out, nil
}
