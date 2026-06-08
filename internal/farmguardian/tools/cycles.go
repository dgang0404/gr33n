package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "gr33n-api/internal/db"
)

func execCreateCropCycle(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	if deps.FarmID <= 0 {
		return nil, errors.New("farm_id required in proposal scope")
	}
	zoneID, err := int64FromArgs(args, "zone_id")
	if err != nil {
		return nil, err
	}
	name, err := stringFromArgs(args, "name")
	if err != nil {
		return nil, err
	}
	strain, err := stringFromArgs(args, "strain_or_variety")
	if err != nil {
		return nil, err
	}
	stage, err := stringFromArgs(args, "current_stage")
	if err != nil {
		return nil, err
	}
	startedAt, err := dateFromArgs(args, "started_at")
	if err != nil {
		return nil, err
	}
	notes, err := optionalStringFromArgs(args, "cycle_notes")
	if err != nil {
		return nil, err
	}

	z, err := deps.Q.GetZoneByID(ctx, zoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("zone %d not found", zoneID)
		}
		return nil, err
	}
	if err := ensureFarmScope(z.FarmID, deps.FarmID); err != nil {
		return nil, err
	}

	active := true
	if v, err := optionalBoolFromArgs(args, "is_active"); err != nil {
		return nil, err
	} else if v != nil {
		active = *v
	}
	if active {
		if err := ensureZoneHasNoActiveCycle(ctx, deps.Q, deps.FarmID, zoneID); err != nil {
			return nil, err
		}
	}

	var plantID *int64
	if v, err := optionalInt64FromArgs(args, "plant_id"); err != nil {
		return nil, err
	} else if v != nil && *v > 0 {
		p, err := deps.Q.GetPlant(ctx, *v)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("plant %d not found", *v)
			}
			return nil, err
		}
		if err := ensureFarmScope(p.FarmID, deps.FarmID); err != nil {
			return nil, err
		}
		plantID = v
	}

	strainPtr := &strain
	parsedStage := parseGrowthStage(stage)
	row, err := deps.Q.CreateCropCycle(ctx, db.CreateCropCycleParams{
		FarmID:          deps.FarmID,
		ZoneID:          zoneID,
		Name:            name,
		StrainOrVariety: strainPtr,
		CurrentStage:    parsedStage,
		IsActive:        active,
		StartedAt:       startedAt,
		CycleNotes:      notes,
		PlantID:         plantID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, errors.New("only one active crop cycle per zone is allowed")
		}
		return nil, err
	}
	if parsedStage != nil {
		enteredAt := startedAt.Time
		if !startedAt.Valid {
			enteredAt = row.CreatedAt
		}
		_, _ = deps.Q.InsertCropCycleStageEvent(ctx, db.InsertCropCycleStageEventParams{
			CropCycleID: row.ID,
			GrowthStage: *parsedStage,
			EnteredAt:   enteredAt.UTC(),
		})
	}
	return map[string]any{
		"crop_cycle_id":     row.ID,
		"name":              row.Name,
		"zone_id":           row.ZoneID,
		"strain_or_variety": strain,
		"current_stage":     string(canonicalGrowthStage(stage)),
	}, nil
}

func ensureZoneHasNoActiveCycle(ctx context.Context, q db.Querier, farmID, zoneID int64) error {
	cycles, err := q.ListCropCyclesByFarm(ctx, farmID)
	if err != nil {
		return err
	}
	for _, c := range cycles {
		if c.IsActive && c.ZoneID == zoneID {
			return fmt.Errorf("zone %d already has active crop cycle %q (#%d)", zoneID, c.Name, c.ID)
		}
	}
	return nil
}

