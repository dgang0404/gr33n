// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	automationworker "gr33n-api/internal/automation"
	db "gr33n-api/internal/db"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Helpers ─────────────────────────────────────────────────────────────────
