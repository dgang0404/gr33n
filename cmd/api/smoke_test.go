package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, rand.Int())
}

var (
	testServer   *httptest.Server
	testPool     *pgxpool.Pool
	testWorker   *automationworker.Worker
	testNotifier *recordingNotifier

	smokeTokenOnce sync.Once
	smokeToken     string
	smokeTokenErr  error
)

// recordingNotifier is a PushNotifier double used by the rule-driven
// send_notification smoke test to assert the worker fans out an alert
// through the push pipeline (without actually hitting FCM).
type recordingNotifier struct {
	mu     sync.Mutex
	alerts []db.Gr33ncoreAlertsNotification
}

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
	return nil
}

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://davidg@/gr33n?host=/var/run/postgresql"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		os.Exit(0)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		os.Exit(0)
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

func smokeJWT(t *testing.T) string {
	t.Helper()
	smokeTokenOnce.Do(func() {
		resp := postNoAuth("/auth/login", map[string]any{
			"username": smokeDevEmail,
			"password": smokeDevPass,
		})
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			smokeTokenErr = fmt.Errorf("login: status %d: %s", resp.StatusCode, string(b))
			return
		}
		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			smokeTokenErr = err
			return
		}
		tok, _ := body["token"].(string)
		if tok == "" {
			smokeTokenErr = fmt.Errorf("no token in login response")
			return
		}
		smokeToken = tok
	})
	if smokeTokenErr != nil {
		t.Fatalf("smoke JWT: %v", smokeTokenErr)
	}
	return smokeToken
}

// ── Health + Auth Mode ──────────────────────────────────────────────────────

func TestHealthEndpoint(t *testing.T) {
	resp := get(t, "/health")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body["status"])
	}
}

func TestAuthModeEndpoint(t *testing.T) {
	resp := get(t, "/auth/mode")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["mode"] != "auth_test" {
		t.Fatalf("expected mode=auth_test, got %v", body["mode"])
	}
}

func TestNotificationPreferencesAndPushTokens(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPatch(t, tok, "/profile/notification-preferences", map[string]any{
		"push_enabled": false,
		"min_priority": "medium",
	})
	expectStatus(t, resp, 200)
	resp = authGet(t, tok, "/profile/notification-preferences")
	expectStatus(t, resp, 200)
	m := decodeMap(t, resp)
	if pe, ok := m["push_enabled"].(bool); !ok || pe {
		t.Fatalf("expected push_enabled false after reset, got %+v", m)
	}
	resp = authPatch(t, tok, "/profile/notification-preferences", map[string]any{
		"push_enabled": true,
		"min_priority": "high",
	})
	expectStatus(t, resp, 200)
	m = decodeMap(t, resp)
	if m["push_enabled"] != true || m["min_priority"] != "high" {
		t.Fatalf("patch result %+v", m)
	}
	fakeTok := "smoke-fcm-" + uniqueName("tok")
	resp = authPost(t, tok, "/profile/push-tokens", map[string]any{
		"platform":  "android",
		"fcm_token": fakeTok,
	})
	expectStatus(t, resp, 200)
	resp = authGet(t, tok, "/profile/push-tokens")
	expectStatus(t, resp, 200)
	slice := decodeSlice(t, resp)
	if len(slice) != 1 {
		t.Fatalf("expected 1 push token, got %d", len(slice))
	}
	resp = authDeleteJSON(t, tok, "/profile/push-tokens", map[string]any{"fcm_token": fakeTok})
	expectStatus(t, resp, 204)
}

func TestJWTRequiredForDashboard(t *testing.T) {
	resp := get(t, "/farms/1")
	expectStatus(t, resp, http.StatusUnauthorized)
}

func TestLoginBadCredentials(t *testing.T) {
	resp := postNoAuth("/auth/login", map[string]any{
		"username": smokeDevEmail,
		"password": "not-the-password",
	})
	expectStatus(t, resp, http.StatusUnauthorized)
}

// ── Farm + Zone + Sensor Reads ──────────────────────────────────────────────

func TestGetFarm(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["name"] == nil {
		t.Fatal("expected farm to have a name")
	}
}

func TestCommonsCatalogBrowseAndImport(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/commons/catalog")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) < 1 {
		t.Fatal("expected at least one published catalog entry")
	}
	resp = authGet(t, tok, "/commons/catalog/gr33n-insert-commons-v1-readme")
	expectStatus(t, resp, 200)
	detail := decodeMap(t, resp)
	if detail["slug"] != "gr33n-insert-commons-v1-readme" {
		t.Fatalf("unexpected slug %v", detail["slug"])
	}
	resp = authGet(t, tok, "/farms/1/commons/catalog-imports")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)

	resp = authPost(t, tok, "/farms/1/commons/catalog-imports", map[string]any{
		"slug": "gr33n-insert-commons-v1-readme",
		"note": "smoke test",
	})
	expectStatus(t, resp, 200)
	out := decodeMap(t, resp)
	if out["import"] == nil || out["catalog_entry"] == nil {
		t.Fatalf("expected import and catalog_entry, got %#v", out)
	}
	resp = authGet(t, tok, "/farms/1/commons/catalog-imports")
	expectStatus(t, resp, 200)
	imports := decodeSlice(t, resp)
	if len(imports) < 1 {
		t.Fatal("expected farm to list at least one catalog import")
	}
}

func TestInsertCommonsPreview(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/insert-commons/preview")
	expectStatus(t, resp, http.StatusOK)
	m := decodeMap(t, resp)
	if v, ok := m["valid"].(bool); !ok || !v {
		t.Fatalf("expected valid preview, got %#v", m)
	}
	if m["payload"] == nil {
		t.Fatal("expected payload in preview response")
	}
}

func TestListZones(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/zones")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Fatal("expected at least one zone from seed data")
	}
}

func TestListSensors(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/sensors")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestSensorReadingsAndStats(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/sensors")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Skip("no sensors in seed")
	}
	m := items[0].(map[string]any)
	sid := int64(m["id"].(float64))
	resp = authGet(t, tok, fmt.Sprintf("/sensors/%d/readings?limit=10", sid))
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)

	resp = authGet(t, tok, fmt.Sprintf("/sensors/%d/readings/stats", sid))
	expectStatus(t, resp, 200)
	_ = decodeMap(t, resp)
}

// TestSensorAlertDurationAndCooldown verifies the Phase 19 WS2 state machine:
//  1. A reading that breaches the threshold but hasn't sustained for alert_duration_seconds
//     does NOT create an alert.
//  2. Once the streak has been backdated past the duration, the next breaching reading DOES
//     create an alert.
//  3. Further breaching readings within alert_cooldown_seconds are suppressed (no duplicate).
//  4. A reading that returns to bounds clears alert_breach_started_at.
func TestSensorAlertDurationAndCooldown(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Pick any unit id (exact unit doesn't matter for the evaluator).
	var unitID int64
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM gr33ncore.units WHERE name = 'celsius' LIMIT 1`,
	).Scan(&unitID); err != nil {
		if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
			t.Fatalf("find a unit id: %v", err)
		}
	}

	sensorName := uniqueName("alert_gate_sensor")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":                   sensorName,
		"sensor_type":            "temperature",
		"unit_id":                unitID,
		"alert_threshold_low":    10.0,
		"alert_threshold_high":   40.0,
		"alert_duration_seconds": 60,
		"alert_cooldown_seconds": 3600,
	})
	expectStatus(t, resp, http.StatusCreated)
	s := decodeMap(t, resp)
	sid := int64(s["id"].(float64))

	countAlerts := func() int {
		t.Helper()
		var n int
		err := testPool.QueryRow(ctx, `
			SELECT COUNT(*) FROM gr33ncore.alerts_notifications
			WHERE farm_id = 1
			  AND triggering_event_source_type = 'sensor_reading'
			  AND triggering_event_source_id = $1`, sid).Scan(&n)
		if err != nil {
			t.Fatalf("count alerts: %v", err)
		}
		return n
	}

	postReading := func(value float64) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{
			"value_raw": value,
			"is_valid":  true,
		})
		req, err := http.NewRequest(http.MethodPost,
			testServer.URL+fmt.Sprintf("/sensors/%d/readings", sid), bytes.NewReader(b))
		if err != nil {
			t.Fatalf("build reading request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", piAPIKey)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("post reading: %v", err)
		}
		expectStatus(t, resp, http.StatusCreated)
		resp.Body.Close()
	}

	waitForAlertCount := func(want int) int {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		var got int
		for time.Now().Before(deadline) {
			got = countAlerts()
			if got == want {
				return got
			}
			time.Sleep(40 * time.Millisecond)
		}
		return got
	}

	// Step 1 — breaching reading, but the streak is brand new: duration gate should suppress.
	postReading(5.0)
	if got := waitForAlertCount(0); got != 0 {
		t.Fatalf("expected 0 alerts while within duration window, got %d", got)
	}

	// The evaluator should have stamped alert_breach_started_at on the sensor.
	var breachStart *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT alert_breach_started_at FROM gr33ncore.sensors WHERE id = $1`, sid,
	).Scan(&breachStart); err != nil {
		t.Fatalf("read breach start: %v", err)
	}
	if breachStart == nil {
		// Evaluator runs in a goroutine; give it a brief moment.
		time.Sleep(200 * time.Millisecond)
		if err := testPool.QueryRow(ctx,
			`SELECT alert_breach_started_at FROM gr33ncore.sensors WHERE id = $1`, sid,
		).Scan(&breachStart); err != nil {
			t.Fatalf("read breach start (retry): %v", err)
		}
		if breachStart == nil {
			t.Fatal("expected alert_breach_started_at to be set after breaching reading")
		}
	}

	// Step 2 — backdate the streak past alert_duration_seconds, then re-post a breach.
	if _, err := testPool.Exec(ctx,
		`UPDATE gr33ncore.sensors SET alert_breach_started_at = NOW() - INTERVAL '10 minutes' WHERE id = $1`,
		sid,
	); err != nil {
		t.Fatalf("backdate breach start: %v", err)
	}
	postReading(4.5)
	if got := waitForAlertCount(1); got != 1 {
		t.Fatalf("expected exactly 1 alert once duration elapsed, got %d", got)
	}

	// Step 3 — another breaching reading within cooldown must NOT produce a second alert.
	postReading(4.0)
	// Give the goroutine a moment and re-check.
	time.Sleep(200 * time.Millisecond)
	if got := countAlerts(); got != 1 {
		t.Fatalf("expected cooldown to suppress duplicate, still 1 alert, got %d", got)
	}

	// Step 4 — a healthy reading should clear alert_breach_started_at.
	postReading(25.0)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var current *time.Time
		if err := testPool.QueryRow(ctx,
			`SELECT alert_breach_started_at FROM gr33ncore.sensors WHERE id = $1`, sid,
		).Scan(&current); err == nil && current == nil {
			return
		}
		time.Sleep(40 * time.Millisecond)
	}
	t.Fatal("expected alert_breach_started_at to be cleared after in-bounds reading")
}

