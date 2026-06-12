package croplibrary_test

import (
	"context"
	"os"
	"testing"

	"gr33n-api/internal/croplibrary"
	"gr33n-api/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestLoadCatalogFromDB_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	q := db.New(pool)
	n, err := q.CountCropCatalogEntries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n < 50 {
		t.Fatalf("want >= 50 catalog entries, got %d", n)
	}

	cat, err := croplibrary.LoadCatalogFromDB(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Crops) < 46 {
		t.Fatalf("want >= 46 supported crops, got %d", len(cat.Crops))
	}
	if len(cat.Unsupported) < 4 {
		t.Fatalf("want >= 4 unsupported, got %d", len(cat.Unsupported))
	}
	m, ok := cat.Aliases["aubergine"]
	if !ok || m != "eggplant" {
		t.Fatalf("aubergine alias: %q ok=%v", m, ok)
	}

	reg := croplibrary.NewRegistry(cat)
	rm, ok := reg.ResolveTerm("ramps")
	if !ok || rm.Kind != croplibrary.MentionUnsupported {
		t.Fatalf("ramps: %+v ok=%v", rm, ok)
	}
}

func TestDefaultCatalog_DBMode_Integration(t *testing.T) {
	t.Setenv("CROP_CATALOG_SOURCE", "db")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	croplibrary.SetRuntimeCatalogQuerier(db.New(pool))
	cat, err := croplibrary.DefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Crops) < 46 {
		t.Fatalf("want >= 46 crops from DefaultCatalog db, got %d", len(cat.Crops))
	}
}
