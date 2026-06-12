package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

type growSetupPack struct {
	Profile  string
	ZoneID   int64
	ZoneName string
	Plant    map[string]any
	Cycle    map[string]any
	Program  map[string]any
	OptTask  map[string]any
}

func execApplyGrowSetupPack(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	pack, err := parseGrowSetupPackArgs(args)
	if err != nil {
		return nil, err
	}
	if err := validateGrowSetupPack(ctx, deps.Q, deps.FarmID, pack); err != nil {
		return nil, err
	}
	if pack.ZoneName == "" {
		if z, zerr := deps.Q.GetZoneByID(ctx, pack.ZoneID); zerr == nil {
			pack.ZoneName = z.Name
		}
	}
	if deps.Pool == nil {
		return nil, errors.New("setup pack requires database pool for transaction")
	}
	baseQ, ok := deps.Q.(*db.Queries)
	if !ok {
		return nil, errors.New("setup pack requires database queries")
	}

	tx, err := deps.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	txDeps := deps
	txDeps.Q = baseQ.WithTx(tx)

	result := map[string]any{
		"profile":   pack.Profile,
		"zone_id":   pack.ZoneID,
		"zone_name": pack.ZoneName,
	}

	var plantVariety string
	if pack.Plant != nil {
		plantArgs, err := plantArgsFromSetupPack(pack.Plant)
		if err != nil {
			return nil, err
		}
		plantOut, err := execCreatePlant(ctx, txDeps, plantArgs)
		if err != nil {
			return nil, err
		}
		result["plant"] = plantOut
		if m, ok := plantOut.(map[string]any); ok {
			if v, ok := m["variety_or_cultivar"].(string); ok {
				plantVariety = v
			}
		}
	}

	cycleArgs, err := cycleArgsFromSetupPack(pack, plantVariety)
	if err != nil {
		return nil, err
	}
	cycleOut, err := execCreateCropCycle(ctx, txDeps, cycleArgs)
	if err != nil {
		return nil, err
	}
	result["cycle"] = cycleOut

	programArgs, err := programArgsFromSetupPack(pack)
	if err != nil {
		return nil, err
	}
	programOut, err := execCreateFertigationProgram(ctx, txDeps, programArgs)
	if err != nil {
		return nil, err
	}
	result["program"] = programOut

	cycleID, programID, err := setupPackIDs(cycleOut, programOut)
	if err != nil {
		return nil, err
	}
	if err := linkCropCyclePrimaryProgram(ctx, txDeps.Q, cycleID, programID); err != nil {
		return nil, err
	}
	result["primary_program_linked"] = true

	if pack.OptTask != nil {
		taskArgs, err := taskArgsFromSetupPack(pack.OptTask, pack.ZoneID)
		if err != nil {
			return nil, err
		}
		taskOut, err := execCreateTask(ctx, txDeps, taskArgs)
		if err != nil {
			return nil, err
		}
		result["optional_task"] = taskOut
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func parseGrowSetupPackArgs(args map[string]any) (growSetupPack, error) {
	var pack growSetupPack
	zoneID, err := int64FromArgs(args, "zone_id")
	if err != nil {
		return pack, err
	}
	pack.ZoneID = zoneID
	if name, err := optionalStringFromArgs(args, "zone_name"); err != nil {
		return pack, err
	} else if name != nil {
		pack.ZoneName = *name
	}
	if profile, err := optionalStringFromArgs(args, "profile"); err != nil {
		return pack, err
	} else if profile != nil {
		pack.Profile = *profile
	}
	cycle, err := objectFromArgs(args, "cycle")
	if err != nil {
		return pack, err
	}
	pack.Cycle = cycle
	program, err := objectFromArgs(args, "program")
	if err != nil {
		return pack, err
	}
	pack.Program = program
	if plant, ok, err := optionalObjectFromArgs(args, "plant"); err != nil {
		return pack, err
	} else if ok {
		pack.Plant = plant
	}
	if task, ok, err := optionalObjectFromArgs(args, "optional_task"); err != nil {
		return pack, err
	} else if ok {
		pack.OptTask = task
	}
	return pack, nil
}

func validateGrowSetupPack(ctx context.Context, q db.Querier, farmID int64, pack growSetupPack) error {
	z, err := q.GetZoneByID(ctx, pack.ZoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("zone %d not found", pack.ZoneID)
		}
		return err
	}
	if err := ensureFarmScope(z.FarmID, farmID); err != nil {
		return err
	}
	if err := ensureZoneHasNoActiveCycle(ctx, q, farmID, pack.ZoneID); err != nil {
		return err
	}
	if pack.Plant != nil {
		displayName, err := stringFromArgs(pack.Plant, "display_name")
		if err != nil {
			return fmt.Errorf("plant.%w", err)
		}
		plants, err := q.ListPlantsByFarm(ctx, farmID)
		if err != nil {
			return err
		}
		for _, p := range plants {
			if strings.EqualFold(strings.TrimSpace(p.DisplayName), displayName) {
				return fmt.Errorf("plant %q already exists on this farm (#%d)", displayName, p.ID)
			}
		}
	}
	return nil
}

func plantArgsFromSetupPack(plant map[string]any) (map[string]any, error) {
	out := map[string]any{
		"display_name": plant["display_name"],
	}
	if v, ok := plant["variety_or_cultivar"]; ok {
		out["variety_or_cultivar"] = v
	}
	meta := map[string]any{}
	if notes, err := optionalStringFromArgs(plant, "notes"); err != nil {
		return nil, err
	} else if notes != nil {
		meta["notes"] = *notes
	}
	if len(meta) > 0 {
		out["meta"] = meta
	}
	return out, nil
}

func cycleArgsFromSetupPack(pack growSetupPack, plantVariety string) (map[string]any, error) {
	out := map[string]any{
		"zone_id":       pack.ZoneID,
		"name":          pack.Cycle["name"],
		"current_stage": pack.Cycle["current_stage"],
		"started_at":    pack.Cycle["started_at"],
	}
	if batch, ok := pack.Cycle["batch_label"]; ok && batch != nil {
		out["batch_label"] = batch
	} else if strain, ok := pack.Cycle["strain_or_variety"]; ok && strain != nil {
		out["batch_label"] = strain
	} else if plantVariety != "" {
		out["batch_label"] = plantVariety
	} else if pack.Plant != nil {
		if v, ok := pack.Plant["display_name"]; ok {
			out["batch_label"] = v
		}
	}
	if notes, err := optionalStringFromArgs(pack.Cycle, "cycle_notes"); err != nil {
		return nil, err
	} else if notes != nil {
		out["cycle_notes"] = *notes
	}
	return out, nil
}

func programArgsFromSetupPack(pack growSetupPack) (map[string]any, error) {
	out := map[string]any{
		"name":                pack.Program["name"],
		"target_zone_id":      pack.ZoneID,
		"total_volume_liters": pack.Program["total_volume_liters"],
		"ec_trigger_low":      pack.Program["ec_trigger_low"],
		"ph_trigger_low":      pack.Program["ph_trigger_low"],
		"ph_trigger_high":     pack.Program["ph_trigger_high"],
	}
	if v, ok := pack.Program["is_active"]; ok {
		out["is_active"] = v
	}
	if desc, err := optionalStringFromArgs(pack.Program, "description"); err != nil {
		return nil, err
	} else if desc != nil {
		out["description"] = *desc
	}
	return out, nil
}

func taskArgsFromSetupPack(task map[string]any, zoneID int64) (map[string]any, error) {
	title, err := stringFromArgs(task, "title")
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"title":   title,
		"zone_id": zoneID,
	}
	if desc, err := optionalStringFromArgs(task, "description"); err != nil {
		return nil, err
	} else if desc != nil {
		out["description"] = *desc
	}
	out["task_type"] = "general"
	return out, nil
}