// TestAlertToTaskLinkage verifies the Phase 19 WS3 flow:
//  1. A breaching reading creates an alert via the evaluator.
//  2. POST /alerts/{id}/create-task synthesises a task, inherits the sensor's zone,
//     derives a sensible priority from severity, and back-links source_alert_id.
//  3. Overrides in the body (title, priority, due_date) win over the derived defaults.
func TestAlertToTaskLinkage(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Unit (exact type doesn't matter for the evaluator).
	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	// Pick any zone on farm 1 so we can verify zone-carry-over.
	var zoneID int64
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM gr33ncore.zones WHERE farm_id = 1 AND deleted_at IS NULL ORDER BY id LIMIT 1`,
	).Scan(&zoneID); err != nil {
		t.Fatalf("find a zone id on farm 1: %v", err)
	}

	// Sensor with duration=0 so the first breaching reading fires immediately.
	sensorName := uniqueName("alert_to_task_sensor")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":                   sensorName,
		"sensor_type":            "temperature",
		"unit_id":                unitID,
		"zone_id":                zoneID,
		"alert_threshold_low":    10.0,
		"alert_threshold_high":   40.0,
		"alert_duration_seconds": 0,
		"alert_cooldown_seconds": 3600,
	})
	expectStatus(t, resp, http.StatusCreated)
	s := decodeMap(t, resp)
	sid := int64(s["id"].(float64))

	// Post a breaching reading via the Pi API-key path.
	b, _ := json.Marshal(map[string]any{"value_raw": 5.0, "is_valid": true})
	req, err := http.NewRequest(http.MethodPost,
		testServer.URL+fmt.Sprintf("/sensors/%d/readings", sid), bytes.NewReader(b))
	if err != nil {
		t.Fatalf("build reading request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", piAPIKey)
	rresp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post reading: %v", err)
	}
	expectStatus(t, rresp, http.StatusCreated)
	rresp.Body.Close()

	// Wait for the evaluator goroutine to persist the alert.
	var alertID int64
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if err := testPool.QueryRow(ctx, `
			SELECT id FROM gr33ncore.alerts_notifications
			WHERE farm_id = 1
			  AND triggering_event_source_type = 'sensor_reading'
			  AND triggering_event_source_id = $1
			ORDER BY id DESC LIMIT 1`, sid).Scan(&alertID); err == nil {
			break
		}
		time.Sleep(40 * time.Millisecond)
	}
	if alertID == 0 {
		t.Fatal("expected an alert to be created for the breaching reading")
	}

	// --- Case A: empty body → server-derived title/priority/zone. ---
	resp = authPost(t, tok, fmt.Sprintf("/alerts/%d/create-task", alertID), map[string]any{})
	expectStatus(t, resp, http.StatusCreated)
	task := decodeMap(t, resp)

	if int64(task["farm_id"].(float64)) != 1 {
		t.Fatalf("expected task.farm_id = 1, got %v", task["farm_id"])
	}
	if int64(task["source_alert_id"].(float64)) != alertID {
		t.Fatalf("expected task.source_alert_id = %d, got %v", alertID, task["source_alert_id"])
	}
	if task["zone_id"] == nil {
		t.Fatalf("expected task.zone_id to be derived from the sensor, got nil")
	}
	if int64(task["zone_id"].(float64)) != zoneID {
		t.Fatalf("expected task.zone_id = %d (sensor zone), got %v", zoneID, task["zone_id"])
	}
	if title, _ := task["title"].(string); strings.TrimSpace(title) == "" {
		t.Fatal("expected a non-empty title derived from the alert")
	}
	// Default task_type from alert-create path.
	if tt, _ := task["task_type"].(string); tt != "alert_follow_up" {
		t.Fatalf("expected task_type=alert_follow_up, got %q", tt)
	}

	// --- Case B: overrides win. ---
	resp = authPost(t, tok, fmt.Sprintf("/alerts/%d/create-task", alertID), map[string]any{
		"title":    "custom follow-up",
		"priority": 3,
		"due_date": "2030-01-15",
	})
	expectStatus(t, resp, http.StatusCreated)
	task2 := decodeMap(t, resp)
	if got, _ := task2["title"].(string); got != "custom follow-up" {
		t.Fatalf("expected override title, got %q", got)
	}
	if int64(task2["priority"].(float64)) != 3 {
		t.Fatalf("expected override priority=3, got %v", task2["priority"])
	}
	if int64(task2["source_alert_id"].(float64)) != alertID {
		t.Fatalf("expected task2.source_alert_id = %d, got %v", alertID, task2["source_alert_id"])
	}

	// Both tasks should land in ListTasksByFarm with source_alert_id set.
	resp = authGet(t, tok, "/farms/1/tasks")
	expectStatus(t, resp, http.StatusOK)
	list := decodeSlice(t, resp)
	linked := 0
	for _, row := range list {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		if sa, ok := m["source_alert_id"].(float64); ok && int64(sa) == alertID {
			linked++
		}
	}
	if linked < 2 {
		t.Fatalf("expected at least 2 tasks with source_alert_id=%d in farm list, got %d", alertID, linked)
	}

	// --- Case C: bogus alert id returns 404. ---
	resp = authPost(t, tok, "/alerts/99999999/create-task", map[string]any{})
	expectStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()
}

// TestSchedulePreconditionFailsRun verifies Phase 19 WS4 interlock-lite:
//  1. Creating a schedule with an invalid precondition (bogus sensor id) is rejected.
//  2. When the latest reading for a sensor fails the predicate, the worker's
//     Tick() records an automation_runs row with status='skipped' and
//     message='precondition_failed' and does NOT fire executable actions.
//  3. When the reading satisfies the predicate, the worker proceeds as usual
//     (no interlock skip).
func TestSchedulePreconditionFailsRun(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	// Sensor on farm 1 — we'll seed a "tank empty" reading below its threshold.
	sensorName := uniqueName("precondition_sensor")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":        sensorName,
		"sensor_type": "level",
		"unit_id":     unitID,
	})
	expectStatus(t, resp, http.StatusCreated)
	sensorRow := decodeMap(t, resp)
	sid := int64(sensorRow["id"].(float64))

	// Seed a failing reading: level = 2, rule will require >= 50.
	if _, err := testPool.Exec(ctx, `
		INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid)
		VALUES (NOW(), $1, 2, TRUE)`, sid); err != nil {
		t.Fatalf("seed failing reading: %v", err)
	}

	// --- Validation: precondition with a sensor from another farm is rejected. ---
	resp = authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            uniqueName("precond_invalid"),
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
		"preconditions": []map[string]any{
			{"sensor_id": 999999, "op": "gte", "value": 50.0},
		},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// --- Validation: unknown op is rejected. ---
	resp = authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            uniqueName("precond_badop"),
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
		"preconditions": []map[string]any{
			{"sensor_id": sid, "op": "totally-invalid", "value": 50.0},
		},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// Every-minute schedule with a precondition that the current reading (2) will FAIL.
	schedName := uniqueName("interlock_schedule")
	resp = authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            schedName,
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
		"preconditions": []map[string]any{
			{"sensor_id": sid, "op": "gte", "value": 50.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	schedRow := decodeMap(t, resp)
	schedID := int64(schedRow["id"].(float64))

	// Remember how many actuator events exist for this farm before the tick — the
	// worker must NOT write any, since no executable actions should run.
	var evBefore int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.actuator_events
		WHERE triggered_by_schedule_id = $1`, schedID).Scan(&evBefore); err != nil {
		t.Fatalf("count actuator events before: %v", err)
	}

	// Run a tick — the precondition should fail and the run should be recorded as skipped.
	testWorker.Tick(ctx)

	// --- Assert a skipped run with message='precondition_failed' exists. ---
	var msg string
	var status string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, message, details::text FROM gr33ncore.automation_runs
		WHERE schedule_id = $1 AND status = 'skipped' AND message = 'precondition_failed'
		ORDER BY executed_at DESC LIMIT 1`, schedID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("expected a skipped run with precondition_failed: %v", err)
	}
	if status != "skipped" || msg != "precondition_failed" {
		t.Fatalf("expected status=skipped message=precondition_failed, got %s/%s", status, msg)
	}
	// PostgreSQL's JSONB rendering collapses whitespace inconsistently
	// between releases, so parse before asserting.
	var details struct {
		Phase  string `json:"phase"`
		Failed []struct {
			SensorID int64   `json:"sensor_id"`
			Op       string  `json:"op"`
			Expected float64 `json:"expected"`
			Actual   float64 `json:"actual"`
			Reason   string  `json:"reason"`
		} `json:"failed"`
	}
	if err := json.Unmarshal(detailsJSON, &details); err != nil {
		t.Fatalf("parse details json: %v (raw=%s)", err, string(detailsJSON))
	}
	if details.Phase != "preconditions" {
		t.Fatalf("expected details.phase=preconditions, got %q", details.Phase)
	}
	if len(details.Failed) != 1 {
		t.Fatalf("expected 1 failed precondition, got %d", len(details.Failed))
	}
	f := details.Failed[0]
	if f.SensorID != sid || f.Op != "gte" || f.Expected != 50 || f.Actual != 2 || f.Reason != "predicate_failed" {
		t.Fatalf("unexpected failed entry: %+v", f)
	}

	// No actuator events should have been written for this schedule.
	var evAfter int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.actuator_events
		WHERE triggered_by_schedule_id = $1`, schedID).Scan(&evAfter); err != nil {
		t.Fatalf("count actuator events after: %v", err)
	}
	if evAfter != evBefore {
		t.Fatalf("expected no actuator events when precondition fails, got %d new", evAfter-evBefore)
	}
	// Last-triggered should remain NULL — the next tick should get another chance.
	var lastTriggered *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.schedules WHERE id = $1`, schedID,
	).Scan(&lastTriggered); err != nil {
		t.Fatalf("read last_triggered_time: %v", err)
	}
	if lastTriggered != nil {
		t.Fatalf("expected last_triggered_time to remain NULL when skipped by precondition, got %v", *lastTriggered)
	}

	// --- Flip the predicate: the reading (2) satisfies op=lte value=50. ---
	resp = authPut(t, tok, fmt.Sprintf("/schedules/%d", schedID), map[string]any{
		"name":            schedName,
		"schedule_type":   "cron",
		"cron_expression": "* * * * *",
		"timezone":        "UTC",
		"is_active":       true,
		"preconditions": []map[string]any{
			{"sensor_id": sid, "op": "lte", "value": 50.0},
		},
	})
	expectStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	testWorker.Tick(ctx)

	// After flipping, preconditions pass and the worker should proceed to
	// executeSchedule. No executable actions are attached, so the run we
	// care about is the post-precondition run — we assert it is NOT a
	// precondition_failed row. Several rows may share executed_at within
	// the same minute, so order by id (monotonic) rather than timestamp.
	var latestStatus, latestMsg string
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, '') FROM gr33ncore.automation_runs
		WHERE schedule_id = $1
		ORDER BY id DESC LIMIT 1`, schedID).Scan(&latestStatus, &latestMsg); err != nil {
		t.Fatalf("read latest run after flip: %v", err)
	}
	if latestMsg == "precondition_failed" {
		t.Fatalf("expected the latest run to pass preconditions after flipping the rule, got message=%s", latestMsg)
	}

	// Double-check that the flip caused a new run to be recorded — counts
	// should reflect at least one non-precondition_failed skipped/success row.
	var nonPrecondCount int
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM gr33ncore.automation_runs
		WHERE schedule_id = $1 AND COALESCE(message, '') <> 'precondition_failed'`, schedID,
	).Scan(&nonPrecondCount); err != nil {
		t.Fatalf("count non-precondition runs: %v", err)
	}
	if nonPrecondCount == 0 {
		t.Fatalf("expected at least one non-precondition_failed run after flip")
	}
}

func TestListActuators(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/actuators")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestListSchedules(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/schedules")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestListAutomationRuns(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/automation/runs")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestListTasks(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/tasks")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestWorkerHealth(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/automation/worker/health")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["simulation_mode"] != true {
		t.Fatal("expected simulation_mode=true")
	}
}

// ── Phase 9 CRUD + authz ─────────────────────────────────────────────────────

func TestTaskCreate(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title": "smoke task",
	})
	expectStatus(t, resp, 201)
}

func TestCropCycleCreateAndStage(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_cycle")
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       1,
		"name":          name,
		"current_stage": "early_veg",
		"started_at":    "2025-01-01",
		"is_active":     false,
	})
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	id := int64(created["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/crop-cycles/%d/stage", id), map[string]any{
		"current_stage": "late_veg",
	})
	expectStatus(t, resp, 200)
}

func TestCostsSummaryListExport(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/costs/summary")
	expectStatus(t, resp, 200)

	resp = authGet(t, tok, "/farms/1/costs?limit=5")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)

	resp = authGet(t, tok, "/farms/1/costs/export?format=csv")
	expectStatus(t, resp, 200)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(body), "date,category,amount,currency,is_income,description,document_type") {
		t.Fatalf("expected CSV header, got %q", string(body[:min(80, len(body))]))
	}

	resp = authGet(t, tok, "/farms/1/costs/export?format=gl_csv")
	expectStatus(t, resp, 200)
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(body), "date,entry_type,account_code") {
		t.Fatalf("expected GL CSV header, got %q", string(body[:min(60, len(body))]))
	}

	resp = authGet(t, tok, "/farms/1/costs/export?format=summary_csv")
	expectStatus(t, resp, 200)
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(body), "period,category,currency,income_total") {
		t.Fatalf("expected summary CSV header, got %q", string(body[:min(70, len(body))]))
	}
}

func TestOrganizationCreateListUsageAndFarmLink(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/organizations", map[string]any{"name": "Smoke Tenant Org"})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	orgID := int64(created["id"].(float64))

	resp = authGet(t, tok, "/organizations")
	expectStatus(t, resp, http.StatusOK)
	list := decodeSlice(t, resp)
	if len(list) == 0 {
		t.Fatal("expected at least one organization")
	}

	resp = authGet(t, tok, fmt.Sprintf("/organizations/%d/usage-summary", orgID))
	expectStatus(t, resp, http.StatusOK)
	summary := decodeMap(t, resp)
	if _, ok := summary["farm_count"]; !ok {
		t.Fatalf("usage summary missing farm_count: %#v", summary)
	}

	resp = authGet(t, tok, fmt.Sprintf("/organizations/%d/audit-events?limit=10", orgID))
	expectStatus(t, resp, http.StatusOK)
	_ = decodeSlice(t, resp)

	resp = authPatch(t, tok, "/farms/1/organization", map[string]any{"organization_id": orgID})
	expectStatus(t, resp, http.StatusOK)
	farm := decodeMap(t, resp)
	if int64(farm["organization_id"].(float64)) != orgID {
		t.Fatalf("expected farm linked to org %d, got %#v", orgID, farm["organization_id"])
	}

	resp = authPatch(t, tok, "/farms/1/organization", map[string]any{"organization_id": nil})
	expectStatus(t, resp, http.StatusOK)
}

func TestOrgDefaultBootstrapOnFarmCreate(t *testing.T) {
	tok := smokeJWT(t)
	resp := authPost(t, tok, "/organizations", map[string]any{"name": uniqueName("org_bootstrap_default")})
	expectStatus(t, resp, http.StatusCreated)
	org := decodeMap(t, resp)
	orgID := int64(org["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/organizations/%d", orgID), map[string]any{
		"default_bootstrap_template": "jadam_indoor_photoperiod_v1",
	})
	expectStatus(t, resp, http.StatusOK)

	name := uniqueName("org_default_farm")
	resp = authPost(t, tok, "/farms", map[string]any{
		"name":               name,
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"organization_id":    orgID,
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	farmObj, ok := payload["farm"].(map[string]any)
	if !ok {
		t.Fatalf("expected farm + bootstrap from org default, got %#v", payload)
	}
	if _, ok := payload["bootstrap"]; !ok {
		t.Fatal("expected bootstrap in response when org default applies")
	}
	fid := int64(farmObj["id"].(float64))
	zones := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/zones", fid)))
	if len(zones) < 4 {
		t.Fatalf("expected org default bootstrap zones, got %d", len(zones))
	}

	name2 := uniqueName("org_default_farm_explicit_none")
	resp = authPost(t, tok, "/farms", map[string]any{
		"name":               name2,
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"organization_id":    orgID,
		"bootstrap_template": "none",
	})
	expectStatus(t, resp, http.StatusCreated)
	payload2 := decodeMap(t, resp)
	farm2 := payload2["farm"].(map[string]any)
	fid2 := int64(farm2["id"].(float64))
	zones2 := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/zones", fid2)))
	if len(zones2) != 0 {
		t.Fatalf("bootstrap_template none should skip org default, got %d zones", len(zones2))
	}
}

func TestCoaMappingsListAndUpdate(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/finance/coa-mappings")
	expectStatus(t, resp, http.StatusOK)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Fatal("expected default coa mappings")
	}

	resp = authPut(t, tok, "/farms/1/finance/coa-mappings", map[string]any{
		"mappings": []map[string]any{
			{
				"category":     "miscellaneous",
				"account_code": "6999",
				"account_name": "Custom misc expense",
			},
		},
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeSlice(t, resp)
	found := false
	for _, it := range updated {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if row["category"] == "miscellaneous" && row["account_code"] == "6999" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected updated miscellaneous coa mapping")
	}

	resp = authDelete(t, tok, "/farms/1/finance/coa-mappings/miscellaneous")
	expectStatus(t, resp, http.StatusOK)
	resetOne := decodeSlice(t, resp)
	for _, it := range resetOne {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if row["category"] == "miscellaneous" && row["source"] != "default" {
			t.Fatal("expected miscellaneous mapping reset to default")
		}
	}

	resp = authDelete(t, tok, "/farms/1/finance/coa-mappings")
	expectStatus(t, resp, http.StatusOK)
	resetAll := decodeSlice(t, resp)
	for _, it := range resetAll {
		row, ok := it.(map[string]any)
		if !ok {
			continue
		}
		if row["source"] != "default" {
			t.Fatal("expected all mappings reset to default")
		}
	}
}

func TestCostReceiptUploadAndDownload(t *testing.T) {
	tok := smokeJWT(t)
	costID := createSmokeCost(t, tok)
	attachmentID := uploadSmokeReceipt(t, tok, costID, "receipt.pdf", []byte("%PDF-1.4 smoke\n"))

	resp := authGet(t, tok, fmt.Sprintf("/file-attachments/%d/download", attachmentID))
	expectStatus(t, resp, http.StatusOK)
	target := decodeMap(t, resp)
	if target["proxied"] != true {
		t.Fatalf("proxied = %v, want true for local storage", target["proxied"])
	}
	if target["backend"] != "local" {
		t.Fatalf("backend = %v, want local", target["backend"])
	}
	if target["url"] != fmt.Sprintf("/file-attachments/%d/content", attachmentID) {
		t.Fatalf("url = %v", target["url"])
	}

	resp = authGet(t, tok, fmt.Sprintf("/file-attachments/%d/content", attachmentID))
	expectStatus(t, resp, http.StatusOK)
	defer resp.Body.Close()
	if got := resp.Header.Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("Content-Type = %q, want application/pdf", got)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "%PDF-1.4 smoke\n" {
		t.Fatalf("downloaded bytes = %q", string(data))
	}
}

func TestCostReceiptReplacementCleansUpOldAttachment(t *testing.T) {
	tok := smokeJWT(t)
	costID := createSmokeCost(t, tok)
	firstID := uploadSmokeReceipt(t, tok, costID, "receipt-a.pdf", []byte("%PDF-1.4 first\n"))
	secondID := uploadSmokeReceipt(t, tok, costID, "receipt-b.pdf", []byte("%PDF-1.4 second\n"))

	resp := authGet(t, tok, fmt.Sprintf("/file-attachments/%d/content", firstID))
	expectStatus(t, resp, http.StatusNotFound)

	resp = authGet(t, tok, fmt.Sprintf("/file-attachments/%d/content", secondID))
	expectStatus(t, resp, http.StatusOK)
}

func TestDeletingCostCleansUpReceiptAttachment(t *testing.T) {
	tok := smokeJWT(t)
	costID := createSmokeCost(t, tok)
	attachmentID := uploadSmokeReceipt(t, tok, costID, "receipt-delete.pdf", []byte("%PDF-1.4 delete\n"))

	resp := authDelete(t, tok, fmt.Sprintf("/costs/%d", costID))
	expectStatus(t, resp, http.StatusNoContent)

	resp = authGet(t, tok, fmt.Sprintf("/file-attachments/%d/content", attachmentID))
	expectStatus(t, resp, http.StatusNotFound)
}

func TestRecipeList(t *testing.T) {
	tok := smokeJWT(t)
	resp := authGet(t, tok, "/farms/1/naturalfarming/recipes")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Fatal("expected seeded recipes")
	}
}

func TestCrossFarmWriteForbidden(t *testing.T) {
	email := fmt.Sprintf("norole_%d@smoke.test", rand.Int())
	resp := postNoAuth("/auth/register", map[string]any{
		"email":     email,
		"password":  "longpassword1",
		"full_name": "No Farm",
	})
	expectStatus(t, resp, http.StatusCreated)
	reg := decodeMap(t, resp)
	otherTok, _ := reg["token"].(string)
	if otherTok == "" {
		t.Fatal("expected token from register")
	}

	resp = authPost(t, otherTok, "/farms/1/tasks", map[string]any{"title": "should fail"})
	expectStatus(t, resp, http.StatusForbidden)
}

// ── Fertigation CRUD ────────────────────────────────────────────────────────

func TestFertigationReservoirRoundtrip(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_reservoir")
	payload := map[string]any{
		"name":                  name,
		"status":                "ready",
		"capacity_liters":       100.0,
		"current_volume_liters": 50.0,
	}
	resp := authPost(t, tok, "/farms/1/fertigation/reservoirs", payload)
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	if created["name"] != name {
		t.Fatalf("expected name=%s, got %v", name, created["name"])
	}

	resp = authGet(t, tok, "/farms/1/fertigation/reservoirs")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	found := false
	for _, item := range items {
		if m, ok := item.(map[string]any); ok && m["name"] == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created reservoir not found in list")
	}
}

func TestFertigationEcTargetRoundtrip(t *testing.T) {
	tok := smokeJWT(t)
	payload := map[string]any{
		"growth_stage": "early_veg",
		"ec_min_mscm":  1.0,
		"ec_max_mscm":  2.5,
		"ph_min":       5.5,
		"ph_max":       6.5,
		"notes":        "smoke test",
	}
	resp := authPost(t, tok, "/farms/1/fertigation/ec-targets", payload)
	expectStatus(t, resp, 201)

	resp = authGet(t, tok, "/farms/1/fertigation/ec-targets")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestFertigationProgramRoundtrip(t *testing.T) {
	tok := smokeJWT(t)
	payload := map[string]any{
		"name":                uniqueName("smoke_program"),
		"total_volume_liters": 5.0,
		"is_active":           false,
		"ec_trigger_low":      0.0,
		"ph_trigger_low":      0.0,
		"ph_trigger_high":     0.0,
	}
	resp := authPost(t, tok, "/farms/1/fertigation/programs", payload)
	expectStatus(t, resp, 201)

	resp = authGet(t, tok, "/farms/1/fertigation/programs")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestFertigationEventRoundtripWithCropCycle(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_cc_fert")
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       1,
		"name":          name,
		"current_stage": "early_veg",
		"started_at":    "2025-02-01",
		"is_active":     false,
	})
	expectStatus(t, resp, 201)
	cc := decodeMap(t, resp)
	ccID := int64(cc["id"].(float64))

	payload := map[string]any{
		"zone_id":               1,
		"crop_cycle_id":         ccID,
		"volume_applied_liters": 2.5,
		"ec_before_mscm":        1.2,
		"ec_after_mscm":         1.8,
		"ph_before":             6.0,
		"ph_after":              6.2,
		"trigger_source":        "manual",
	}
	resp = authPost(t, tok, "/farms/1/fertigation/events", payload)
	expectStatus(t, resp, 201)

	resp = authGet(t, tok, fmt.Sprintf("/farms/1/fertigation/events?crop_cycle_id=%d", ccID))
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Fatal("expected filtered fertigation events")
	}
}

// ── Phase 16: Schedule CRUD ─────────────────────────────────────────────────

func TestScheduleCreateUpdateDelete(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("smoke_schedule")
	resp := authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            name,
		"schedule_type":   "cron",
		"cron_expression": "0 6 * * *",
		"timezone":        "UTC",
		"is_active":       true,
	})
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	if created["name"] != name {
		t.Fatalf("expected name=%s, got %v", name, created["name"])
	}
	id := int64(created["id"].(float64))

	updatedName := uniqueName("smoke_schedule_upd")
	resp = authPut(t, tok, fmt.Sprintf("/schedules/%d", id), map[string]any{
		"name":            updatedName,
		"schedule_type":   "cron",
		"cron_expression": "0 8 * * *",
		"timezone":        "America/New_York",
		"is_active":       false,
	})
	expectStatus(t, resp, 200)
	updated := decodeMap(t, resp)
	if updated["name"] != updatedName {
		t.Fatalf("expected updated name=%s, got %v", updatedName, updated["name"])
	}
	if updated["is_active"] != false {
		t.Fatal("expected is_active=false after update")
	}

	resp = authDelete(t, tok, fmt.Sprintf("/schedules/%d", id))
	expectStatus(t, resp, 204)

	resp = authGet(t, tok, "/farms/1/schedules")
	expectStatus(t, resp, 200)
	schedList := decodeSlice(t, resp)
	for _, s := range schedList {
		if m, ok := s.(map[string]any); ok && m["name"] == updatedName {
			t.Fatal("deleted schedule still appears in list")
		}
	}
}

// ── Phase 16: Mixing Event Creation ─────────────────────────────────────────

func TestMixingEventCreateWithComponents(t *testing.T) {
	tok := smokeJWT(t)

	resName := uniqueName("smoke_mix_res")
	resp := authPost(t, tok, "/farms/1/fertigation/reservoirs", map[string]any{
		"name":                  resName,
		"status":                "ready",
		"capacity_liters":       50.0,
		"current_volume_liters": 40.0,
	})
	expectStatus(t, resp, 201)
	res := decodeMap(t, resp)
	resID := int64(res["id"].(float64))

	inputsResp := authGet(t, tok, "/farms/1/naturalfarming/inputs")
	expectStatus(t, inputsResp, 200)
	inputs := decodeSlice(t, inputsResp)
	if len(inputs) == 0 {
		t.Skip("no NF inputs in seed data")
	}
	inputDef := inputs[0].(map[string]any)
	inputDefID := int64(inputDef["id"].(float64))

	resp = authPost(t, tok, "/farms/1/fertigation/mixing-events", map[string]any{
		"reservoir_id":        resID,
		"water_volume_liters": 20.0,
		"water_source":        "municipal",
		"water_ec_mscm":       0.3,
		"water_ph":            7.0,
		"final_ec_mscm":       1.5,
		"final_ph":            6.2,
		"notes":               "smoke test mix",
		"components": []map[string]any{
			{
				"input_definition_id": inputDefID,
				"volume_added_ml":     40.0,
				"dilution_ratio":      "1:500",
			},
		},
	})
	expectStatus(t, resp, 201)
	result := decodeMap(t, resp)
	if result["event"] == nil {
		t.Fatal("expected event in response")
	}
	comps, ok := result["components"].([]any)
	if !ok || len(comps) != 1 {
		t.Fatalf("expected 1 component, got %v", result["components"])
	}
}

// ── Phase 16: Task Update + Delete ──────────────────────────────────────────

func TestTaskUpdateAndDelete(t *testing.T) {
	tok := smokeJWT(t)

	resp := authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title":    "smoke update task",
		"priority": 1,
	})
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	taskID := int64(created["id"].(float64))

	updatedTitle := uniqueName("smoke_task_upd")
	resp = authPut(t, tok, fmt.Sprintf("/tasks/%d", taskID), map[string]any{
		"title":    updatedTitle,
		"priority": 2,
	})
	expectStatus(t, resp, 200)
	updated := decodeMap(t, resp)
	if updated["title"] != updatedTitle {
		t.Fatalf("expected title=%s, got %v", updatedTitle, updated["title"])
	}
	prio, _ := updated["priority"].(float64)
	if int(prio) != 2 {
		t.Fatalf("expected priority=2, got %v", updated["priority"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/tasks/%d", taskID))
	expectStatus(t, resp, 204)

	resp = authGet(t, tok, "/farms/1/tasks")
	expectStatus(t, resp, 200)
	taskList := decodeSlice(t, resp)
	for _, item := range taskList {
		if m, ok := item.(map[string]any); ok {
			if int64(m["id"].(float64)) == taskID {
				t.Fatal("soft-deleted task still appears in list")
			}
		}
	}
}

func TestPlantCRUD(t *testing.T) {
	tok := smokeJWT(t)

	// Create
	name := uniqueName("smoke_plant")
	resp := authPost(t, tok, "/farms/1/plants", map[string]any{
		"display_name":        name,
		"variety_or_cultivar": "Indica",
		"meta":                map[string]any{"photoperiod": "short-day"},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	plantID := int64(created["id"].(float64))
	if created["display_name"] != name {
		t.Fatalf("expected display_name=%s, got %v", name, created["display_name"])
	}

	// List
	resp = authGet(t, tok, "/farms/1/plants")
	expectStatus(t, resp, http.StatusOK)
	plants := decodeSlice(t, resp)
	found := false
	for _, item := range plants {
		if m, ok := item.(map[string]any); ok {
			if int64(m["id"].(float64)) == plantID {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("created plant not found in list")
	}

	// Get
	resp = authGet(t, tok, fmt.Sprintf("/plants/%d", plantID))
	expectStatus(t, resp, http.StatusOK)
	got := decodeMap(t, resp)
	if got["display_name"] != name {
		t.Fatalf("get: expected display_name=%s, got %v", name, got["display_name"])
	}

	// Update
	updatedName := uniqueName("smoke_plant_upd")
	resp = authPut(t, tok, fmt.Sprintf("/plants/%d", plantID), map[string]any{
		"display_name":        updatedName,
		"variety_or_cultivar": "Sativa",
		"meta":                map[string]any{"photoperiod": "long-day"},
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["display_name"] != updatedName {
		t.Fatalf("expected updated name=%s, got %v", updatedName, updated["display_name"])
	}

	// Soft delete
	resp = authDelete(t, tok, fmt.Sprintf("/plants/%d", plantID))
	expectStatus(t, resp, http.StatusNoContent)

	// Verify gone from list
	resp = authGet(t, tok, "/farms/1/plants")
	expectStatus(t, resp, http.StatusOK)
	plantsAfter := decodeSlice(t, resp)
	for _, item := range plantsAfter {
		if m, ok := item.(map[string]any); ok {
			if int64(m["id"].(float64)) == plantID {
				t.Fatal("soft-deleted plant still appears in list")
			}
		}
	}
}

// ── Phase 18: Smoke Test Gap Fill ────────────────────────────────────────────

func TestAlertLifecycle(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/farms/1/alerts")
	expectStatus(t, resp, http.StatusOK)
	alerts := decodeSlice(t, resp)

	resp = authGet(t, tok, "/farms/1/alerts/unread-count")
	expectStatus(t, resp, http.StatusOK)
	countMap := decodeMap(t, resp)
	if _, ok := countMap["unread_count"]; !ok {
		t.Fatalf("expected unread_count field in response, got %#v", countMap)
	}

	if len(alerts) == 0 {
		t.Skip("no alerts in seed data to test read/acknowledge")
	}

	first := alerts[0].(map[string]any)
	alertID := int64(first["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/alerts/%d/read", alertID), map[string]any{})
	expectStatus(t, resp, http.StatusOK)

	resp = authPatch(t, tok, fmt.Sprintf("/alerts/%d/acknowledge", alertID), map[string]any{})
	expectStatus(t, resp, http.StatusOK)
}

func TestCropCycleFullCRUD(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_cc_crud")
	resp := authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       1,
		"name":          name,
		"current_stage": "seedling",
		"started_at":    "2025-03-01",
		"is_active":     false,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ccID := int64(created["id"].(float64))

	resp = authGet(t, tok, fmt.Sprintf("/crop-cycles/%d", ccID))
	expectStatus(t, resp, http.StatusOK)
	got := decodeMap(t, resp)
	if got["name"] != name {
		t.Fatalf("GET crop cycle: expected name=%s, got %v", name, got["name"])
	}

	updName := uniqueName("smoke_cc_upd")
	resp = authPut(t, tok, fmt.Sprintf("/crop-cycles/%d", ccID), map[string]any{
		"name":      updName,
		"zone_id":   1,
		"is_active": false,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["name"] != updName {
		t.Fatalf("PUT crop cycle: expected name=%s, got %v", updName, updated["name"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/crop-cycles/%d", ccID))
	expectStatus(t, resp, http.StatusNoContent)

	resp = authGet(t, tok, "/farms/1/crop-cycles")
	expectStatus(t, resp, http.StatusOK)
	cycles := decodeSlice(t, resp)
	for _, c := range cycles {
		m := c.(map[string]any)
		if int64(m["id"].(float64)) == ccID && m["is_active"] == true {
			t.Fatal("deleted crop cycle still active in list")
		}
	}
}

func TestFertigationReservoirUpdateDelete(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_res_ud")
	resp := authPost(t, tok, "/farms/1/fertigation/reservoirs", map[string]any{
		"name":                  name,
		"status":                "ready",
		"capacity_liters":       80.0,
		"current_volume_liters": 40.0,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	resID := int64(created["id"].(float64))

	updName := uniqueName("smoke_res_upd")
	resp = authPatch(t, tok, fmt.Sprintf("/fertigation/reservoirs/%d", resID), map[string]any{
		"name":                  updName,
		"status":                "mixing",
		"capacity_liters":       80.0,
		"current_volume_liters": 35.0,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["name"] != updName {
		t.Fatalf("expected updated name=%s, got %v", updName, updated["name"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/fertigation/reservoirs/%d", resID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestFertigationProgramUpdateDelete(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_prog_ud")
	resp := authPost(t, tok, "/farms/1/fertigation/programs", map[string]any{
		"name":                name,
		"total_volume_liters": 10.0,
		"is_active":           false,
		"ec_trigger_low":      0.0,
		"ph_trigger_low":      0.0,
		"ph_trigger_high":     0.0,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	progID := int64(created["id"].(float64))

	updName := uniqueName("smoke_prog_upd")
	resp = authPatch(t, tok, fmt.Sprintf("/fertigation/programs/%d", progID), map[string]any{
		"name":      updName,
		"is_active": true,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["name"] != updName {
		t.Fatalf("expected updated name=%s, got %v", updName, updated["name"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/fertigation/programs/%d", progID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestNfInputDefinitionCRUD(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_nf_input")
	resp := authPost(t, tok, "/farms/1/naturalfarming/inputs", map[string]any{
		"name":        name,
		"category":    "fermented_plant_juice",
		"description": "smoke test input",
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	inputID := int64(created["id"].(float64))

	updName := uniqueName("smoke_nf_upd")
	resp = authPut(t, tok, fmt.Sprintf("/naturalfarming/inputs/%d", inputID), map[string]any{
		"name":        updName,
		"category":    "fermented_plant_juice",
		"description": "updated",
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["name"] != updName {
		t.Fatalf("expected updated name=%s, got %v", updName, updated["name"])
	}

	resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/inputs/%d", inputID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestNfBatchCRUD(t *testing.T) {
	tok := smokeJWT(t)

	inputsResp := authGet(t, tok, "/farms/1/naturalfarming/inputs")
	expectStatus(t, inputsResp, http.StatusOK)
	inputs := decodeSlice(t, inputsResp)
	if len(inputs) == 0 {
		t.Skip("no NF inputs to create batch against")
	}
	inputID := int64(inputs[0].(map[string]any)["id"].(float64))

	code := uniqueName("batch")
	resp := authPost(t, tok, "/farms/1/naturalfarming/batches", map[string]any{
		"input_definition_id": inputID,
		"batch_identifier":    code,
		"status":              "fermenting_brewing",
		"creation_start_date": "2025-06-01",
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	batchID := int64(created["id"].(float64))

	resp = authPut(t, tok, fmt.Sprintf("/naturalfarming/batches/%d", batchID), map[string]any{
		"input_definition_id": inputID,
		"batch_identifier":    code,
		"status":              "ready_for_use",
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/batches/%d", batchID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestRecipeFullCRUD(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_recipe")
	resp := authPost(t, tok, "/farms/1/naturalfarming/recipes", map[string]any{
		"name":                    name,
		"description":             "smoke recipe",
		"target_application_type": "soil_drench",
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	recipeID := int64(created["id"].(float64))

	resp = authGet(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", recipeID))
	expectStatus(t, resp, http.StatusOK)

	inputsResp := authGet(t, tok, "/farms/1/naturalfarming/inputs")
	expectStatus(t, inputsResp, http.StatusOK)
	inputs := decodeSlice(t, inputsResp)
	if len(inputs) > 0 {
		inputID := int64(inputs[0].(map[string]any)["id"].(float64))

		resp = authPost(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/components", recipeID), map[string]any{
			"input_definition_id": inputID,
			"volume_ml":          20.0,
			"dilution_ratio":     "1:500",
		})
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
			t.Fatalf("add component: expected 2xx, got %d", resp.StatusCode)
		}

		resp = authGet(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/components", recipeID))
		expectStatus(t, resp, http.StatusOK)
		comps := decodeSlice(t, resp)
		if len(comps) == 0 {
			t.Fatal("expected at least one recipe component")
		}

		resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d/components/%d", recipeID, inputID))
		expectStatus(t, resp, http.StatusNoContent)
	}

	updName := uniqueName("smoke_recipe_upd")
	resp = authPut(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", recipeID), map[string]any{
		"name":                    updName,
		"description":             "updated smoke recipe",
		"target_application_type": "foliar_spray",
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authDelete(t, tok, fmt.Sprintf("/naturalfarming/recipes/%d", recipeID))
	expectStatus(t, resp, http.StatusNoContent)
}

func TestProfileGetAndUpdate(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/profile")
	expectStatus(t, resp, http.StatusOK)
	profile := decodeMap(t, resp)
	if profile["email"] == nil && profile["user_id"] == nil {
		t.Fatalf("profile missing expected fields: %#v", profile)
	}

	resp = authPut(t, tok, "/profile", map[string]any{
		"full_name": "Smoke Test User",
		"timezone":  "America/New_York",
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if updated["full_name"] != "Smoke Test User" {
		t.Fatalf("expected full_name update, got %v", updated["full_name"])
	}
}

func TestScheduleActiveToggle(t *testing.T) {
	tok := smokeJWT(t)

	name := uniqueName("smoke_toggle_sched")
	resp := authPost(t, tok, "/farms/1/schedules", map[string]any{
		"name":            name,
		"schedule_type":   "cron",
		"cron_expression": "0 12 * * *",
		"timezone":        "UTC",
		"is_active":       true,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	schedID := int64(created["id"].(float64))

	resp = authPatch(t, tok, fmt.Sprintf("/schedules/%d/active", schedID), map[string]any{
		"is_active": false,
	})
	expectStatus(t, resp, http.StatusOK)
	toggled := decodeMap(t, resp)
	if toggled["is_active"] != false {
		t.Fatal("expected is_active=false after toggle")
	}

	resp = authPatch(t, tok, fmt.Sprintf("/schedules/%d/active", schedID), map[string]any{
		"is_active": true,
	})
	expectStatus(t, resp, http.StatusOK)

	resp = authGet(t, tok, fmt.Sprintf("/schedules/%d/actuator-events", schedID))
	expectStatus(t, resp, http.StatusOK)
	_ = decodeSlice(t, resp)

	resp = authDelete(t, tok, fmt.Sprintf("/schedules/%d", schedID))
	expectStatus(t, resp, http.StatusNoContent)
}

// ── Phase 20 WS1: Automation Rule CRUD ──────────────────────────────────────

// TestAutomationRuleCRUD exercises the full CRUD surface for
// automation_rules and rule-bound executable_actions, plus the input
// validation the handler layers in front of the DB constraints:
//   - unknown trigger_source rejected at 400
//   - predicates that reference sensors on another farm rejected at 400
//   - deferred action_type values (http_webhook_call etc.) rejected at 400
//   - cascade-delete on automation_rules cleans up child actions and
//     nulls out tasks.source_rule_id rather than deleting the task.
func TestAutomationRuleCRUD(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	// Seed a sensor on farm 1 to use in predicates.
	sensorName := uniqueName("rule_sensor")
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":        sensorName,
		"sensor_type": "moisture",
		"unit_id":     unitID,
	})
	expectStatus(t, resp, http.StatusCreated)
	sensorRow := decodeMap(t, resp)
	sid := int64(sensorRow["id"].(float64))

	// --- Validation: unknown trigger_source is rejected. ---
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":                  uniqueName("rule_bad_trigger"),
		"trigger_source":        "totally-bogus",
		"trigger_configuration": map[string]any{},
		"condition_logic":       "ALL",
		"conditions":            []map[string]any{},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// --- Validation: predicate sensor not on this farm is rejected. ---
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":           uniqueName("rule_foreign_sensor"),
		"trigger_source": "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": 99999999, "op": "gte", "value": 1.0},
		},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// --- Validation: unknown precondition op rejected. ---
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_bad_op"),
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "nope", "value": 1.0},
		},
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// --- Happy path: create a sensor_reading_threshold rule. ---
	ruleName := uniqueName("rule_crud")
	resp = authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":           ruleName,
		"description":    "smoke test rule",
		"is_active":      false,
		"trigger_source": "sensor_reading_threshold",
		"trigger_configuration": map[string]any{
			"sensor_id": sid,
			"op":        "lt",
			"value":     10.0,
		},
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
		"cooldown_period_seconds": 60,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))
	if created["name"] != ruleName {
		t.Fatalf("expected name=%s, got %v", ruleName, created["name"])
	}
	if created["is_active"] != false {
		t.Fatal("expected is_active=false on created rule")
	}

	// GET by id.
	resp = authGet(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID))
	expectStatus(t, resp, http.StatusOK)
	_ = decodeMap(t, resp)

	// List by farm includes it.
	resp = authGet(t, tok, "/farms/1/automation/rules")
	expectStatus(t, resp, http.StatusOK)
	ruleList := decodeSlice(t, resp)
	found := false
	for _, r := range ruleList {
		if m, ok := r.(map[string]any); ok && int64(m["id"].(float64)) == ruleID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("rule %d missing from farm list", ruleID)
	}

	// Toggle active.
	resp = authPatch(t, tok, fmt.Sprintf("/automation/rules/%d/active", ruleID), map[string]any{
		"is_active": true,
	})
	expectStatus(t, resp, http.StatusOK)
	toggled := decodeMap(t, resp)
	if toggled["is_active"] != true {
		t.Fatal("expected is_active=true after toggle")
	}

	// Full update — change cooldown + predicate value.
	resp = authPut(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID), map[string]any{
		"name":           ruleName,
		"is_active":      true,
		"trigger_source": "sensor_reading_threshold",
		"trigger_configuration": map[string]any{
			"sensor_id": sid,
			"op":        "lt",
			"value":     5.0,
		},
		"condition_logic": "ANY",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 5.0},
		},
		"cooldown_period_seconds": 120,
	})
	expectStatus(t, resp, http.StatusOK)
	updated := decodeMap(t, resp)
	if cp, _ := updated["cooldown_period_seconds"].(float64); int(cp) != 120 {
		t.Fatalf("expected cooldown_period_seconds=120, got %v", updated["cooldown_period_seconds"])
	}

	// --- Deferred action types MUST be rejected with 400. ---
	for _, deferred := range []string{
		"trigger_another_automation_rule",
		"http_webhook_call",
		"update_record_in_gr33n",
		"log_custom_event",
	} {
		resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
			"execution_order": 0,
			"action_type":     deferred,
		})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 for deferred action_type=%s, got %d", deferred, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// --- Happy path: attach a create_task action (no actuator needed). ---
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order": 1,
		"action_type":     "create_task",
		"action_parameters": map[string]any{
			"title":    "auto-generated smoke task",
			"priority": 1,
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	action := decodeMap(t, resp)
	actionID := int64(action["id"].(float64))

	// Action missing required shape is rejected.
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":   2,
		"action_type":       "create_task",
		"action_parameters": map[string]any{}, // empty payload
	})
	expectStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()

	// List actions on the rule.
	resp = authGet(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID))
	expectStatus(t, resp, http.StatusOK)
	actionList := decodeSlice(t, resp)
	if len(actionList) == 0 {
		t.Fatal("expected at least one action on the rule")
	}

	// Update the action — bump execution_order.
	resp = authPut(t, tok, fmt.Sprintf("/automation/actions/%d", actionID), map[string]any{
		"execution_order": 5,
		"action_type":     "create_task",
		"action_parameters": map[string]any{
			"title":    "auto-generated smoke task (updated)",
			"priority": 2,
		},
	})
	expectStatus(t, resp, http.StatusOK)

	// --- Seed a task with source_rule_id set to the rule and verify ON DELETE
	// SET NULL behavior when the rule goes away. ---
	resp = authPost(t, tok, "/farms/1/tasks", map[string]any{
		"title":          uniqueName("task_from_rule"),
		"priority":       1,
		"source_rule_id": ruleID,
	})
	expectStatus(t, resp, http.StatusCreated)
	linkedTask := decodeMap(t, resp)
	linkedTaskID := int64(linkedTask["id"].(float64))
	if srid, _ := linkedTask["source_rule_id"].(float64); int64(srid) != ruleID {
		t.Fatalf("expected source_rule_id=%d, got %v", ruleID, linkedTask["source_rule_id"])
	}

	// Delete the rule — cascades to actions, nulls source_rule_id on tasks.
	resp = authDelete(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID))
	expectStatus(t, resp, http.StatusNoContent)

	// Rule is gone.
	resp = authGet(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID))
	expectStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()

	// Child action was cascade-deleted.
	var actionCount int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.executable_actions WHERE id = $1`, actionID,
	).Scan(&actionCount); err != nil {
		t.Fatalf("count actions after rule delete: %v", err)
	}
	if actionCount != 0 {
		t.Fatalf("expected 0 actions after rule delete, got %d", actionCount)
	}

	// Task still exists but with source_rule_id nulled out.
	var nullCount int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.tasks WHERE id = $1 AND source_rule_id IS NULL`,
		linkedTaskID,
	).Scan(&nullCount); err != nil {
		t.Fatalf("check task source_rule_id after rule delete: %v", err)
	}
	if nullCount != 1 {
		t.Fatalf("expected task %d to remain with source_rule_id=NULL, got %d", linkedTaskID, nullCount)
	}
}

// ── Phase 20 WS2: Rule evaluator ────────────────────────────────────────────

// seedRuleSensorWithReading creates a sensor on farm 1 and seeds a single
// reading. Returns the sensor id. Test helper for WS2 rule tick tests.
func seedRuleSensorWithReading(t *testing.T, tok string, unitID int64, value float64) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/sensors", map[string]any{
		"name":        uniqueName("rule_tick_sensor"),
		"sensor_type": "moisture",
		"unit_id":     unitID,
	})
	expectStatus(t, resp, http.StatusCreated)
	row := decodeMap(t, resp)
	sid := int64(row["id"].(float64))
	if _, err := testPool.Exec(context.Background(), `
		INSERT INTO gr33ncore.sensor_readings (reading_time, sensor_id, value_raw, is_valid)
		VALUES (NOW(), $1, $2, TRUE)`, sid, value); err != nil {
		t.Fatalf("seed reading for sensor %d: %v", sid, err)
	}
	return sid
}

// TestAutomationRuleTickALLvsANY verifies the rule evaluator honors
// condition_logic. One predicate passes (sensor reads 5, predicate lt 10)
// and one fails (other sensor reads 50, predicate lt 10). Under ALL the
// rule must skip with message=conditions_not_met; under ANY the rule must
// fire (success). last_evaluated_time must be stamped in both cases.
func TestAutomationRuleTickALLvsANY(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	// Sensor A reads 5 (predicate lt 10 → passes).
	// Sensor B reads 50 (predicate lt 10 → fails).
	sidA := seedRuleSensorWithReading(t, tok, unitID, 5)
	sidB := seedRuleSensorWithReading(t, tok, unitID, 50)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_all_any"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sidA, "op": "lt", "value": 10.0},
			{"sensor_id": sidB, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))

	// --- ALL: must skip with conditions_not_met. ---
	testWorker.TickRules(ctx)

	var status, msg string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, ''), details::text
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("read latest rule run (ALL): %v", err)
	}
	if status != "skipped" || msg != "conditions_not_met" {
		t.Fatalf("expected ALL tick to skip with conditions_not_met, got status=%s msg=%s", status, msg)
	}
	var details struct {
		Phase         string `json:"phase"`
		Logic         string `json:"logic"`
		ConditionsMet bool   `json:"conditions_met"`
		Failed        []struct {
			SensorID int64   `json:"sensor_id"`
			Reason   string  `json:"reason"`
			Expected float64 `json:"expected"`
		} `json:"failed"`
	}
	if err := json.Unmarshal(detailsJSON, &details); err != nil {
		t.Fatalf("parse details: %v (raw=%s)", err, detailsJSON)
	}
	if details.Phase != "conditions" || details.Logic != "ALL" || details.ConditionsMet {
		t.Fatalf("unexpected details on ALL skip: %+v", details)
	}
	if len(details.Failed) != 1 || details.Failed[0].SensorID != sidB {
		t.Fatalf("expected exactly sensor B (%d) to be in failed list, got %+v", sidB, details.Failed)
	}

	// last_evaluated_time must be stamped even on skip.
	var lastEval *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_evaluated_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&lastEval); err != nil {
		t.Fatalf("read last_evaluated_time: %v", err)
	}
	if lastEval == nil {
		t.Fatal("expected last_evaluated_time to be set after a skip tick")
	}
	// Must NOT fire: last_triggered_time stays NULL.
	var lastTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&lastTrig); err != nil {
		t.Fatalf("read last_triggered_time: %v", err)
	}
	if lastTrig != nil {
		t.Fatalf("expected last_triggered_time to stay NULL after ALL skip, got %v", *lastTrig)
	}

	// --- Flip to ANY: must fire. ---
	resp = authPut(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID), map[string]any{
		"name":            created["name"],
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ANY",
		"conditions": []map[string]any{
			{"sensor_id": sidA, "op": "lt", "value": 10.0},
			{"sensor_id": sidB, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, '')
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg); err != nil {
		t.Fatalf("read latest rule run (ANY): %v", err)
	}
	// No actions attached → the evaluator records a skipped run with
	// message="rule has no executable actions" after conditions met.
	// Either way the fire path ran: last_triggered_time MUST be set.
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&lastTrig); err != nil {
		t.Fatalf("read last_triggered_time after ANY tick: %v", err)
	}
	if lastTrig == nil {
		t.Fatal("expected last_triggered_time to be stamped after ANY tick with conditions met")
	}
	if status == "skipped" && msg == "conditions_not_met" {
		t.Fatalf("ANY tick incorrectly reported conditions_not_met")
	}
}

// TestAutomationRuleTickCooldown verifies cooldown_period_seconds:
//  1. First tick satisfies conditions → fires (last_triggered_time set).
//  2. Second tick within the cooldown window → skipped with message=cooldown
//     and last_triggered_time NOT advanced.
//  3. After rolling last_triggered_time back past the cooldown window, the
//     next tick fires again.
func TestAutomationRuleTickCooldown(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_cooldown"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
		"cooldown_period_seconds": 300,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))

	// Tick 1: fires.
	testWorker.TickRules(ctx)
	var firstTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&firstTrig); err != nil {
		t.Fatalf("read last_triggered_time after fire: %v", err)
	}
	if firstTrig == nil {
		t.Fatal("expected last_triggered_time to be set after first tick")
	}

	// Tick 2: in cooldown, must skip with message=cooldown.
	testWorker.TickRules(ctx)
	var status, msg string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, ''), details::text
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("read latest rule run after second tick: %v", err)
	}
	if status != "skipped" || msg != "cooldown" {
		t.Fatalf("expected second tick to skip with cooldown, got status=%s msg=%s", status, msg)
	}
	var cdDetails struct {
		Phase            string `json:"phase"`
		CooldownSeconds  int    `json:"cooldown_seconds"`
		RemainingSeconds int    `json:"remaining_seconds"`
	}
	if err := json.Unmarshal(detailsJSON, &cdDetails); err != nil {
		t.Fatalf("parse cooldown details: %v", err)
	}
	if cdDetails.Phase != "cooldown" || cdDetails.CooldownSeconds != 300 {
		t.Fatalf("unexpected cooldown details: %+v", cdDetails)
	}

	// last_triggered_time must not have moved forward on a cooldown skip.
	var secondTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&secondTrig); err != nil {
		t.Fatalf("re-read last_triggered_time: %v", err)
	}
	if secondTrig == nil || !secondTrig.Equal(*firstTrig) {
		t.Fatalf("expected last_triggered_time to stay %v after cooldown skip, got %v", firstTrig, secondTrig)
	}

	// Roll last_triggered_time back past the cooldown window. After this the
	// next tick MUST fire again.
	if _, err := testPool.Exec(ctx,
		`UPDATE gr33ncore.automation_rules SET last_triggered_time = NOW() - INTERVAL '10 minutes' WHERE id = $1`,
		ruleID,
	); err != nil {
		t.Fatalf("rewind last_triggered_time: %v", err)
	}

	testWorker.TickRules(ctx)
	var thirdTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&thirdTrig); err != nil {
		t.Fatalf("read last_triggered_time after third tick: %v", err)
	}
	if thirdTrig == nil {
		t.Fatal("expected last_triggered_time to be set after third tick")
	}
	if !thirdTrig.After(*firstTrig) {
		t.Fatalf("expected third-tick last_triggered_time (%v) to advance past first-tick (%v) after cooldown window elapses", thirdTrig, firstTrig)
	}
}

// TestAutomationRuleTickInactiveRuleSkipped verifies ListActiveAutomationRules
// does not return rules with is_active=false, so the evaluator never touches
// them. Negative bookkeeping test — ensures last_evaluated_time STAYS null
// for an inactive rule even after several ticks.
func TestAutomationRuleTickInactiveRuleSkipped(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_inactive"),
		"is_active":       false,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))

	testWorker.TickRules(ctx)
	testWorker.TickRules(ctx)

	var lastEval, lastTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_evaluated_time, last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`,
		ruleID,
	).Scan(&lastEval, &lastTrig); err != nil {
		t.Fatalf("read rule times: %v", err)
	}
	if lastEval != nil || lastTrig != nil {
		t.Fatalf("expected last_evaluated_time and last_triggered_time to stay NULL for inactive rule, got eval=%v trig=%v", lastEval, lastTrig)
	}
	var runCount int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.automation_runs WHERE rule_id = $1`, ruleID,
	).Scan(&runCount); err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if runCount != 0 {
		t.Fatalf("expected 0 runs for inactive rule, got %d", runCount)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Phase 20 / WS3 — rule action dispatchers
//
// These tests drive the worker through TickRules and assert the DB side
// effects of each supported `executable_action_type`. The worker runs in
// simulation mode in the test bootstrap, so `control_actuator` dispatches
// go through the "simulated" leg (status = execution_completed_success_on_device,
// no device pending command).
// ─────────────────────────────────────────────────────────────────────────────

// seedRuleActuator inserts a bare actuator row on farm 1 and returns its id.
// Smoke tests don't have a POST /actuators endpoint, so we fabricate one via
// direct SQL. Device/zone are both nullable on the schema, so this minimal row
// is enough for the worker to dispatch against.
func seedRuleActuator(t *testing.T, name string) int64 {
	t.Helper()
	var id int64
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO gr33ncore.actuators (farm_id, name, actuator_type)
		VALUES (1, $1, 'relay')
		RETURNING id`, name).Scan(&id); err != nil {
		t.Fatalf("seed actuator: %v", err)
	}
	return id
}

