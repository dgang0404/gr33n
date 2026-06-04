package mixplan

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

// ProgramQuerier is the DB interface the builder needs. Using a minimal
// interface keeps the package testable without a full Querier mock.
type ProgramQuerier interface {
	GetFertigationProgramByID(ctx context.Context, id int64) (db.Gr33nfertigationProgram, error)
	GetFertigationReservoirByID(ctx context.Context, id int64) (db.Gr33nfertigationReservoir, error)
	GetEcTargetByID(ctx context.Context, id int64) (db.Gr33nfertigationEcTarget, error)
	GetRecipeByID(ctx context.Context, id int64) (db.Gr33nnaturalfarmingApplicationRecipe, error)
	ListRecipeComponents(ctx context.Context, applicationRecipeID int64) ([]db.ListRecipeComponentsRow, error)
}

// ErrProgramHasNoRecipe is returned when a program carries no
// application_recipe_id — no mix is needed for water-only programs.
var ErrProgramHasNoRecipe = errors.New("program has no application_recipe_id; no mix plan required (plain irrigation)")

// ErrProgramHasNoReservoir is returned when a program has no reservoir_id.
var ErrProgramHasNoReservoir = errors.New("program has no reservoir_id; set a reservoir before calculating a mix plan")

// ErrReservoirBaseECUnknown is returned when the reservoir's last_ec_mscm is
// zero or null (WS6 will add an operator UI to fix this).
var ErrReservoirBaseECUnknown = errors.New("reservoir base EC is unknown; set it via \"Set base water EC\" on the reservoir card before calculating a mix plan")

// BuildFromProgram loads all the ingredients for a MixPlan from the DB
// and returns a ready-to-call Input. It does not call Calculate itself so
// callers can inspect / tweak the Input before computing.
//
// Returns ErrProgramHasNoRecipe when the program is a plain-irrigation type
// (no mix needed). Callers that encounter this should skip the mix_batch
// enqueue and proceed directly to the pulse irrigate step.
func BuildFromProgram(ctx context.Context, q ProgramQuerier, programID int64, opts BuildOptions) (Input, error) {
	prog, err := q.GetFertigationProgramByID(ctx, programID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Input{}, fmt.Errorf("program %d not found", programID)
		}
		return Input{}, fmt.Errorf("load program: %w", err)
	}
	return BuildFromProgramRow(ctx, q, prog, opts)
}

// BuildOptions are optional overrides that the caller can pass to
// BuildFromProgramRow. Zero values fall back to DB values or defaults.
type BuildOptions struct {
	// PumpFlowRateMlPerSec overrides the default 10 ml/s.
	// Zero means "use default".
	PumpFlowRateMlPerSec float64
}

// BuildFromProgramRow builds an Input from an already-loaded program row.
// This variant is used by the automation worker (which already holds the row).
func BuildFromProgramRow(ctx context.Context, q ProgramQuerier, prog db.Gr33nfertigationProgram, opts BuildOptions) (Input, error) {
	if prog.ApplicationRecipeID == nil {
		return Input{}, ErrProgramHasNoRecipe
	}
	if prog.ReservoirID == nil {
		return Input{}, ErrProgramHasNoReservoir
	}

	// ── Reservoir ────────────────────────────────────────────────────────────
	res, err := q.GetFertigationReservoirByID(ctx, *prog.ReservoirID)
	if err != nil {
		return Input{}, fmt.Errorf("load reservoir: %w", err)
	}
	baseEC := numericToFloat(res.LastEcMscm)
	if baseEC == 0 {
		return Input{}, ErrReservoirBaseECUnknown
	}

	// ── EC target ────────────────────────────────────────────────────────────
	var targetEC float64
	if prog.EcTargetID != nil {
		ecTarget, err := q.GetEcTargetByID(ctx, *prog.EcTargetID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return Input{}, fmt.Errorf("load ec_target: %w", err)
		}
		if err == nil {
			// Use the lower bound of the target window.
			targetEC = numericToFloat(ecTarget.EcMinMscm)
		}
	}

	// ── Recipe ───────────────────────────────────────────────────────────────
	recipe, err := q.GetRecipeByID(ctx, *prog.ApplicationRecipeID)
	if err != nil {
		return Input{}, fmt.Errorf("load recipe: %w", err)
	}

	// ── Components ───────────────────────────────────────────────────────────
	rows, err := q.ListRecipeComponents(ctx, recipe.ID)
	if err != nil {
		return Input{}, fmt.Errorf("list recipe components: %w", err)
	}
	if len(rows) == 0 {
		return Input{}, ErrNoComponents
	}
	components := make([]ComponentInput, len(rows))
	for i, r := range rows {
		notes := ""
		if r.Notes != nil {
			notes = *r.Notes
		}
		components[i] = ComponentInput{
			InputDefinitionID: r.InputDefinitionID,
			InputName:         r.InputName,
			PartValue:         numericToFloat(r.PartValue),
			Notes:             notes,
		}
	}

	// ── Dilution ratio ───────────────────────────────────────────────────────
	// Program takes precedence over recipe (operator can override per-run).
	dilution := ""
	if prog.DilutionRatio != nil && *prog.DilutionRatio != "" {
		dilution = *prog.DilutionRatio
	} else if recipe.DilutionRatio != nil && *recipe.DilutionRatio != "" {
		dilution = *recipe.DilutionRatio
	}
	if dilution == "" {
		return Input{}, ErrNoDilutionRatio
	}

	// ── Volume ───────────────────────────────────────────────────────────────
	totalVol := numericToFloat(prog.TotalVolumeLiters)
	if totalVol <= 0 {
		totalVol = numericToFloat(res.CurrentVolumeLiters)
	}
	if totalVol <= 0 {
		return Input{}, fmt.Errorf("program.total_volume_liters is zero and reservoir.current_volume_liters is also zero; set a batch volume on the program")
	}

	return Input{
		ReservoirID:          res.ID,
		WaterVolumeLiters:    totalVol,
		BaseEcMscm:           baseEC,
		TargetEcMscm:         targetEC,
		DilutionRatioStr:     dilution,
		Components:           components,
		PumpFlowRateMlPerSec: opts.PumpFlowRateMlPerSec,
	}, nil
}

// numericToFloat converts a pgtype.Numeric to float64.
// Returns 0 on any error (null, NaN, overflow).
func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}