func setupPackIDs(cycleOut, programOut any) (int64, int64, error) {
	cycleMap, ok := cycleOut.(map[string]any)
	if !ok {
		return 0, 0, errors.New("cycle result missing")
	}
	programMap, ok := programOut.(map[string]any)
	if !ok {
		return 0, 0, errors.New("program result missing")
	}
	cycleID, err := int64FromArgs(cycleMap, "crop_cycle_id")
	if err != nil {
		return 0, 0, err
	}
	programID, err := int64FromArgs(programMap, "program_id")
	if err != nil {
		return 0, 0, err
	}
	return cycleID, programID, nil
}

func linkCropCyclePrimaryProgram(ctx context.Context, q db.Querier, cycleID, programID int64) error {
	cc, err := q.GetCropCycleByID(ctx, cycleID)
	if err != nil {
		return err
	}
	pid := programID
	_, err = q.UpdateCropCycle(ctx, db.UpdateCropCycleParams{
		ID:               cc.ID,
		Name:             cc.Name,
		BatchLabel:  cc.BatchLabel,
		ZoneID:           cc.ZoneID,
		IsActive:         cc.IsActive,
		CycleNotes:       cc.CycleNotes,
		HarvestedAt:      cc.HarvestedAt,
		YieldGrams:       cc.YieldGrams,
		YieldNotes:       cc.YieldNotes,
		PrimaryProgramID: &pid,
	})
	return err
}

// GrowSetupPackSummary renders a one-line operator summary for proposal cards.
func GrowSetupPackSummary(pack map[string]any) string {
	zone := "zone"
	if n, ok := pack["zone_name"].(string); ok && strings.TrimSpace(n) != "" {
		zone = strings.TrimSpace(n)
	}
	plant := "grow setup"
	if p, ok := pack["plant"].(map[string]any); ok {
		if n, ok := p["display_name"].(string); ok && strings.TrimSpace(n) != "" {
			plant = strings.TrimSpace(n)
		}
	}
	return fmt.Sprintf("Setup pack: %s in %s (plant + cycle + program)", plant, zone)
}