// seedRuleNotificationTemplate inserts a per-farm notification template and
// returns its id. The rule evaluator resolves the template by id and uses
// its subject/body for the rendered alerts_notifications row.
func seedRuleNotificationTemplate(t *testing.T, key, subject, body string, priority string) int64 {
	t.Helper()
	var id int64
	if err := testPool.QueryRow(context.Background(), `
		INSERT INTO gr33ncore.notification_templates
		  (farm_id, template_key, subject_template, body_template_text, default_priority)
		VALUES (1, $1, $2, $3, $4::gr33ncore.notification_priority_enum)
		RETURNING id`, key, subject, body, priority).Scan(&id); err != nil {
		t.Fatalf("seed notification template: %v", err)
	}
	return id
}

// TestAutomationRuleDispatchControlActuator verifies the control_actuator
// dispatcher:
//  1. One tick with conditions met writes a gr33ncore.actuator_events row
//     whose `triggered_by_rule_id` is the rule id and `source` is
//     'automation_rule_trigger'.
//  2. The run is recorded as status=success with actions_total=actions_success=1.
//  3. Rule `last_triggered_time` advances.
func TestAutomationRuleDispatchControlActuator(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	sid := seedRuleSensorWithReading(t, tok, unitID, 5)
	actID := seedRuleActuator(t, uniqueName("ws3_actuator"))

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws3_actuator"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":    0,
		"action_type":        "control_actuator",
		"target_actuator_id": actID,
		"action_command":     "on",
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	// Run bookkeeping: status=success, actions_success=1.
	var status, msg string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, ''), details::text
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("read latest rule run: %v", err)
	}
	if status != "success" {
		t.Fatalf("expected success, got status=%s msg=%s details=%s", status, msg, detailsJSON)
	}
	var details struct {
		Phase          string `json:"phase"`
		ActionsTotal   int    `json:"actions_total"`
		ActionsSuccess int    `json:"actions_success"`
	}
	if err := json.Unmarshal(detailsJSON, &details); err != nil {
		t.Fatalf("parse details: %v", err)
	}
	if details.Phase != "actions" || details.ActionsTotal != 1 || details.ActionsSuccess != 1 {
		t.Fatalf("unexpected details: %+v (raw=%s)", details, detailsJSON)
	}

	// Side effect: one actuator_events row stamped with this rule.
	var eventCount int
	var eventSource, commandSent string
	if err := testPool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(MAX(source::text), ''), COALESCE(MAX(command_sent), '')
		FROM gr33ncore.actuator_events
		WHERE triggered_by_rule_id = $1 AND actuator_id = $2`,
		ruleID, actID,
	).Scan(&eventCount, &eventSource, &commandSent); err != nil {
		t.Fatalf("count actuator events: %v", err)
	}
	if eventCount != 1 {
		t.Fatalf("expected exactly 1 actuator_events row for rule %d, got %d", ruleID, eventCount)
	}
	if eventSource != "automation_rule_trigger" {
		t.Fatalf("expected source=automation_rule_trigger, got %s", eventSource)
	}
	if commandSent != "on" {
		t.Fatalf("expected command_sent=on, got %s", commandSent)
	}

	// last_triggered_time advanced.
	var lastTrig *time.Time
	if err := testPool.QueryRow(ctx,
		`SELECT last_triggered_time FROM gr33ncore.automation_rules WHERE id = $1`, ruleID,
	).Scan(&lastTrig); err != nil {
		t.Fatalf("read last_triggered_time: %v", err)
	}
	if lastTrig == nil {
		t.Fatal("expected last_triggered_time to be stamped after successful dispatch")
	}
}

// TestAutomationRuleDispatchCreateTask verifies the create_task dispatcher:
//  1. A ticked rule inserts a task with source_rule_id pointing back at the rule.
//  2. action_parameters.{title,priority,due_in_days} are honored.
//  3. The run is recorded success with actions_success=1.
func TestAutomationRuleDispatchCreateTask(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}

	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws3_task"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	taskTitle := uniqueName("ws3_task_title")
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order": 0,
		"action_type":     "create_task",
		"action_parameters": map[string]any{
			"title":       taskTitle,
			"priority":    2,
			"due_in_days": 1,
			"task_type":   "inspection",
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	var status string
	if err := testPool.QueryRow(ctx,
		`SELECT status FROM gr33ncore.automation_runs WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&status); err != nil {
		t.Fatalf("read run: %v", err)
	}
	if status != "success" {
		t.Fatalf("expected run status=success, got %s", status)
	}

	// The generated task carries source_rule_id and our parameters.
	var gotTitle, gotType string
	var gotPriority int32
	var gotDue *time.Time
	var gotSourceRuleID *int64
	if err := testPool.QueryRow(ctx, `
		SELECT title, COALESCE(task_type, ''), COALESCE(priority, 0), due_date, source_rule_id
		FROM gr33ncore.tasks
		WHERE source_rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&gotTitle, &gotType, &gotPriority, &gotDue, &gotSourceRuleID); err != nil {
		t.Fatalf("read generated task: %v", err)
	}
	if gotTitle != taskTitle {
		t.Fatalf("expected task title %q, got %q", taskTitle, gotTitle)
	}
	if gotType != "inspection" {
		t.Fatalf("expected task_type=inspection, got %q", gotType)
	}
	if gotPriority != 2 {
		t.Fatalf("expected priority=2, got %d", gotPriority)
	}
	if gotDue == nil {
		t.Fatal("expected due_date to be set from due_in_days=1")
	}
	if gotSourceRuleID == nil || *gotSourceRuleID != ruleID {
		t.Fatalf("expected source_rule_id=%d, got %v", ruleID, gotSourceRuleID)
	}
}

// TestAutomationRuleDispatchSendNotification verifies the send_notification
// dispatcher:
//  1. The template's subject/body are rendered into alerts_notifications.
//  2. notification_template_id and triggering_event_source_type='automation_rule'
//     are set on the inserted alert.
//  3. Severity defaults to the template's default_priority.
func TestAutomationRuleDispatchSendNotification(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	tmplKey := uniqueName("ws3_tmpl")
	tmplID := seedRuleNotificationTemplate(t,
		tmplKey,
		"Alert from rule {{rule_name}}",
		"Sensor reading triggered rule {{rule_id}}",
		"high",
	)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws3_notify"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	ruleID := int64(created["id"].(float64))
	ruleName := created["name"].(string)

	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":                 0,
		"action_type":                     "send_notification",
		"target_notification_template_id": tmplID,
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	var status string
	if err := testPool.QueryRow(ctx,
		`SELECT status FROM gr33ncore.automation_runs WHERE rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&status); err != nil {
		t.Fatalf("read run: %v", err)
	}
	if status != "success" {
		t.Fatalf("expected run status=success, got %s", status)
	}

	var subject, body, srcType, severity string
	var gotTmpl, gotSrcID *int64
	if err := testPool.QueryRow(ctx, `
		SELECT COALESCE(subject_rendered, ''), COALESCE(message_text_rendered, ''),
		       COALESCE(triggering_event_source_type, ''), severity::text,
		       notification_template_id, triggering_event_source_id
		FROM gr33ncore.alerts_notifications
		WHERE notification_template_id = $1 AND triggering_event_source_id = $2
		ORDER BY id DESC LIMIT 1`, tmplID, ruleID,
	).Scan(&subject, &body, &srcType, &severity, &gotTmpl, &gotSrcID); err != nil {
		t.Fatalf("read alert: %v", err)
	}
	expectedSubject := "Alert from rule " + ruleName
	if subject != expectedSubject {
		t.Fatalf("expected subject %q, got %q", expectedSubject, subject)
	}
	expectedBody := fmt.Sprintf("Sensor reading triggered rule %d", ruleID)
	if body != expectedBody {
		t.Fatalf("expected body %q, got %q", expectedBody, body)
	}
	if srcType != "automation_rule" {
		t.Fatalf("expected triggering_event_source_type=automation_rule, got %s", srcType)
	}
	if severity != "high" {
		t.Fatalf("expected severity=high (from template default_priority), got %s", severity)
	}
	if gotTmpl == nil || *gotTmpl != tmplID {
		t.Fatalf("expected notification_template_id=%d, got %v", tmplID, gotTmpl)
	}
	if gotSrcID == nil || *gotSrcID != ruleID {
		t.Fatalf("expected triggering_event_source_id=%d, got %v", ruleID, gotSrcID)
	}

	// The worker also fans the alert through the push pipeline. The
	// test wires a recording PushNotifier that captures every dispatched
	// alert — one rule fire should produce exactly one push dispatch
	// stamped with this rule's id.
	if got := testNotifier.countForRule(ruleID); got != 1 {
		t.Fatalf("expected push notifier to receive 1 alert for rule %d, got %d", ruleID, got)
	}
}

