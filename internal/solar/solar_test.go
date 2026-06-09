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
