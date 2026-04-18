package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	automationworker "gr33n-api/internal/automation"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/filestorage"
)

const (
	smokeDevEmail    = "dev@gr33n.local"
	smokeDevPass     = "devpassword"
	smokeDevUserUUID = "00000000-0000-0000-0000-000000000001"
)

func (n *recordingNotifier) DispatchFarmAlert(_ context.Context, a db.Gr33ncoreAlertsNotification) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.alerts = append(n.alerts, a)
}

// countForSource returns how many dispatched alerts reference the given
// rule id (via triggering_event_source_id with source type automation_rule).

func (n *recordingNotifier) countForRule(ruleID int64) int {
	n.mu.Lock()
	defer n.mu.Unlock()
	count := 0
	for _, a := range n.alerts {
		if a.TriggeringEventSourceType != nil && *a.TriggeringEventSourceType == "automation_rule" &&
			a.TriggeringEventSourceID != nil && *a.TriggeringEventSourceID == ruleID {
			count++
		}
	}
	return count
}

func bootstrapSmokeAuth(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	hash, err := bcrypt.GenerateFromPassword([]byte(smokeDevPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `UPDATE auth.users SET password_hash = $1 WHERE id = $2`, hash, smokeDevUserUUID); err != nil {
		return err
	}
	uid := uuid.MustParse(smokeDevUserUUID)
	if _, err := pool.Exec(ctx, `
INSERT INTO gr33ncore.farm_memberships (farm_id, user_id, role_in_farm, permissions, joined_at)
VALUES ($1, $2, 'owner', '{}'::jsonb, NOW())
ON CONFLICT (farm_id, user_id) DO NOTHING`, int64(1), uid); err != nil {
		return err
	}
	// Phase 11 columns / enum values (no-op if already applied)
	if _, err := pool.Exec(ctx, `
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e
    JOIN pg_type t ON e.enumtypid = t.oid
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname = 'gr33ncore' AND t.typname = 'farm_member_role_enum' AND e.enumlabel = 'operator'
  ) THEN
    ALTER TYPE gr33ncore.farm_member_role_enum ADD VALUE 'operator';
  END IF;
END $$;
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e
    JOIN pg_type t ON e.enumtypid = t.oid
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname = 'gr33ncore' AND t.typname = 'farm_member_role_enum' AND e.enumlabel = 'finance'
  ) THEN
    ALTER TYPE gr33ncore.farm_member_role_enum ADD VALUE 'finance';
  END IF;
END $$;
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS insert_commons_opt_in BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS insert_commons_last_sync_at TIMESTAMPTZ;
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS insert_commons_last_attempt_at TIMESTAMPTZ;
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS insert_commons_last_delivery_status TEXT;
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS insert_commons_last_error TEXT;
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS insert_commons_backoff_until TIMESTAMPTZ;
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS insert_commons_consecutive_failures INT NOT NULL DEFAULT 0;
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS insert_commons_require_approval BOOLEAN NOT NULL DEFAULT FALSE;
CREATE TABLE IF NOT EXISTS gr33ncore.insert_commons_sync_events (
    id               BIGSERIAL PRIMARY KEY,
    farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    idempotency_key  TEXT,
    status           TEXT NOT NULL,
    http_status      INT,
    error            TEXT,
    payload          JSONB NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_insert_commons_sync_farm_idem
    ON gr33ncore.insert_commons_sync_events (farm_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_insert_commons_sync_farm_created
    ON gr33ncore.insert_commons_sync_events (farm_id, created_at DESC);
CREATE TABLE IF NOT EXISTS gr33ncore.insert_commons_bundles (
    id                  BIGSERIAL PRIMARY KEY,
    farm_id             BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    idempotency_key     TEXT,
    payload_hash        TEXT        NOT NULL,
    payload             JSONB       NOT NULL,
    status              TEXT        NOT NULL CHECK (status IN (
        'pending_approval', 'approved', 'rejected', 'delivered', 'delivery_failed'
    )),
    reviewer_user_id    UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    review_note         TEXT,
    delivery_http_status INT,
    delivery_error      TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_insert_commons_bundles_farm_hash
    ON gr33ncore.insert_commons_bundles (farm_id, payload_hash);
CREATE INDEX IF NOT EXISTS idx_insert_commons_bundles_farm_status_created
    ON gr33ncore.insert_commons_bundles (farm_id, status, created_at DESC);
ALTER TABLE gr33ncore.insert_commons_sync_events ADD COLUMN IF NOT EXISTS bundle_id BIGINT REFERENCES gr33ncore.insert_commons_bundles(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_insert_commons_sync_bundle
    ON gr33ncore.insert_commons_sync_events (bundle_id)
    WHERE bundle_id IS NOT NULL;
CREATE TABLE IF NOT EXISTS gr33ncore.farm_finance_account_mappings (
    id            BIGSERIAL PRIMARY KEY,
    farm_id       BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    cost_category gr33ncore.cost_category_enum NOT NULL,
    account_code  TEXT   NOT NULL,
    account_name  TEXT   NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE (farm_id, cost_category)
);
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger
    WHERE tgname = 'trg_farm_finance_account_mappings_updated_at'
  ) THEN
    CREATE TRIGGER trg_farm_finance_account_mappings_updated_at
      BEFORE UPDATE ON gr33ncore.farm_finance_account_mappings
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;
ALTER TABLE gr33ncore.cost_transactions ADD COLUMN IF NOT EXISTS document_type TEXT;
ALTER TABLE gr33ncore.cost_transactions ADD COLUMN IF NOT EXISTS document_reference TEXT;
ALTER TABLE gr33ncore.cost_transactions ADD COLUMN IF NOT EXISTS counterparty TEXT;
CREATE TABLE IF NOT EXISTS gr33ncore.organizations (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT        NOT NULL,
    plan_tier       TEXT        NOT NULL DEFAULT 'pilot',
    billing_status  TEXT        NOT NULL DEFAULT 'none',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_organizations_updated_at') THEN
    CREATE TRIGGER trg_organizations_updated_at
      BEFORE UPDATE ON gr33ncore.organizations
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;
CREATE TABLE IF NOT EXISTS gr33ncore.organization_memberships (
    organization_id BIGINT NOT NULL REFERENCES gr33ncore.organizations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE CASCADE,
    role_in_org     TEXT   NOT NULL CHECK (role_in_org IN ('owner', 'admin', 'member')),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_org_memberships_user_smoke
    ON gr33ncore.organization_memberships (user_id);
ALTER TABLE gr33ncore.farms ADD COLUMN IF NOT EXISTS organization_id BIGINT
    REFERENCES gr33ncore.organizations(id) ON DELETE SET NULL;
`); err != nil {
		return err
	}
	bootstrapSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260423_farm_bootstrap_templates.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(bootstrapSQL)); err != nil {
		return err
	}
	orgBootstrapSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260424_organization_default_bootstrap_template.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(orgBootstrapSQL)); err != nil {
		return err
	}
	commonsCatalogSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260426_commons_catalog.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(commonsCatalogSQL)); err != nil {
		return err
	}
	pushTokSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260427_user_push_tokens.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(pushTokSQL)); err != nil {
		return err
	}
	domainModsSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260428_phase14_domain_module_stubs.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(domainModsSQL)); err != nil {
		return err
	}
	bootstrapFertSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260429_bootstrap_fertigation_inventory_tasks.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(bootstrapFertSQL)); err != nil {
		return err
	}
	alertDurSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260430_phase19_alert_duration_cooldown.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(alertDurSQL)); err != nil {
		return err
	}
	sourceAlertSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260501_phase19_task_source_alert.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(sourceAlertSQL)); err != nil {
		return err
	}
	precondSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260502_phase19_schedule_preconditions.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(precondSQL)); err != nil {
		return err
	}
	sourceRuleSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260503_phase20_task_source_rule.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(sourceRuleSQL)); err != nil {
		return err
	}
	phase205SQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260504_phase205_husbandry_climate_bootstraps.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase205SQL)); err != nil {
		return err
	}
	phase2095LaborSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260505_phase2095_labor_schema.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase2095LaborSQL)); err != nil {
		return err
	}
	phase2095CostEnergySQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260506_phase2095_cost_energy_columns.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase2095CostEnergySQL)); err != nil {
		return err
	}
	phase2095ExecActProgramSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260507_phase2095_executable_actions_program_id.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase2095ExecActProgramSQL)); err != nil {
		return err
	}
	phase2095AnimalAquaSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260508_phase2095_animal_aquaponics_scope.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase2095AnimalAquaSQL)); err != nil {
		return err
	}
	phase206SetpointsSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260509_phase206_zone_setpoints.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase206SetpointsSQL)); err != nil {
		return err
	}
	phase207TaskConsumptionsSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260510_phase207_task_input_consumptions.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase207TaskConsumptionsSQL)); err != nil {
		return err
	}
	phase208AnimalHusbandrySQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260512_phase208_animal_husbandry.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase208AnimalHusbandrySQL)); err != nil {
		return err
	}
	phase208BootstrapUpgradeSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260513_phase208_bootstrap_upgrade.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase208BootstrapUpgradeSQL)); err != nil {
		return err
	}
	phase209LaborAutocostSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260514_phase209_labor_autocost.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase209LaborAutocostSQL)); err != nil {
		return err
	}
	phase209ProgramBackfillSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260515_phase209_program_actions_backfill.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase209ProgramBackfillSQL)); err != nil {
		return err
	}
	phase22ProgramRunsSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260516_phase22_program_runs.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase22ProgramRunsSQL)); err != nil {
		return err
	}
	phase22BackfillSweepSQL, err := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "20260517_phase22_program_actions_backfill_sweep.sql"))
	if err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, string(phase22BackfillSweepSQL)); err != nil {
		return err
	}
	return nil
}

