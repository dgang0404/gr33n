// Command rag-ingest embeds farm-scoped operational rows into gr33ncore.rag_embedding_chunks (Phase 24 WS3).
//
// Requires DATABASE_URL, pgvector-enabled Postgres, and EMBEDDING_API_KEY (plus optional EMBEDDING_BASE_URL / EMBEDDING_MODEL).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/pgxutil"
	"gr33n-api/internal/rag/embed"
	"gr33n-api/internal/rag/ingest"
)

func main() {
	var (
		farmID       = flag.Int64("farm-id", 0, "farm id (required)")
		doTasks      = flag.Bool("tasks", false, "index tasks")
		doRuns       = flag.Bool("automation-runs", false, "index automation_runs")
		doCycles     = flag.Bool("crop-cycles", false, "index gr33nfertigation.crop_cycles")
		batchRuns    = flag.Int("run-batch-size", 500, "cursor batch size for automation runs")
		startAfterID = flag.Int64("runs-after-id", 0, "only automation runs with id > this")
		dryRun       = flag.Bool("dry-run", false, "print counts only (no embeddings / DB writes)")
	)
	flag.Parse()

	if *farmID <= 0 {
		log.Fatal("-farm-id is required")
	}
	if !*doTasks && !*doRuns && !*doCycles {
		log.Fatal("specify at least one of -tasks, -automation-runs, or -crop-cycles")
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

	if *dryRun {
		q := db.New(pool)
		var nTasks, nRuns, nCycles int
		if *doTasks {
			tasks, err := q.ListTasksByFarm(ctx, *farmID)
			if err != nil {
				log.Fatal(err)
			}
			nTasks = len(tasks)
		}
		if *doRuns {
			runs, err := q.ListAutomationRunsByFarm(ctx, db.ListAutomationRunsByFarmParams{
				FarmID: *farmID,
				Limit:  1000000,
			})
			if err != nil {
				log.Fatal(err)
			}
			nRuns = len(runs)
		}
		if *doCycles {
			cycles, err := q.ListCropCyclesByFarm(ctx, *farmID)
			if err != nil {
				log.Fatal(err)
			}
			nCycles = len(cycles)
		}
		fmt.Printf("dry-run farm=%d tasks=%d automation_runs=%d crop_cycles=%d\n", *farmID, nTasks, nRuns, nCycles)
		return
	}

	emb, err := embed.NewOpenAICompatibleFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	w := &ingest.Worker{Q: db.New(pool), Embedder: emb}

	if *doTasks {
		n, err := w.IngestFarmTasks(ctx, *farmID)
		if err != nil {
			log.Fatalf("tasks: %v", err)
		}
		log.Printf("embedded tasks: %d", n)
	}
	if *doRuns {
		n, err := w.IngestFarmAutomationRuns(ctx, *farmID, int32(*batchRuns), *startAfterID)
		if err != nil {
			log.Fatalf("automation_runs: %v", err)
		}
		log.Printf("embedded automation_runs: %d", n)
	}
	if *doCycles {
		n, err := w.IngestFarmCropCycles(ctx, *farmID)
		if err != nil {
			log.Fatalf("crop_cycles: %v", err)
		}
		log.Printf("embedded crop_cycles: %d", n)
	}
}
