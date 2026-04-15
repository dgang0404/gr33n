package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	automationworker "gr33n-api/internal/automation"
)

const (
	smokeDevEmail    = "dev@gr33n.local"
	smokeDevPass = "devpassword"
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

	mux := http.NewServeMux()
	worker := automationworker.NewWorker(pool, true)
	registerRoutes(mux, pool, worker, "admin", nil, "")
	testServer = httptest.NewServer(corsMiddleware(mux))

	code := m.Run()
	testServer.Close()
	pool.Close()
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
		"zone_id":      1,
		"name":         name,
		"current_stage": "early_veg",
		"started_at":   "2025-01-01",
		"is_active":    false,
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
	if !strings.HasPrefix(string(body), "date,category,amount") {
		t.Fatalf("expected CSV header, got %q", string(body[:min(40, len(body))]))
	}
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
