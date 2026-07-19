package solar

import (
	"math"
	"testing"
	"time"
)

func TestSolarForDate_PortlandSummer(t *testing.T) {
	tz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("tz: %v", err)
	}
	date := time.Date(2026, 6, 8, 12, 0, 0, 0, tz)
	day := SolarForDate(45.52, -122.68, tz, date)

	if day.DaylengthHours < 14 || day.DaylengthHours > 16.5 {
		t.Fatalf("daylength = %.2f h, want ~15h summer Portland", day.DaylengthHours)
	}
	if !day.Sunrise.Before(day.SolarNoon) || !day.SolarNoon.Before(day.Sunset) {
		t.Fatalf("sunrise/noon/sunset order wrong: %v %v %v", day.Sunrise, day.SolarNoon, day.Sunset)
	}
	if day.ClearSkyDLI < 25 || day.ClearSkyDLI > 65 {
		t.Fatalf("clear-sky DLI = %.2f, want summer mid-lat range", day.ClearSkyDLI)
	}
	if day.MaxSunElevationDeg < 55 || day.MaxSunElevationDeg > 75 {
		t.Fatalf("max elevation = %.2f°, want high summer sun", day.MaxSunElevationDeg)
	}
}

// Regression: sunrise/sunset must land in the correct LOCAL clock hour, not just
// have the right ordering/duration. The bug this guards against anchored the
// UTC-relative solar-noon minutes to tz-local midnight instead of UTC midnight,
// which applied the tz offset twice (e.g. sunrise showed up ~4h late for a
// UTC-4 farm — 10:22am instead of ~6:22am for Ohio in July).
func TestSolarForDate_LocalClockHour(t *testing.T) {
	tz, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("tz: %v", err)
	}
	date := time.Date(2026, 7, 18, 12, 0, 0, 0, tz)
	day := SolarForDate(40.8938, -81.4055, tz, date)

	sunriseHour := day.Sunrise.In(tz).Hour()
	sunsetHour := day.Sunset.In(tz).Hour()
	if sunriseHour < 5 || sunriseHour > 7 {
		t.Fatalf("sunrise hour = %d (%v), want ~6am EDT for Ohio in July", sunriseHour, day.Sunrise.In(tz))
	}
	sunriseMin := day.Sunrise.In(tz).Hour()*60 + day.Sunrise.In(tz).Minute()
	// Almanac for ~40.89°N 81.41°W mid-July is ~6:05–6:15 AM EDT; allow ±12 min.
	if sunriseMin < 353 || sunriseMin > 387 {
		t.Fatalf("sunrise = %v (%d min), want ~6:05–6:25 AM EDT", day.Sunrise.In(tz).Format("3:04 PM"), sunriseMin)
	}
	if sunsetHour < 20 || sunsetHour > 22 {
		t.Fatalf("sunset hour = %d (%v), want ~9pm EDT for Ohio in July", sunsetHour, day.Sunset.In(tz))
	}
	if day.Sunset.In(tz).Day() != day.Sunrise.In(tz).Day() {
		t.Fatalf("sunset landed on a different day than sunrise: %v vs %v", day.Sunset.In(tz), day.Sunrise.In(tz))
	}
}

func TestSolarForDate_PolarNight(t *testing.T) {
	tz := time.UTC
	date := time.Date(2026, 12, 21, 12, 0, 0, 0, tz)
	day := SolarForDate(78.0, 15.0, tz, date) // high arctic winter
	if day.DaylengthHours > 0.5 {
		t.Fatalf("expected near-zero daylength in polar winter, got %.2f", day.DaylengthHours)
	}
}

func TestSupplementalDLIGap(t *testing.T) {
	gap := SupplementalDLIGap(35, 28, 1)
	if math.Abs(gap-7) > 0.01 {
		t.Fatalf("gap = %v want 7", gap)
	}
	if SupplementalDLIGap(20, 30, 1) != 0 {
		t.Fatal("no gap when natural exceeds target")
	}
}
