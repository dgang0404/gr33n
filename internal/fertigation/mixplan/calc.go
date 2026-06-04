// Package mixplan implements the Phase 39 WS2 server-side mix dose calculator.
//
// # Design scope (v1)
//
// Cloud owns the recipe math — the Pi only receives run_seconds per channel.
//
// Algorithm:
//  1. Parse dilution ratio ("1:500") → concentrate_fraction = 1/500.
//  2. total_concentrate_ml = water_volume_liters * concentrate_fraction * 1000.
//  3. Distribute concentrate across components by their relative part_value.
//     If all part_values are zero or absent, each component gets an equal share.
//  4. run_seconds[i] = component_ml[i] / pump_flow_rate_ml_per_sec (default 10 ml/s).
//  5. EC estimate: linear heuristic — delta added = (concentrate_fraction * ECFactor)
//     where ECFactor defaults to 500 mS/cm (documents that calibration is farm-specific).
//
// Limitations (v1, documented):
//   - EC estimate is uncalibrated; it is a rough guide, not a closed-loop guarantee.
//     Phase 39 WS2 v2 will add closed-loop dosing with inline sensor feedback.
//   - DilutionRatio is required; programs without a ratio return ErrNoDilutionRatio.
//   - PumpFlowRateMlPerSec defaults to 10 ml/s; operators should calibrate pumps and
//     override via Pi config before relying on precise volumes.
package mixplan

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ErrNoDilutionRatio is returned when neither the program nor the recipe
// carries a parseable dilution_ratio string (required for v1 volume math).
var ErrNoDilutionRatio = errors.New("dilution_ratio is required for mix calculation (e.g. \"1:500\"); set it on the program or recipe")

// ErrNoComponents is returned when the recipe has no input components.
var ErrNoComponents = errors.New("recipe has no input components; add at least one before calculating a mix plan")

// ErrBaseECUnknown is returned when the reservoir has no base EC reading.
// This is enforced by the enqueue path (WS3/WS6) — the calculator itself
// accepts zero and emits a warning rather than refusing.
var ErrBaseECUnknown = errors.New("reservoir base EC is unknown (0); set it via \"Set base water EC\" before calculating a mix plan")

// DefaultPumpFlowRateMLPerSec is the assumed pump flow rate when none is
// provided by the caller. Operators should calibrate their pumps and pass
// the measured rate via Input.PumpFlowRateMlPerSec.
const DefaultPumpFlowRateMLPerSec = 10.0

// ECFactorMsPerConcentrateFraction is a rough dilution-factor heuristic:
// at a 1:1 concentrate fraction, a typical JADAM/liquid fertiliser
// contributes approximately 500 mS/cm. Actual values vary widely by
// input type; this is explicitly an estimate.
//
// estimated_ec_delta ≈ concentrate_fraction * ECFactor
const ECFactorMsPerConcentrateFraction = 500.0

// MixPlan is the output of Calculate. It carries the ordered steps the Pi
// will execute (channel → seconds) plus human-readable context.
type MixPlan struct {
	ReservoirID int64 `json:"reservoir_id"`

	WaterVolumeLiters float64 `json:"water_volume_liters"`
	WaterEcMscm       float64 `json:"water_ec_mscm"`
	TargetEcMscm      float64 `json:"target_ec_mscm"`
	DilutionRatio     string  `json:"dilution_ratio"`

	Steps []MixStep `json:"steps"`

	EstimatedFinalEcMscm float64  `json:"estimated_final_ec_mscm"`
	Warnings             []string `json:"warnings,omitempty"`
}

// MixStep is one pump operation in the ordered sequence.
type MixStep struct {
	Step int `json:"step"`

	InputDefinitionID int64  `json:"input_definition_id"`
	InputName         string `json:"input_name"`

	// ChannelIndex is the 1-indexed Pi channel number (pump slot).
	// The Pi maps channels to GPIO pins via its local config.
	// Steps are assigned channels 1…N in component order.
	ChannelIndex int `json:"channel"`

	VolumeMl   float64 `json:"volume_ml"`
	RunSeconds int     `json:"run_seconds"`

	Notes string `json:"notes,omitempty"`
}

// ComponentInput is one line from the recipe's input_components table,
// converted to plain Go floats.
type ComponentInput struct {
	InputDefinitionID int64
	InputName         string
	// PartValue is the relative part (e.g. 1 in JLF:JLW = 1:1).
	// A value of 0 means "equal share with other zero-valued components".
	PartValue float64
	Notes     string
}

// Input gathers all the information needed to produce a MixPlan.
type Input struct {
	ReservoirID int64

	// WaterVolumeLiters is the total batch volume (program.total_volume_liters).
	WaterVolumeLiters float64

	// BaseEcMscm is the measured EC of the source water (reservoir.last_ec_mscm).
	// May be zero if unknown; a warning is emitted.
	BaseEcMscm float64

	// TargetEcMscm is the desired EC after mixing (ec_target.ec_min_mscm or
	// the lower bound of the target range).
	TargetEcMscm float64

	// DilutionRatioStr is a string like "1:500" from program.dilution_ratio
	// or recipe.dilution_ratio. Required.
	DilutionRatioStr string

	// Components are the recipe input lines ordered by channel assignment.
	// Must be non-empty.
	Components []ComponentInput

	// PumpFlowRateMlPerSec overrides DefaultPumpFlowRateMLPerSec when > 0.
	PumpFlowRateMlPerSec float64
}