// TestAutomationRuleDispatchPartialSuccess verifies that when a rule has
// multiple actions and one of them fails at dispatch time, the run is
// recorded as `partial_success` with details.errors[] populated, and
// the successful action's side effect still lands.
//
// We fabricate the failure by direct-inserting a deferred-type action
// (log_custom_event) into executable_actions — the API CRUD validator
// rejects these, but the DB CHECK constraint permits them, so this is
// the realistic "row written by a newer binary, read by a worker that
// doesn't know that action type" path. The sibling `create_task`
// action still runs and records its side effect.
func TestAutomationRuleDispatchPartialSuccess(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws3_partial"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	// Direct-insert a deferred-type action the worker doesn't know about.
	// Bypasses the CRUD validator (which would 400) but respects the DB
	// CHECK constraint that log_custom_event needs action_parameters.
	var brokenActionID int64
	if err := testPool.QueryRow(ctx, `
		INSERT INTO gr33ncore.executable_actions
		  (rule_id, execution_order, action_type, action_parameters)
		VALUES ($1, 0, 'log_custom_event', '{"note":"forced failure"}'::jsonb)
		RETURNING id`, ruleID,
	).Scan(&brokenActionID); err != nil {
		t.Fatalf("seed deferred action: %v", err)
	}

	taskTitle := uniqueName("ws3_partial_task")
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order":   1,
		"action_type":       "create_task",
		"action_parameters": map[string]any{"title": taskTitle},
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	var status, msg string
	var detailsJSON []byte
	if err := testPool.QueryRow(ctx, `
		SELECT status, COALESCE(message, ''), details::text
		FROM gr33ncore.automation_runs
		WHERE rule_id = $1
		ORDER BY id DESC LIMIT 1`, ruleID).Scan(&status, &msg, &detailsJSON); err != nil {
		t.Fatalf("read run: %v", err)
	}
	if status != "partial_success" {
		t.Fatalf("expected partial_success, got status=%s msg=%s details=%s", status, msg, detailsJSON)
	}
	var details struct {
		ActionsTotal   int `json:"actions_total"`
		ActionsSuccess int `json:"actions_success"`
		Errors         []struct {
			ActionID int64  `json:"action_id"`
			Message  string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(detailsJSON, &details); err != nil {
		t.Fatalf("parse details: %v", err)
	}
	if details.ActionsTotal != 2 || details.ActionsSuccess != 1 {
		t.Fatalf("expected 2 total / 1 success, got %+v", details)
	}
	if len(details.Errors) != 1 || details.Errors[0].ActionID != brokenActionID {
		t.Fatalf("expected single error for action %d, got %+v", brokenActionID, details.Errors)
	}

	// The create_task action still landed its task.
	var gotTitle string
	if err := testPool.QueryRow(ctx,
		`SELECT title FROM gr33ncore.tasks WHERE source_rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&gotTitle); err != nil {
		t.Fatalf("read generated task: %v", err)
	}
	if gotTitle != taskTitle {
		t.Fatalf("expected task %q from successful sibling action, got %q", taskTitle, gotTitle)
	}
}

// TestAutomationRuleDeleteNullsTaskSourceRuleID verifies the task-provenance
// invariant from Phase 20 WS1: deleting a rule that previously generated
// tasks leaves those tasks in place but nulls out `source_rule_id`, so the
// audit trail is preserved even when the originating rule is gone.
//
// Rule of thumb: "the task was real work, even if the rule that spawned it
// no longer exists." The FK uses ON DELETE SET NULL, not CASCADE.
func TestAutomationRuleDeleteNullsTaskSourceRuleID(t *testing.T) {
	if testPool == nil || testWorker == nil {
		t.Skip("testPool/testWorker unavailable")
	}
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var unitID int64
	if err := testPool.QueryRow(ctx, `SELECT id FROM gr33ncore.units LIMIT 1`).Scan(&unitID); err != nil {
		t.Fatalf("find a unit id: %v", err)
	}
	sid := seedRuleSensorWithReading(t, tok, unitID, 5)

	resp := authPost(t, tok, "/farms/1/automation/rules", map[string]any{
		"name":            uniqueName("rule_ws5_cascade"),
		"is_active":       true,
		"trigger_source":  "manual_api_trigger",
		"condition_logic": "ALL",
		"conditions": []map[string]any{
			{"sensor_id": sid, "op": "lt", "value": 10.0},
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	ruleID := int64(decodeMap(t, resp)["id"].(float64))

	taskTitle := uniqueName("ws5_cascade_task")
	resp = authPost(t, tok, fmt.Sprintf("/automation/rules/%d/actions", ruleID), map[string]any{
		"execution_order": 0,
		"action_type":     "create_task",
		"action_parameters": map[string]any{
			"title":     taskTitle,
			"task_type": "inspection",
		},
	})
	expectStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	testWorker.TickRules(ctx)

	// Capture the task id while the rule is still alive so we can re-check
	// the same row after the delete — we care that the row survives with
	// source_rule_id NULLed, not that it was replaced.
	var taskID int64
	if err := testPool.QueryRow(ctx,
		`SELECT id FROM gr33ncore.tasks WHERE source_rule_id = $1 ORDER BY id DESC LIMIT 1`, ruleID,
	).Scan(&taskID); err != nil {
		t.Fatalf("locate generated task: %v", err)
	}

	resp = authDelete(t, tok, fmt.Sprintf("/automation/rules/%d", ruleID))
	expectStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()

	// The task must still exist, but with source_rule_id cleared.
	var postDeleteSourceRuleID *int64
	var postDeleteTitle string
	if err := testPool.QueryRow(ctx,
		`SELECT title, source_rule_id FROM gr33ncore.tasks WHERE id = $1`, taskID,
	).Scan(&postDeleteTitle, &postDeleteSourceRuleID); err != nil {
		t.Fatalf("re-read task %d after rule delete: %v", taskID, err)
	}
	if postDeleteTitle != taskTitle {
		t.Fatalf("expected task to survive with title %q, got %q", taskTitle, postDeleteTitle)
	}
	if postDeleteSourceRuleID != nil {
		t.Fatalf("expected source_rule_id to be NULL after parent rule delete, got %d", *postDeleteSourceRuleID)
	}

	// Sanity: the executable_actions row is gone (CASCADE), so the rule
	// really was torn down, not just "soft-hidden".
	var actionsLeft int
	if err := testPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gr33ncore.executable_actions WHERE rule_id = $1`, ruleID,
	).Scan(&actionsLeft); err != nil {
		t.Fatalf("count actions: %v", err)
	}
	if actionsLeft != 0 {
		t.Fatalf("expected rule's actions to be cascaded away, got %d left", actionsLeft)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func get(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(testServer.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func authGet(t *testing.T, token, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, testServer.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func authPost(t *testing.T, token, path string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, testServer.URL+path, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func authPatch(t *testing.T, token, path string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPatch, testServer.URL+path, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH %s: %v", path, err)
	}
	return resp
}

func authPut(t *testing.T, token, path string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPut, testServer.URL+path, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT %s: %v", path, err)
	}
	return resp
}

func authDelete(t *testing.T, token, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, testServer.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s: %v", path, err)
	}
	return resp
}

func authDeleteJSON(t *testing.T, token, path string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodDelete, testServer.URL+path, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s: %v", path, err)
	}
	return resp
}

func authMultipartPost(t *testing.T, token, path, fieldName, fileName, contentType string, fileBody []byte, fields map[string]string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	partHeaders := make(textproto.MIMEHeader)
	partHeaders.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName))
	partHeaders.Set("Content-Type", contentType)
	part, err := w.CreatePart(partHeaders)
	if err != nil {
		t.Fatalf("CreatePart: %v", err)
	}
	if _, err := part.Write(fileBody); err != nil {
		t.Fatalf("part.Write: %v", err)
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("WriteField(%s): %v", k, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, testServer.URL+path, &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func TestFarmBootstrapOnCreate(t *testing.T) {
	tok := smokeJWT(t)
	name := uniqueName("bootstrap_farm")
	resp := authPost(t, tok, "/farms", map[string]any{
		"name":               name,
		"owner_user_id":      smokeDevUserUUID,
		"timezone":           "UTC",
		"currency":           "USD",
		"operational_status": "active",
		"scale_tier":         "small",
		"bootstrap_template": "jadam_indoor_photoperiod_v1",
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	farmObj, ok := payload["farm"].(map[string]any)
	if !ok {
		t.Fatalf("expected farm in response, got %v", payload)
	}
	fid := int64(farmObj["id"].(float64))
	zones := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/zones", fid)))
	if len(zones) < 4 {
		t.Fatalf("expected at least 4 zones from bootstrap, got %d", len(zones))
	}
	resp2 := authPost(t, tok, fmt.Sprintf("/farms/%d/bootstrap-template", fid), map[string]any{
		"template": "jadam_indoor_photoperiod_v1",
	})
	expectStatus(t, resp2, http.StatusOK)
	again := decodeMap(t, resp2)
	boot := again["bootstrap"].(map[string]any)
	if applied, _ := boot["already_applied"].(bool); !applied {
		t.Fatalf("expected already_applied on second template apply, got %#v", boot)
	}
}

// TestPhase205BootstrapTemplates verifies Phase 20.5 WS2 farm bootstrap keys
// (chicken_coop_v1, greenhouse_climate_v1, drying_room_v1, small_aquaponics_v1)
// each land the expected zones, sensors, automation rules, and (for aquaponics)
// a gr33naquaponics.loops row.
func TestPhase205BootstrapTemplates(t *testing.T) {
	tok := smokeJWT(t)
	cases := []struct {
		key           string
		wantZones     []string
		minRules      int
		wantLoopLabel string
	}{
		{
			key:       "chicken_coop_v1",
			wantZones: []string{"Chicken Coop"},
			minRules:  4,
		},
		{
			key:       "greenhouse_climate_v1",
			wantZones: []string{"Greenhouse"},
			minRules:  4,
		},
		{
			key:       "drying_room_v1",
			wantZones: []string{"Drying Room"},
			minRules:  3,
		},
		{
			key:           "small_aquaponics_v1",
			wantZones:     []string{"Fish Tank", "Grow Bed"},
			minRules:      2,
			wantLoopLabel: "Main aquaponics loop",
		},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			name := uniqueName("farm_" + tc.key)
			resp := authPost(t, tok, "/farms", map[string]any{
				"name":               name,
				"owner_user_id":      smokeDevUserUUID,
				"timezone":           "UTC",
				"currency":           "USD",
				"operational_status": "active",
				"scale_tier":         "small",
				"bootstrap_template": tc.key,
			})
			expectStatus(t, resp, http.StatusCreated)
			payload := decodeMap(t, resp)
			farmObj := payload["farm"].(map[string]any)
			fid := int64(farmObj["id"].(float64))
			boot := payload["bootstrap"].(map[string]any)
			if errStr, _ := boot["error"].(string); errStr != "" {
				t.Fatalf("bootstrap error for %s: %s — %#v", tc.key, errStr, boot)
			}
			if applied, _ := boot["applied"].(bool); !applied {
				t.Fatalf("expected applied=true for %s, got %#v", tc.key, boot)
			}

			zones := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/zones", fid)))
			zoneNames := map[string]struct{}{}
			for _, z := range zones {
				if m, ok := z.(map[string]any); ok {
					if n, ok := m["name"].(string); ok {
						zoneNames[n] = struct{}{}
					}
				}
			}
			for _, wz := range tc.wantZones {
				if _, ok := zoneNames[wz]; !ok {
					t.Fatalf("farm %d template %s: missing zone %q (have %v)", fid, tc.key, wz, zoneNames)
				}
			}

			rules := decodeSlice(t, authGet(t, tok, fmt.Sprintf("/farms/%d/automation/rules", fid)))
			if len(rules) < tc.minRules {
				t.Fatalf("farm %d template %s: expected at least %d rules, got %d", fid, tc.key, tc.minRules, len(rules))
			}

			if tc.wantLoopLabel != "" && testPool != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				var cnt int
				err := testPool.QueryRow(ctx,
					`SELECT COUNT(*) FROM gr33naquaponics.loops WHERE farm_id = $1 AND label = $2 AND deleted_at IS NULL`,
					fid, tc.wantLoopLabel,
				).Scan(&cnt)
				if err != nil {
					t.Fatalf("count loops: %v", err)
				}
				if cnt != 1 {
					t.Fatalf("expected 1 aquaponics loop %q for farm %d, got %d", tc.wantLoopLabel, fid, cnt)
				}
			}
		})
	}
}

func createSmokeCost(t *testing.T, tok string) int64 {
	t.Helper()
	resp := authPost(t, tok, "/farms/1/costs", map[string]any{
		"transaction_date": "2026-04-16",
		"category":         "miscellaneous",
		"amount":           12.5,
		"currency":         "USD",
		"description":      "receipt smoke test",
		"is_income":        false,
	})
	expectStatus(t, resp, http.StatusCreated)
	created := decodeMap(t, resp)
	return int64(created["id"].(float64))
}

func uploadSmokeReceipt(t *testing.T, tok string, costID int64, fileName string, body []byte) int64 {
	t.Helper()
	resp := authMultipartPost(t, tok, "/farms/1/cost-receipts", "file", fileName, "application/pdf", body, map[string]string{
		"cost_transaction_id": fmt.Sprintf("%d", costID),
	})
	expectStatus(t, resp, http.StatusCreated)
	payload := decodeMap(t, resp)
	attachment := payload["file_attachment"].(map[string]any)
	return int64(attachment["id"].(float64))
}

func postNoAuth(path string, body any) *http.Response {
	b, _ := json.Marshal(body)
	resp, err := http.Post(testServer.URL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	return resp
}

func expectStatus(t *testing.T, resp *http.Response, code int) {
	t.Helper()
	if resp.StatusCode != code {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d: %s", code, resp.StatusCode, string(b))
	}
}

func decodeMap(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var m map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatalf("failed to decode JSON map: %v", err)
	}
	return m
}

func decodeSlice(t *testing.T, resp *http.Response) []any {
	t.Helper()
	defer resp.Body.Close()
	var s []any
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		t.Fatalf("failed to decode JSON slice: %v", err)
	}
	return s
}