// smokeAbortWithoutDB prints why integration tests did not run, then exits.
// Outside CI this exits 0 so `go test ./...` stays usable on laptops without Postgres.
// In CI (GITHUB_ACTIONS or CI=true) we exit 1 so a missing DB service cannot look green.
func smokeAbortWithoutDB(context string, err error) {
	fmt.Fprintf(os.Stderr, "smoke_test: %s", context)
	if err != nil {
		fmt.Fprintf(os.Stderr, ": %v", err)
	}
	fmt.Fprintf(os.Stderr, "\n  hint: set DATABASE_URL to a migrated database; optional seed: psql \"$DATABASE_URL\" -f db/seeds/master_seed.sql\n")
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://davidg@/gr33n?host=/var/run/postgresql"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		smokeAbortWithoutDB("could not open DATABASE_URL (using default if env unset)", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		smokeAbortWithoutDB("database ping failed", err)
	}
	if err := bootstrapSmokeAuth(pool); err != nil {
		pool.Close()
		fmt.Fprintf(os.Stderr, "smoke_test bootstrap: %v\n", err)
		os.Exit(1)
	}
	testPool = pool

	jwtSecret = []byte("smoke-test-jwt-secret-key-for-local-tests-only!")
	piAPIKey = "smoke-test-pi-key"
	authMode = "auth_test"
	corsOrigin = "*"

	smokeFiles, err := os.MkdirTemp("", "gr33n-smoke-files")
	if err != nil {
		pool.Close()
		fmt.Fprintf(os.Stderr, "smoke_test mkdir temp files: %v\n", err)
		os.Exit(1)
	}
	mux := http.NewServeMux()
	// Shorten cooldown so smoke tests can fire the worker repeatedly without
	// being rate-limited by the default 2m window between successful runs.
	testNotifier = &recordingNotifier{}
	worker := automationworker.NewWorker(pool, true,
		automationworker.WithCooldown(0),
		automationworker.WithPushNotifier(testNotifier),
	)
	testWorker = worker
	store, err := filestorage.NewLocal(filepath.Join(smokeFiles, "blobs"))
	if err != nil {
		pool.Close()
		fmt.Fprintf(os.Stderr, "smoke_test init storage: %v\n", err)
		os.Exit(1)
	}
	registerRoutes(mux, pool, worker, nil, "admin", nil, "", store, filestorage.Config{Backend: "local"})
	testServer = httptest.NewServer(corsMiddleware(mux))

	code := m.Run()
	testServer.Close()
	pool.Close()
	_ = os.RemoveAll(smokeFiles)
	os.Exit(code)
}
