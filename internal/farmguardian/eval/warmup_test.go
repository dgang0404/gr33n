package eval

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWarmupFarmCounsel_sendsChatModel(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/guardian/warmup":
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &gotBody)
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"state":"stirring","chat_model":"phi3:mini"}`))
		case "/v1/chat/health":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"awakening":{"state":"ready"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := &APIClient{
		BaseURL: srv.URL,
		FarmID:  1,
		HTTP:    srv.Client(),
	}
	if err := client.WarmupFarmCounsel(context.Background(), "phi3:mini", 5*time.Second); err != nil {
		t.Fatal(err)
	}
	if gotBody["chat_model"] != "phi3:mini" {
		t.Fatalf("chat_model=%v body=%v", gotBody["chat_model"], gotBody)
	}
	if gotBody["mode"] != "farm_counsel" {
		t.Fatalf("mode=%v", gotBody["mode"])
	}
}

func TestWarmupFarmCounsel_omitsChatModelWhenEmpty(t *testing.T) {
	t.Parallel()
	var rawBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guardian/warmup" {
			http.NotFound(w, r)
			return
		}
		b, _ := io.ReadAll(r.Body)
		rawBody = string(b)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"state":"stirring"}`))
	}))
	defer srv.Close()

	client := &APIClient{BaseURL: srv.URL, FarmID: 1, HTTP: srv.Client()}
	_ = client.WarmupFarmCounsel(context.Background(), "  ", time.Second)
	if strings.Contains(rawBody, "chat_model") {
		t.Fatalf("unexpected chat_model in body: %s", rawBody)
	}
}
