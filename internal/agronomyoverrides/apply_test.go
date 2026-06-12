package agronomyoverrides_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/agronomyoverrides"
	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/pgxutil"
)

func TestApplyPack_RejectsUnsupported(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	pgxutil.RegisterVectorTypes(cfg)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)

	pack := &croplibrary.OverridePack{
		Version: 1,
		Overrides: []croplibrary.CropOverride{{
			CropKey: "ramps",
			Stages:  []croplibrary.StageOverride{{Stage: "early_veg", ECMin: ptr(1.0)}},
		}},
	}
	_, err = agronomyoverrides.ApplyPack(ctx, q, 1, pack)
	if err == nil {
		t.Fatal("expected error applying override for unsupported ramps")
	}
}

func TestGetCropProfileByKey_FarmOverrideWins(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	pgxutil.RegisterVectorTypes(cfg)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)

	farmID := int64(1)
	ecMax := 1.4
	pack := &croplibrary.OverridePack{
		Version: 1,
		Source:  "integration test",
		Overrides: []croplibrary.CropOverride{{
			CropKey: "cannabis",
			Stages:  []croplibrary.StageOverride{{Stage: "late_flower", ECMax: &ecMax}},
		}},
	}
	if _, err := agronomyoverrides.ApplyPack(ctx, q, farmID, pack); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		fid := farmID
		_ = q.DeleteFarmCropProfileByKey(ctx, db.DeleteFarmCropProfileByKeyParams{FarmID: &fid, CropKey: "cannabis"})
	})

	got, err := q.GetCropProfileByKey(ctx, db.GetCropProfileByKeyParams{CropKey: "cannabis", FarmID: &farmID})
	if err != nil {
		t.Fatal(err)
	}
	if got.IsBuiltin || got.FarmID == nil || *got.FarmID != farmID {
		t.Fatalf("expected farm override profile, got builtin=%v farm_id=%v", got.IsBuiltin, got.FarmID)
	}
	st, err := q.GetCropProfileStage(ctx, db.GetCropProfileStageParams{
		CropProfileID: got.ID,
		Stage:         "late_flower",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !st.EcMax.Valid {
		t.Fatal("expected ec_max on override stage")
	}
}

func ptr(v float64) *float64 { return &v }
