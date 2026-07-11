// Phase 152 WS2 — DB-backed coverage for ResolveCitationRoute's success
// paths. Skips gracefully when DATABASE_URL isn't reachable (e.g. a
// sandboxed CI lane with no Postgres), matching the smoke-test skip pattern
// used elsewhere in this repo.

package farmguardian

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://gr33n:gr33n@127.0.0.1:5433/gr33n?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Skipf("skipping — could not open DATABASE_URL: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("skipping — DATABASE_URL not reachable: %v", err)
	}
	return pool
}

// firstZoneForFarm picks an existing zone under the given farm so the test
// doesn't need to also stand up a farm/org fixture — every seeded dev DB
// used by guardian-qa (FARM_ID=1) already has zones.
func firstZoneForFarm(t *testing.T, ctx context.Context, q *db.Queries, farmID int64) int64 {
	t.Helper()
	zones, err := q.ListZonesByFarm(ctx, farmID)
	if err != nil || len(zones) == 0 {
		t.Skipf("skipping — no zones found for farm %d (run scripts/bootstrap-local.sh --seed first): %v", farmID, err)
	}
	return zones[0].ID
}

func TestResolveCitationRoute_cropCycle(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := t.Context()
	q := db.New(pool)
	const farmID = int64(1)
	zoneID := firstZoneForFarm(t, ctx, q, farmID)

	cycle, err := q.CreateCropCycle(ctx, db.CreateCropCycleParams{
		FarmID:    farmID,
		ZoneID:    zoneID,
		Name:      "Phase 152 WS2 route test cycle",
		IsActive:  false, // zone likely already has an active cycle (uq_active_crop_cycle)
		StartedAt: pgtype.Date{Time: time.Now(), Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateCropCycle: %v", err)
	}
	defer pool.Exec(context.Background(), "DELETE FROM gr33nfertigation.crop_cycles WHERE id = $1", cycle.ID)

	route, ok := ResolveCitationRoute(ctx, q, farmID, "crop_cycle", cycle.ID)
	if !ok {
		t.Fatal("expected route to resolve")
	}
	want := "/crop-cycles/" + strconv.FormatInt(cycle.ID, 10) + "/summary"
	if route != want {
		t.Fatalf("route = %q, want %q", route, want)
	}

	// Wrong farm must never resolve, even for a real row.
	if _, ok := ResolveCitationRoute(ctx, q, farmID+999, "crop_cycle", cycle.ID); ok {
		t.Fatal("expected cross-farm lookup to fail")
	}
}

func TestResolveCitationRoute_fertigationProgram(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := t.Context()
	q := db.New(pool)
	const farmID = int64(1)
	zoneID := firstZoneForFarm(t, ctx, q, farmID)

	prog, err := q.CreateProgram(ctx, db.CreateProgramParams{
		FarmID:       farmID,
		Name:         "Phase 152 WS2 route test program",
		TargetZoneID: &zoneID,
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("CreateProgram: %v", err)
	}
	defer pool.Exec(context.Background(), "DELETE FROM gr33nfertigation.programs WHERE id = $1", prog.ID)

	route, ok := ResolveCitationRoute(ctx, q, farmID, "fertigation_program", prog.ID)
	if !ok {
		t.Fatal("expected route to resolve")
	}
	want := "/zones/" + strconv.FormatInt(zoneID, 10) + "?tab=water"
	if route != want {
		t.Fatalf("route = %q, want %q", route, want)
	}
}

func TestResolveCitationRoute_task(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := t.Context()
	q := db.New(pool)
	const farmID = int64(1)
	zoneID := firstZoneForFarm(t, ctx, q, farmID)

	task, err := q.CreateTask(ctx, db.CreateTaskParams{
		FarmID: farmID,
		ZoneID: &zoneID,
		Title:  "Phase 152 WS2 route test task",
		Status: "todo",
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	defer pool.Exec(context.Background(), "DELETE FROM gr33ncore.tasks WHERE id = $1", task.ID)

	route, ok := ResolveCitationRoute(ctx, q, farmID, "task", task.ID)
	if !ok {
		t.Fatal("expected route to resolve")
	}
	want := "/zones/" + strconv.FormatInt(zoneID, 10)
	if route != want {
		t.Fatalf("route = %q, want %q", route, want)
	}
}

func TestResolveCitationRoute_taskWithoutZoneUnresolved(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := t.Context()
	q := db.New(pool)
	const farmID = int64(1)

	task, err := q.CreateTask(ctx, db.CreateTaskParams{
		FarmID: farmID,
		Title:  "Phase 152 WS2 route test task (no zone)",
		Status: "todo",
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	defer pool.Exec(context.Background(), "DELETE FROM gr33ncore.tasks WHERE id = $1", task.ID)

	if _, ok := ResolveCitationRoute(ctx, q, farmID, "task", task.ID); ok {
		t.Fatal("expected no route for a task without a zone")
	}
}
