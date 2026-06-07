// Phase 48 WS7 — dev seed idempotency smokes.
package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestPhase48WS7_SensorUniqueIndexPresent(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	err := testPool.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1 FROM pg_indexes
  WHERE schemaname = 'gr33ncore'
    AND tablename = 'sensors'
    AND indexname = 'uq_sensors_farm_name_active'
)`).Scan(&exists)
	if err != nil {
		t.Fatalf("index query: %v", err)
	}
	if !exists {
		t.Fatal("expected uq_sensors_farm_name_active index (run phase48 migration)")
	}
}

func TestPhase48WS7_ReSeedSensorCountStable(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	seedPath := filepath.Join(phase48RepoRoot(), "db", "seeds", "master_seed.sql")
	if _, err := os.Stat(seedPath); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	countSensors := func() int {
		var n int
		if err := testPool.QueryRow(ctx, `
SELECT count(*) FROM gr33ncore.sensors
WHERE farm_id = 1 AND deleted_at IS NULL`).Scan(&n); err != nil {
			t.Fatalf("count sensors: %v", err)
		}
		return n
	}

	before := countSensors()
	runSeed := func(label string) {
		t.Helper()
		cmd := exec.CommandContext(ctx, "psql", phase48DatabaseURL(), "-v", "ON_ERROR_STOP=1", "-f", seedPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s seed: %v\n%s", label, err, string(out))
		}
	}
	runSeed("first")
	afterFirst := countSensors()
	runSeed("second")
	afterSecond := countSensors()

	if afterFirst < before {
		t.Fatalf("sensor count dropped after first re-seed: before=%d after=%d", before, afterFirst)
	}
	if afterSecond != afterFirst {
		t.Fatalf("sensor count changed on second re-seed: first=%d second=%d", afterFirst, afterSecond)
	}
}

func TestPhase48WS7_DevSeedProfileOnFarm1(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var profile *string
	err := testPool.QueryRow(ctx, `
SELECT meta_data->>'dev_seed_profile'
FROM gr33ncore.farms WHERE id = 1`).Scan(&profile)
	if err != nil {
		t.Fatalf("farm meta: %v", err)
	}
	if profile == nil || *profile == "" {
		t.Skip("dev_seed_profile unset — run make seed after phase48 migration")
	}
}

func phase48RepoRoot() string {
	wd, _ := os.Getwd()
	// cmd/api tests run from module root or cmd/api — walk up to find db/seeds
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "db", "seeds", "master_seed.sql")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return wd
}

func phase48DatabaseURL() string {
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	if u := os.Getenv("TEST_DATABASE_URL"); u != "" {
		return u
	}
	return "postgres://gr33n:gr33n@localhost:5432/gr33n?sslmode=disable"
}
