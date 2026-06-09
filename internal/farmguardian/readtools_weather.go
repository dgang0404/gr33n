// Phase 66 WS5 — site weather read tool (offline solar + local readings).

package farmguardian

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/solar"
)

// SiteWeatherPersonaRule guides Guardian weather answers.
const SiteWeatherPersonaRule = `Site weather (Phase 66): Use site_weather for daylight, DLI, frost, supplemental-light gaps. State tier: solar_math (offline), local sensor/manual, or online forecast — never invent forecast when offline.`

var siteWeatherIntent = regexp.MustCompile(`(?i)\b(daylight|daylength|sunrise|sunset|solar noon|frost|supplemental light|extra light|dli|photoperiod|vent the greenhouse|outdoor temp|site weather|weather today)\b|\bhow long is (the )?day\b|\bneed (supplemental|extra) light\b`)

func shouldRunSiteWeatherReadIntent(question string) bool {
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	return siteWeatherIntent.MatchString(q)
}

func renderSiteWeather(ctx context.Context, q db.Querier, farmID int64) (string, error) {
	site, err := q.GetFarmSiteCoords(ctx, farmID)
	if err != nil {
		return "", err
	}
	tzName := strings.TrimSpace(site.Timezone)
	if tzName == "" {
		tzName = "UTC"
	}
	tz, err := time.LoadLocation(tzName)
	if err != nil {
		tz = time.UTC
	}
	now := time.Now().In(tz)

	var b strings.Builder
	b.WriteString("site_weather — ")
	farmLabel := fmt.Sprintf("farm #%d", farmID)
	if farm, ferr := q.GetFarmByID(ctx, farmID); ferr == nil && strings.TrimSpace(farm.Name) != "" {
		farmLabel = strings.TrimSpace(farm.Name)
	}
	b.WriteString(farmLabel)
	b.WriteString("\nTiers used:")

	lat, latOK := ifaceCoordFloat(site.Latitude)
	lng, lngOK := ifaceCoordFloat(site.Longitude)
	if latOK && lngOK {
		day := solar.SolarForDate(lat, lng, tz, now)
		b.WriteString(" solar_math")
		b.WriteString(fmt.Sprintf("\nSolar (offline math, lat %.4f lng %.4f):", lat, lng))
		b.WriteString(fmt.Sprintf("\n- Date: %s", now.Format("2006-01-02")))
		if day.DaylengthHours > 0 {
			b.WriteString(fmt.Sprintf("\n- Sunrise: %s", day.Sunrise.Format("15:04")))
			b.WriteString(fmt.Sprintf("\n- Sunset: %s", day.Sunset.Format("15:04")))
			b.WriteString(fmt.Sprintf("\n- Daylength: %.1f h", day.DaylengthHours))
			b.WriteString(fmt.Sprintf("\n- Clear-sky DLI: %.1f mol/m²/day", day.ClearSkyDLI))
		} else {
			b.WriteString("\n- Polar night / no sunrise today at this latitude")
		}

		if gap, target, ok := supplementalGapForFarm(ctx, q, farmID, day.ClearSkyDLI, 1.0); ok {
			b.WriteString(fmt.Sprintf("\n- Crop DLI target: %.1f mol/m²/day", target))
			if gap > 0.5 {
				b.WriteString(fmt.Sprintf("\n- Supplemental light gap (clear sky): ~%.1f mol/m²/day — add grow lights or extend photoperiod", gap))
			} else {
				b.WriteString("\n- Natural clear-sky DLI meets or exceeds active crop target")
			}
		}
	} else {
		b.WriteString(" none (set farm site coordinates in Settings)")
	}

	latest, err := q.GetLatestWeatherForFarm(ctx, farmID)
	if err == nil {
		b.WriteString("\nLatest reading:")
		b.WriteString(fmt.Sprintf("\n- Source: %s at %s", latest.DataSource, latest.RecordedAt.Format(time.RFC3339)))
		if t := numericToFloatPtr(latest.TemperatureCelsius); t != nil {
			b.WriteString(fmt.Sprintf("\n- Outdoor temp: %.1f °C", *t))
		}
		if h := numericToFloatPtr(latest.HumidityPercent); h != nil {
			b.WriteString(fmt.Sprintf("\n- Outdoor RH: %.0f%%", *h))
		}
		if c := numericToFloatPtr(latest.CloudCoverPercent); c != nil {
			b.WriteString(fmt.Sprintf("\n- Cloud cover: %.0f%%", *c))
			if latOK && lngOK {
				day := solar.SolarForDate(lat, lng, tz, now)
				cloudFactor := 1.0 - (*c / 100.0 * 0.65)
				if gap, target, ok := supplementalGapForFarm(ctx, q, farmID, day.ClearSkyDLI, cloudFactor); ok && gap > 0.5 {
					b.WriteString(fmt.Sprintf("\n- Supplemental light gap (cloud-adjusted): ~%.1f mol/m²/day (target %.1f)", gap, target))
				}
			}
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	b.WriteString("\nPersona: cite tier — solar math needs no internet; sensor/manual is LAN; forecast is optional opt-in.")
	return b.String(), nil
}

func supplementalGapForFarm(ctx context.Context, q db.Querier, farmID int64, clearSkyDLI, cloudFactor float64) (gap, target float64, ok bool) {
	cycles, err := q.ListCropCyclesByFarm(ctx, farmID)
	if err != nil {
		return 0, 0, false
	}
	var cycle db.Gr33nfertigationCropCycle
	for _, c := range cycles {
		if c.IsActive {
			cycle = c
			break
		}
	}
	if cycle.ID == 0 {
		return 0, 0, false
	}
	profileID, stage, _, err := resolveCropProfileContext(ctx, q, farmID, "", &ContextRef{
		CropCycleID: cycle.ID,
	})
	if err != nil || profileID <= 0 || stage == "" {
		return 0, 0, false
	}
	st, err := q.GetCropProfileStage(ctx, db.GetCropProfileStageParams{
		CropProfileID: profileID,
		Stage:         stage,
	})
	if err != nil || !st.DliTarget.Valid {
		return 0, 0, false
	}
	f, err := st.DliTarget.Float64Value()
	if err != nil || !f.Valid || f.Float64 <= 0 {
		return 0, 0, false
	}
	target = f.Float64
	gap = solar.SupplementalDLIGap(target, clearSkyDLI, cloudFactor)
	return gap, target, true
}

func ifaceCoordFloat(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	default:
		return 0, false
	}
}

func numericToFloatPtr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}
