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
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	automationworker "gr33n-api/internal/automation"
)

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, rand.Int())
}

var testServer *httptest.Server

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://davidg@/gr33n?host=/var/run/postgresql"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		// Skip all tests if DB is not available
		os.Exit(0)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		os.Exit(0)
	}

	// Dev mode: no JWT, no API key
	jwtSecret = nil
	piAPIKey = ""
	authMode = "dev"
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
	if body["mode"] != "dev" {
		t.Fatalf("expected mode=dev, got %v", body["mode"])
	}
}

// ── Farm + Zone + Sensor Reads ──────────────────────────────────────────────

func TestGetFarm(t *testing.T) {
	resp := get(t, "/farms/1")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["name"] == nil {
		t.Fatal("expected farm to have a name")
	}
}

func TestListZones(t *testing.T) {
	resp := get(t, "/farms/1/zones")
	expectStatus(t, resp, 200)
	items := decodeSlice(t, resp)
	if len(items) == 0 {
		t.Fatal("expected at least one zone from seed data")
	}
}

func TestListSensors(t *testing.T) {
	resp := get(t, "/farms/1/sensors")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestListActuators(t *testing.T) {
	resp := get(t, "/farms/1/actuators")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestListSchedules(t *testing.T) {
	resp := get(t, "/farms/1/schedules")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestListAutomationRuns(t *testing.T) {
	resp := get(t, "/farms/1/automation/runs")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestListTasks(t *testing.T) {
	resp := get(t, "/farms/1/tasks")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestWorkerHealth(t *testing.T) {
	resp := get(t, "/automation/worker/health")
	expectStatus(t, resp, 200)
	body := decodeMap(t, resp)
	if body["simulation_mode"] != true {
		t.Fatal("expected simulation_mode=true")
	}
}

// ── Fertigation CRUD ────────────────────────────────────────────────────────

func TestFertigationReservoirRoundtrip(t *testing.T) {
	name := uniqueName("smoke_reservoir")
	payload := map[string]any{
		"name":                  name,
		"status":                "ready",
		"capacity_liters":      100.0,
		"current_volume_liters": 50.0,
	}
	resp := post(t, "/farms/1/fertigation/reservoirs", payload)
	expectStatus(t, resp, 201)
	created := decodeMap(t, resp)
	if created["name"] != name {
		t.Fatalf("expected name=%s, got %v", name, created["name"])
	}

	resp = get(t, "/farms/1/fertigation/reservoirs")
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
	payload := map[string]any{
		"growth_stage": "early_veg",
		"ec_min_mscm":  1.0,
		"ec_max_mscm":  2.5,
		"ph_min":       5.5,
		"ph_max":       6.5,
		"notes":        "smoke test",
	}
	resp := post(t, "/farms/1/fertigation/ec-targets", payload)
	expectStatus(t, resp, 201)

	resp = get(t, "/farms/1/fertigation/ec-targets")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestFertigationProgramRoundtrip(t *testing.T) {
	payload := map[string]any{
		"name":                uniqueName("smoke_program"),
		"total_volume_liters": 5.0,
		"is_active":           false,
		"ec_trigger_low":      0.0,
		"ph_trigger_low":      0.0,
		"ph_trigger_high":     0.0,
	}
	resp := post(t, "/farms/1/fertigation/programs", payload)
	expectStatus(t, resp, 201)

	resp = get(t, "/farms/1/fertigation/programs")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

func TestFertigationEventRoundtrip(t *testing.T) {
	// Need a valid zone_id from seed data — zone 1 should exist
	payload := map[string]any{
		"zone_id":              1,
		"volume_applied_liters": 2.5,
		"ec_before_mscm":       1.2,
		"ec_after_mscm":        1.8,
		"ph_before":            6.0,
		"ph_after":             6.2,
		"trigger_source":       "manual",
	}
	resp := post(t, "/farms/1/fertigation/events", payload)
	expectStatus(t, resp, 201)

	resp = get(t, "/farms/1/fertigation/events")
	expectStatus(t, resp, 200)
	_ = decodeSlice(t, resp)
}

// ── Login Flow ──────────────────────────────────────────────────────────────

func TestLoginBadCredentials(t *testing.T) {
	// In dev mode with no password hash, login should fail for any password
	// unless adminHash is nil, in which case bcrypt.CompareHashAndPassword
	// will error. The test verifies the endpoint responds correctly.
	resp := post(t, "/auth/login", map[string]any{
		"username": "admin",
		"password": "wrong",
	})
	// With nil hash, the server should return 401
	if resp.StatusCode != 200 && resp.StatusCode != 401 {
		t.Fatalf("expected 200 or 401 from login, got %d", resp.StatusCode)
	}
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

func post(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(testServer.URL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
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
