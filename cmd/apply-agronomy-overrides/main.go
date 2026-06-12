// Command apply-agronomy-overrides applies Phase 83 WS2 farm crop profile deltas.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/agronomyoverrides"
	"gr33n-api/internal/croplibrary"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/pgxutil"
)

func main() {
	var (
		farmID   = flag.Int64("farm-id", 0, "farm id (required)")
		filePath = flag.String("file", "", "override pack YAML (required)")
		dryRun   = flag.Bool("dry-run", false, "validate pack only")
	)
	flag.Parse()

	if *farmID <= 0 {
		log.Fatal("-farm-id is required")
	}
	if strings.TrimSpace(*filePath) == "" {
		log.Fatal("-file is required")
	}

	pack, err := croplibrary.LoadOverridePack(*filePath)
	if err != nil {
		log.Fatal(err)
	}
	if *dryRun {
		fmt.Printf("dry-run OK: %d override(s) in %s\n", len(pack.Overrides), *filePath)
		for _, o := range pack.Overrides {
			fmt.Printf("  - %s (%d stage delta(s))\n", o.CropKey, len(o.Stages))
		}
		return
	}

	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("parse DATABASE_URL: %v", err)
	}
	pgxutil.RegisterVectorTypes(cfg)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	n, err := agronomyoverrides.ApplyPack(ctx, db.New(pool), *farmID, pack)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("applied %d crop override(s) for farm_id=%d\n", n, *farmID)
}
