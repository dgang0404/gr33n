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
	testServer *httptest.Server

	smokeTokenOnce sync.Once
	smokeToken     string
	smokeTokenErr  error
)

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
	worker := automationworker.NewWorker(pool, true)
	store, err := filestorage.NewLocal(filepath.Join(smokeFiles, "blobs"))
	if err != nil {
		pool.Close()
		fmt.Fprintf(os.Stderr, "smoke_test init storage: %v\n", err)
		os.Exit(1)
	}
	registerRoutes(mux, pool, worker, "admin", nil, "", store, filestorage.Config{Backend: "local"})
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
