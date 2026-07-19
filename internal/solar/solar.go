// Package solar computes sun position and clear-sky daylight metrics offline (Phase 66).
// Pure arithmetic — no network calls.
package solar

import (
	"math"
	"time"
)

// SolarDay holds sunrise/sunset and clear-sky light integrals for one calendar day.
type SolarDay struct {
	Sunrise            time.Time
	Sunset             time.Time
	SolarNoon          time.Time
	DaylengthHours     float64
	ClearSkyDLI        float64 // mol/m²/day, theoretical clear-sky PAR integral
	MaxSunElevationDeg float64
}

// SolarForDate returns solar times and clear-sky DLI for lat/lng on the given local calendar date.
func SolarForDate(lat, lng float64, tz *time.Location, date time.Time) SolarDay {
	if tz == nil {
		tz = time.UTC
	}
	year, month, dom := date.In(tz).Date()
	jd := julianDay(year, int(month), dom)

	// NOAA solar position (radians unless noted).
	n := jd - 2451545.0
	L := math.Mod(280.460+0.9856474*n, 360)
	g := math.Mod(357.528+0.9856003*n, 360)
	lambda := L + 1.915*math.Sin(deg2rad(g)) + 0.020*math.Sin(deg2rad(2*g))
	epsilon := 23.439 - 0.0000004*n
	sinDec := math.Sin(deg2rad(lambda)) * math.Sin(deg2rad(epsilon))
	dec := rad2deg(math.Asin(sinDec))

	eqTime := equationOfTimeMinutes(n)

	// Solar noon in minutes from local midnight (approx).
	solarNoonMin := 720 - 4*lng - eqTime
	hourAngle := rad2deg(math.Acos(
		(math.Sin(deg2rad(-0.833)) - math.Sin(deg2rad(lat))*math.Sin(deg2rad(dec))) /
			(math.Cos(deg2rad(lat)) * math.Cos(deg2rad(dec))),
	))

	// solarNoonMin etc. are minutes past *UTC* midnight (the formula above has no
	// timezone term, only longitude) — anchor to UTC midnight, not tz midnight, or
	// the tz offset gets applied twice (e.g. sunrise lands ~4h late for a UTC-4 farm).
	midnightUTC := time.Date(year, month, dom, 0, 0, 0, 0, time.UTC)

	var out SolarDay
	if hourAngle >= 0 && hourAngle < 180 {
		sunriseMin := solarNoonMin - 4*hourAngle
		sunsetMin := solarNoonMin + 4*hourAngle
		out.Sunrise = midnightUTC.Add(time.Duration(sunriseMin) * time.Minute).In(tz)
		out.Sunset = midnightUTC.Add(time.Duration(sunsetMin) * time.Minute).In(tz)
		out.SolarNoon = midnightUTC.Add(time.Duration(solarNoonMin) * time.Minute).In(tz)
		out.DaylengthHours = (sunsetMin - sunriseMin) / 60.0
	}

	// Max elevation at solar noon.
	elev := 90 - math.Abs(lat-dec)
	if elev < 0 {
		elev = 0
	}
	out.MaxSunElevationDeg = elev

	// Clear-sky DLI: integrate approximate PPFD curve (µmol/m²/s) over daylight.
	// Peak PPFD ~ 2000*sin(elev) at solar noon; trapezoid over daylight hours.
	if out.DaylengthHours > 0 && elev > 0 {
		peakPPFD := 2000.0 * math.Sin(deg2rad(elev))
		avgPPFD := peakPPFD * 0.55 // rough clear-sky daily average vs peak
		seconds := out.DaylengthHours * 3600.0
		out.ClearSkyDLI = avgPPFD * seconds / 1_000_000.0
	}

	return out
}

// SupplementalDLIGap returns how much DLI (mol/m²/day) supplemental light must add to reach cropTarget.
func SupplementalDLIGap(cropTarget, naturalClearSkyDLI, cloudFactor float64) float64 {
	if cropTarget <= 0 || naturalClearSkyDLI <= 0 {
		return 0
	}
	if cloudFactor <= 0 {
		cloudFactor = 1
	}
	effective := naturalClearSkyDLI * cloudFactor
	gap := cropTarget - effective
	if gap < 0 {
		return 0
	}
	return gap
}

// equationOfTimeMinutes — NOAA approximate equation of time (minutes).
// Replaces the old atan2 shortcut, which drifted ~15–20 min vs almanac times.
func equationOfTimeMinutes(n float64) float64 {
	b := deg2rad(360.0 / 365.0 * (n - 81.0))
	return 9.87*math.Sin(2*b) - 7.53*math.Cos(b) - 1.5*math.Sin(b)
}

func julianDay(year, month, day int) float64 {
	if month <= 2 {
		year--
		month += 12
	}
	a := math.Floor(float64(year) / 100.0)
	b := 2 - a + math.Floor(a/4)
	return math.Floor(365.25*(float64(year)+4716)) + math.Floor(30.6001*float64(month+1)) + float64(day) + b - 1524.5
}

func deg2rad(d float64) float64 { return d * math.Pi / 180 }
func rad2deg(r float64) float64 { return r * 180 / math.Pi }