// Calculate produces a MixPlan from the given Input.
//
// It returns ErrNoDilutionRatio if DilutionRatioStr cannot be parsed, and
// ErrNoComponents if Components is empty. Other errors are not expected from
// pure computation (no I/O).
func Calculate(in Input) (MixPlan, error) {
	if len(in.Components) == 0 {
		return MixPlan{}, ErrNoComponents
	}

	// ── 1. Parse dilution ratio ──────────────────────────────────────────────
	num, den, err := parseDilutionRatio(in.DilutionRatioStr)
	if err != nil {
		return MixPlan{}, ErrNoDilutionRatio
	}
	concentrateFraction := num / den // e.g. 1/500 = 0.002

	// ── 2. Total concentrate volume ──────────────────────────────────────────
	totalConcentrateML := in.WaterVolumeLiters * concentrateFraction * 1000.0

	// ── 3. Distribute across components by part_value ────────────────────────
	partsTotal := 0.0
	for _, c := range in.Components {
		partsTotal += c.PartValue
	}
	equalShare := partsTotal == 0

	flowRate := in.PumpFlowRateMlPerSec
	if flowRate <= 0 {
		flowRate = DefaultPumpFlowRateMLPerSec
	}

	steps := make([]MixStep, 0, len(in.Components))
	for i, c := range in.Components {
		var share float64
		if equalShare {
			share = 1.0 / float64(len(in.Components))
		} else {
			share = c.PartValue / partsTotal
		}
		componentML := totalConcentrateML * share
		runSeconds := int(math.Ceil(componentML / flowRate))
		if runSeconds < 1 {
			runSeconds = 1
		}

		notes := c.Notes
		if notes == "" {
			notes = fmt.Sprintf("%s %.1f ml @ %.0f ml/s", c.InputName, componentML, flowRate)
		}

		steps = append(steps, MixStep{
			Step:              i + 1,
			InputDefinitionID: c.InputDefinitionID,
			InputName:         c.InputName,
			ChannelIndex:      i + 1,
			VolumeMl:          math.Round(componentML*10) / 10,
			RunSeconds:        runSeconds,
			Notes:             notes,
		})
	}

	// ── 4. EC estimate (heuristic v1) ────────────────────────────────────────
	ecDelta := concentrateFraction * ECFactorMsPerConcentrateFraction
	estimatedFinalEC := in.BaseEcMscm + ecDelta

	// ── 5. Warnings ──────────────────────────────────────────────────────────
	var warnings []string
	if in.BaseEcMscm == 0 {
		warnings = append(warnings, "base EC is unknown (0); EC estimate is unreliable — set reservoir base EC before mixing")
	}
	if in.TargetEcMscm > 0 && math.Abs(estimatedFinalEC-in.TargetEcMscm)/in.TargetEcMscm > 0.3 {
		warnings = append(warnings, fmt.Sprintf(
			"estimated EC %.2f mS/cm deviates >30%% from target %.2f mS/cm — calibrate pump flow rate or adjust dilution ratio",
			estimatedFinalEC, in.TargetEcMscm))
	}
	if equalShare && len(in.Components) > 1 {
		warnings = append(warnings, "no part_value set on recipe components — using equal share per component; set part_value for precise dosing")
	}
	warnings = append(warnings,
		"EC estimate is a volume-based heuristic (v1); actual EC depends on input EC concentration. Verify with sensor after mixing.")

	plan := MixPlan{
		ReservoirID:          in.ReservoirID,
		WaterVolumeLiters:    in.WaterVolumeLiters,
		WaterEcMscm:          in.BaseEcMscm,
		TargetEcMscm:         in.TargetEcMscm,
		DilutionRatio:        in.DilutionRatioStr,
		Steps:                steps,
		EstimatedFinalEcMscm: math.Round(estimatedFinalEC*100) / 100,
		Warnings:             warnings,
	}
	return plan, nil
}

// parseDilutionRatio parses "1:500" → (1.0, 500.0).
// Also accepts "1/500" and bare decimal fractions like "0.002".
func parseDilutionRatio(s string) (num, den float64, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, fmt.Errorf("empty dilution ratio")
	}

	// Try "N:D" or "N/D"
	for _, sep := range []string{":", "/"} {
		if idx := strings.Index(s, sep); idx >= 0 {
			nStr := strings.TrimSpace(s[:idx])
			dStr := strings.TrimSpace(s[idx+1:])
			n, errN := strconv.ParseFloat(nStr, 64)
			d, errD := strconv.ParseFloat(dStr, 64)
			if errN != nil || errD != nil || d == 0 {
				return 0, 0, fmt.Errorf("invalid dilution ratio %q", s)
			}
			return n, d, nil
		}
	}

	// Try bare float (already a fraction e.g. 0.002)
	f, errF := strconv.ParseFloat(s, 64)
	if errF != nil || f <= 0 || f > 1 {
		return 0, 0, fmt.Errorf("unrecognised dilution ratio %q (expected N:D format)", s)
	}
	return f, 1.0, nil
}
