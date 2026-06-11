package croplibrary

import (
	"fmt"
	"strings"
)

// GenerateSeedSQL emits an idempotent migration patch (Phase 64 pattern) for crops with stages.
func GenerateSeedSQL(c *Catalog) string {
	var b strings.Builder
	b.WriteString("-- Generated from data/crop_library.yaml — do not edit by hand.\n")
	b.WriteString("-- Regenerate: ./scripts/generate-crop-seed.sql.sh\n")
	b.WriteString(fmt.Sprintf("-- crop_library version: %d\n\n", c.Version))

	b.WriteString(`-- Built-in profiles (idempotent — skip if crop_key already exists as builtin).
INSERT INTO gr33ncrops.crop_profiles (farm_id, crop_key, display_name, category, source, version, is_builtin)
SELECT NULL, v.crop_key, v.display_name, v.category, v.source, v.version, TRUE
FROM (VALUES
`)
	crops := c.CropsWithStages()
	for i, crop := range crops {
		source := crop.Source
		if source == "" {
			source = "Curated from data/crop_library.yaml"
		}
		comma := ","
		if i == len(crops)-1 {
			comma = ""
		}
		fmt.Fprintf(&b, "    (%s, %s, %s, %s, %d)%s\n",
			sqlQuote(crop.Key),
			sqlQuote(crop.DisplayName),
			sqlQuote(crop.Category),
			sqlQuote(source),
			c.Version,
			comma,
		)
	}
	b.WriteString(`) AS v(crop_key, display_name, category, source, version)
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profiles p
    WHERE p.farm_id IS NULL AND p.crop_key = v.crop_key AND p.is_builtin = TRUE
);

`)

	b.WriteString(`-- Stage rows (insert only when profile exists and stage missing).
INSERT INTO gr33ncrops.crop_profile_stages (
    crop_profile_id, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
    vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
    dli_target, photoperiod_hrs, notes
)
SELECT p.id, s.stage::gr33nfertigation.growth_stage_enum, s.ec_min, s.ec_target, s.ec_max, s.ph_min, s.ph_max,
       s.vpd_min_kpa, s.vpd_max_kpa, s.temp_min_c, s.temp_max_c, s.rh_min_pct, s.rh_max_pct,
       s.dli_target, s.photoperiod_hrs, s.notes
FROM gr33ncrops.crop_profiles p
JOIN (VALUES
`)

	rowN := 0
	totalRows := 0
	for _, crop := range crops {
		totalRows += len(crop.Stages)
	}
	for _, crop := range crops {
		for _, st := range crop.Stages {
			rowN++
			comma := ","
			if rowN == totalRows {
				comma = ""
			}
			fmt.Fprintf(&b, "    (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)%s\n",
				sqlQuote(crop.Key),
				sqlQuote(st.Stage),
				sqlNum(st.ECMin), sqlNum(st.ECTarget), sqlNum(st.ECMax),
				sqlNum(st.PHMin), sqlNum(st.PHMax),
				sqlNum(st.VPDMinKPa), sqlNum(st.VPDMaxKPa),
				sqlNum(st.TempMinC), sqlNum(st.TempMaxC),
				sqlNum(st.RHMinPct), sqlNum(st.RHMaxPct),
				sqlNum(st.DLITarget), sqlNum(st.PhotoperiodHrs),
				sqlQuoteNull(st.Notes),
				comma,
			)
		}
	}
	b.WriteString(`) AS s(crop_key, stage, ec_min, ec_target, ec_max, ph_min, ph_max,
         vpd_min_kpa, vpd_max_kpa, temp_min_c, temp_max_c, rh_min_pct, rh_max_pct,
         dli_target, photoperiod_hrs, notes)
  ON p.farm_id IS NULL AND p.is_builtin = TRUE AND p.crop_key = s.crop_key
WHERE NOT EXISTS (
    SELECT 1 FROM gr33ncrops.crop_profile_stages existing
    WHERE existing.crop_profile_id = p.id AND existing.stage = s.stage::gr33nfertigation.growth_stage_enum
);
`)
	return b.String()
}

func sqlQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func sqlQuoteNull(s string) string {
	if strings.TrimSpace(s) == "" {
		return "NULL"
	}
	return sqlQuote(s)
}

func sqlNum(v *float64) string {
	if v == nil {
		return "NULL"
	}
	// Match SQL seed formatting: integers without decimals when whole.
	if *v == float64(int64(*v)) {
		return fmt.Sprintf("%d", int64(*v))
	}
	return fmt.Sprintf("%.2f", *v)
}