func execUpdateCycleStage(ctx context.Context, deps ExecutorDeps, args map[string]any) (any, error) {
	cycleID, err := int64FromArgs(args, "crop_cycle_id")
	if err != nil {
		cycleID, err = int64FromArgs(args, "cycle_id")
		if err != nil {
			return nil, errors.New("crop_cycle_id required")
		}
	}
	stage, err := stringFromArgs(args, "current_stage")
	if err != nil {
		return nil, err
	}
	cc, err := deps.Q.GetCropCycleByID(ctx, cycleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("crop cycle %d not found", cycleID)
		}
		return nil, err
	}
	if err := ensureFarmScope(cc.FarmID, deps.FarmID); err != nil {
		return nil, err
	}
	row, err := deps.Q.UpdateCropCycleStage(ctx, db.UpdateCropCycleStageParams{
		ID:           cycleID,
		CurrentStage: parseGrowthStage(stage),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"crop_cycle_id": row.ID,
		"current_stage": string(canonicalGrowthStage(stage)),
		"cycle_name":    row.Name,
	}, nil
}

// growthStageAliases maps canonical enum values *and* common operator/Guardian
// phrasings (e.g. "vegetative", "flower", "drying") to a valid
// gr33nfertigation.growth_stage_enum member. The DB enum has no "vegetative" /
// "flower" / "drying" values, so a raw passthrough would fail on Confirm with
// SQLSTATE 22P02. Normalizing here keeps every write path (create + advance)
// safe regardless of how the stage was phrased.
var growthStageAliases = map[string]db.Gr33nfertigationGrowthStageEnum{
	// canonical identities
	"clone":        db.Gr33nfertigationGrowthStageEnumClone,
	"seedling":     db.Gr33nfertigationGrowthStageEnumSeedling,
	"early_veg":    db.Gr33nfertigationGrowthStageEnumEarlyVeg,
	"late_veg":     db.Gr33nfertigationGrowthStageEnumLateVeg,
	"transition":   db.Gr33nfertigationGrowthStageEnumTransition,
	"early_flower": db.Gr33nfertigationGrowthStageEnumEarlyFlower,
	"mid_flower":   db.Gr33nfertigationGrowthStageEnumMidFlower,
	"late_flower":  db.Gr33nfertigationGrowthStageEnumLateFlower,
	"flush":        db.Gr33nfertigationGrowthStageEnumFlush,
	"harvest":      db.Gr33nfertigationGrowthStageEnumHarvest,
	"dry_cure":     db.Gr33nfertigationGrowthStageEnumDryCure,
	// loose synonyms
	"veg":         db.Gr33nfertigationGrowthStageEnumEarlyVeg,
	"vegetative":  db.Gr33nfertigationGrowthStageEnumEarlyVeg,
	"vegetation":  db.Gr33nfertigationGrowthStageEnumEarlyVeg,
	"flower":      db.Gr33nfertigationGrowthStageEnumEarlyFlower,
	"flowering":   db.Gr33nfertigationGrowthStageEnumEarlyFlower,
	"bloom":       db.Gr33nfertigationGrowthStageEnumEarlyFlower,
	"blooming":    db.Gr33nfertigationGrowthStageEnumEarlyFlower,
	"dry":         db.Gr33nfertigationGrowthStageEnumDryCure,
	"drying":      db.Gr33nfertigationGrowthStageEnumDryCure,
	"cure":        db.Gr33nfertigationGrowthStageEnumDryCure,
	"curing":      db.Gr33nfertigationGrowthStageEnumDryCure,
	"flushing":    db.Gr33nfertigationGrowthStageEnumFlush,
	"harvesting":  db.Gr33nfertigationGrowthStageEnumHarvest,
	"cutting":     db.Gr33nfertigationGrowthStageEnumClone,
	"sprout":      db.Gr33nfertigationGrowthStageEnumSeedling,
	"germination": db.Gr33nfertigationGrowthStageEnumSeedling,
}

// canonicalGrowthStage resolves a free-text stage to a valid enum value,
// falling back to "seedling" for empty or unrecognized input.
func canonicalGrowthStage(s string) db.Gr33nfertigationGrowthStageEnum {
	key := strings.ToLower(strings.TrimSpace(s))
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")
	if v, ok := growthStageAliases[key]; ok {
		return v
	}
	return db.Gr33nfertigationGrowthStageEnumSeedling
}

func parseGrowthStage(s string) *db.Gr33nfertigationGrowthStageEnum {
	v := canonicalGrowthStage(s)
	return &v
}
