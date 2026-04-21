// Command rag-ingest embeds farm-scoped operational rows into gr33ncore.rag_embedding_chunks (Phase 24 WS3).
//
// Requires DATABASE_URL, pgvector-enabled Postgres, and EMBEDDING_API_KEY (plus optional EMBEDDING_BASE_URL / EMBEDDING_MODEL).
// Incremental watermark: use -updated-after or env RAG_INGEST_UPDATED_AFTER (RFC3339 / RFC3339Nano); flag wins when both are set.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

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
		doPrograms   = flag.Bool("programs", false, "index gr33nfertigation.programs (metadata allowlisted)")
		doSchedules  = flag.Bool("schedules", false, "index gr33ncore.schedules")
		doRules      = flag.Bool("automation-rules", false, "index gr33ncore.automation_rules")
		doActions    = flag.Bool("executable-actions", false, "index gr33ncore.executable_actions (farm-linked; action_parameters scrubbed)")
		doCosts      = flag.Bool("cost-transactions", false, "index gr33ncore.cost_transactions (no amount/currency in text)")
		doInputs     = flag.Bool("inventory-definitions", false, "index gr33nnaturalfarming.input_definitions (no unit cost)")
		doBatches    = flag.Bool("inventory-batches", false, "index gr33nnaturalfarming.input_batches (no qty / cost numerics)")
		doAlerts     = flag.Bool("alerts", false, "index gr33ncore.alerts_notifications")
		batchRuns    = flag.Int("run-batch-size", 500, "cursor batch size for automation runs")
		startAfterID = flag.Int64("runs-after-id", 0, "only automation runs with id > this")
		batchCosts   = flag.Int("cost-batch-size", 500, "cursor batch size for cost transactions")
		costAfterID  = flag.Int64("cost-after-id", 0, "only cost_transactions with id > this")
		batchAlerts  = flag.Int("alert-batch-size", 500, "cursor batch size for alerts")
		alertAfterID = flag.Int64("alert-after-id", 0, "only alerts_notifications with id > this")
		dryRun       = flag.Bool("dry-run", false, "print counts only (no embeddings / DB writes)")
		updatedAfter = flag.String("updated-after", "", "incremental ingest: RFC3339 or RFC3339Nano timestamp; embed only rows changed after this time (tasks/cycles/programs/schedules/rules/inventory use updated_at; automation_runs use executed_at; costs use updated_at; alerts use created_at). Ignores *-after-id cursor flags when set. executable_actions still re-indexes all farm-linked rows (table has no updated_at). Empty uses env RAG_INGEST_UPDATED_AFTER if set; otherwise full ingest.")
	)
	flag.Parse()

	if *farmID <= 0 {
		log.Fatal("-farm-id is required")
	}
	if !*doTasks && !*doRuns && !*doCycles && !*doPrograms && !*doSchedules && !*doRules && !*doActions && !*doCosts && !*doInputs && !*doBatches && !*doAlerts {
		log.Fatal("specify at least one ingest flag (see -help)")
	}

	updatedAfterEffective := *updatedAfter
	if updatedAfterEffective == "" {
		updatedAfterEffective = os.Getenv("RAG_INGEST_UPDATED_AFTER")
	}

	since, err := parseUpdatedAfter(updatedAfterEffective)
	if err != nil {
		log.Fatalf("updated-after: %v", err)
	}
	if since != nil && (*startAfterID != 0 || *costAfterID != 0 || *alertAfterID != 0) {
		log.Printf("warning: updated-after is set; runs-after-id, cost-after-id, and alert-after-id are ignored for their respective domains")
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
		var nTasks, nRuns, nCycles, nPrograms, nSchedules, nRules, nActions int
		var nCosts, nAlerts int64
		var nInputs, nBatches int
		if *doTasks {
			if since != nil {
				tasks, err := q.ListTasksByFarmUpdatedAfter(ctx, db.ListTasksByFarmUpdatedAfterParams{FarmID: *farmID, UpdatedAfter: *since})
				if err != nil {
					log.Fatal(err)
				}
				nTasks = len(tasks)
			} else {
				tasks, err := q.ListTasksByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nTasks = len(tasks)
			}
		}
		if *doRuns {
			if since != nil {
				cnt, err := q.CountAutomationRunsByFarmExecutedAfter(ctx, db.CountAutomationRunsByFarmExecutedAfterParams{FarmID: *farmID, Since: *since})
				if err != nil {
					log.Fatal(err)
				}
				nRuns = int(cnt)
			} else {
				runs, err := q.ListAutomationRunsByFarm(ctx, db.ListAutomationRunsByFarmParams{
					FarmID: *farmID,
					Limit:  1000000,
				})
				if err != nil {
					log.Fatal(err)
				}
				nRuns = len(runs)
			}
		}
		if *doCycles {
			if since != nil {
				cycles, err := q.ListCropCyclesByFarmUpdatedAfter(ctx, db.ListCropCyclesByFarmUpdatedAfterParams{FarmID: *farmID, UpdatedAfter: *since})
				if err != nil {
					log.Fatal(err)
				}
				nCycles = len(cycles)
			} else {
				cycles, err := q.ListCropCyclesByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nCycles = len(cycles)
			}
		}
		if *doPrograms {
			if since != nil {
				progs, err := q.ListProgramsByFarmUpdatedAfter(ctx, db.ListProgramsByFarmUpdatedAfterParams{FarmID: *farmID, UpdatedAfter: *since})
				if err != nil {
					log.Fatal(err)
				}
				nPrograms = len(progs)
			} else {
				progs, err := q.ListProgramsByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nPrograms = len(progs)
			}
		}
		if *doSchedules {
			if since != nil {
				sch, err := q.ListSchedulesByFarmUpdatedAfter(ctx, db.ListSchedulesByFarmUpdatedAfterParams{FarmID: *farmID, UpdatedAfter: *since})
				if err != nil {
					log.Fatal(err)
				}
				nSchedules = len(sch)
			} else {
				sch, err := q.ListSchedulesByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nSchedules = len(sch)
			}
		}
		if *doRules {
			if since != nil {
				rules, err := q.ListAutomationRulesByFarmUpdatedAfter(ctx, db.ListAutomationRulesByFarmUpdatedAfterParams{FarmID: *farmID, UpdatedAfter: *since})
				if err != nil {
					log.Fatal(err)
				}
				nRules = len(rules)
			} else {
				rules, err := q.ListAutomationRulesByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nRules = len(rules)
			}
		}
		if *doActions {
			acts, err := q.ListExecutableActionsByFarmForRAG(ctx, *farmID)
			if err != nil {
				log.Fatal(err)
			}
			nActions = len(acts)
		}
		if *doCosts {
			if since != nil {
				cnt, err := q.CountCostTransactionsByFarmUpdatedAfter(ctx, db.CountCostTransactionsByFarmUpdatedAfterParams{FarmID: *farmID, Since: *since})
				if err != nil {
					log.Fatal(err)
				}
				nCosts = cnt
			} else {
				cnt, err := q.CountCostTransactionsByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nCosts = cnt
			}
		}
		if *doInputs {
			if since != nil {
				defs, err := q.ListInputDefinitionsByFarmUpdatedAfter(ctx, db.ListInputDefinitionsByFarmUpdatedAfterParams{FarmID: *farmID, UpdatedAfter: *since})
				if err != nil {
					log.Fatal(err)
				}
				nInputs = len(defs)
			} else {
				defs, err := q.ListInputDefinitionsByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nInputs = len(defs)
			}
		}
		if *doBatches {
			if since != nil {
				bat, err := q.ListInputBatchesByFarmUpdatedAfter(ctx, db.ListInputBatchesByFarmUpdatedAfterParams{FarmID: *farmID, UpdatedAfter: *since})
				if err != nil {
					log.Fatal(err)
				}
				nBatches = len(bat)
			} else {
				bat, err := q.ListInputBatchesByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nBatches = len(bat)
			}
		}
		if *doAlerts {
			if since != nil {
				cnt, err := q.CountAlertsByFarmCreatedAfter(ctx, db.CountAlertsByFarmCreatedAfterParams{FarmID: *farmID, Since: *since})
				if err != nil {
					log.Fatal(err)
				}
				nAlerts = cnt
			} else {
				cnt, err := q.CountAlertsByFarm(ctx, *farmID)
				if err != nil {
					log.Fatal(err)
				}
				nAlerts = cnt
			}
		}
		fmt.Printf("dry-run farm=%d tasks=%d automation_runs=%d crop_cycles=%d programs=%d schedules=%d automation_rules=%d executable_actions=%d cost_transactions=%d input_definitions=%d input_batches=%d alerts=%d\n",
			*farmID, nTasks, nRuns, nCycles, nPrograms, nSchedules, nRules, nActions, nCosts, nInputs, nBatches, nAlerts)
		return
	}

	emb, err := embed.NewOpenAICompatibleFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	w := &ingest.Worker{Q: db.New(pool), Embedder: emb}

	if *doTasks {
		n, err := w.IngestFarmTasks(ctx, *farmID, since)
		if err != nil {
			log.Fatalf("tasks: %v", err)
		}
		log.Printf("embedded tasks: %d", n)
	}
	if *doRuns {
		n, err := w.IngestFarmAutomationRuns(ctx, *farmID, int32(*batchRuns), *startAfterID, since)
		if err != nil {
			log.Fatalf("automation_runs: %v", err)
		}
		log.Printf("embedded automation_runs: %d", n)
	}
	if *doCycles {
		n, err := w.IngestFarmCropCycles(ctx, *farmID, since)
		if err != nil {
			log.Fatalf("crop_cycles: %v", err)
		}
		log.Printf("embedded crop_cycles: %d", n)
	}
	if *doPrograms {
		n, err := w.IngestFarmFertigationPrograms(ctx, *farmID, since)
		if err != nil {
			log.Fatalf("programs: %v", err)
		}
		log.Printf("embedded programs: %d", n)
	}
	if *doSchedules {
		n, err := w.IngestFarmSchedules(ctx, *farmID, since)
		if err != nil {
			log.Fatalf("schedules: %v", err)
		}
		log.Printf("embedded schedules: %d", n)
	}
	if *doRules {
		n, err := w.IngestFarmAutomationRules(ctx, *farmID, since)
		if err != nil {
			log.Fatalf("automation_rules: %v", err)
		}
		log.Printf("embedded automation_rules: %d", n)
	}
	if *doActions {
		if since != nil {
			log.Printf("executable_actions: incremental watermark not supported (no updated_at); embedding all farm-linked rows")
		}
		n, err := w.IngestFarmExecutableActions(ctx, *farmID)
		if err != nil {
			log.Fatalf("executable_actions: %v", err)
		}
		log.Printf("embedded executable_actions: %d", n)
	}
	if *doCosts {
		n, err := w.IngestFarmCostTransactions(ctx, *farmID, int32(*batchCosts), *costAfterID, since)
		if err != nil {
			log.Fatalf("cost_transactions: %v", err)
		}
		log.Printf("embedded cost_transactions: %d", n)
	}
	if *doInputs {
		n, err := w.IngestFarmInputDefinitions(ctx, *farmID, since)
		if err != nil {
			log.Fatalf("input_definitions: %v", err)
		}
		log.Printf("embedded input_definitions: %d", n)
	}
	if *doBatches {
		n, err := w.IngestFarmInputBatches(ctx, *farmID, since)
		if err != nil {
			log.Fatalf("input_batches: %v", err)
		}
		log.Printf("embedded input_batches: %d", n)
	}
	if *doAlerts {
		n, err := w.IngestFarmAlertNotifications(ctx, *farmID, int32(*batchAlerts), *alertAfterID, since)
		if err != nil {
			log.Fatalf("alerts: %v", err)
		}
		log.Printf("embedded alerts_notifications: %d", n)
	}
}

func parseUpdatedAfter(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, fmt.Errorf("parse RFC3339 / RFC3339Nano: %w", err)
		}
	}
	return &t, nil
}
