package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]int{"id": 42})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q", ct)
	}
	var body map[string]int
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["id"] != 42 {
		t.Fatalf("body = %#v", body)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusForbidden, "not allowed")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "not allowed" {
		t.Fatalf("error = %q", body["error"])
	}
}

func TestPathID(t *testing.T) {
	tests := []struct {
		path string
		n    int
		want int64
		err  bool
	}{
		{"/farms/42/zones", 2, 42, false},
		{"/farms/42/zones", 3, 0, true},
		{"/", 1, 0, true},
	}
	for _, tc := range tests {
		got, err := PathID(tc.path, tc.n)
		if tc.err {
			if err == nil {
				t.Fatalf("PathID(%q, %d) expected error", tc.path, tc.n)
			}
			continue
		}
		if err != nil {
			t.Fatalf("PathID(%q, %d): %v", tc.path, tc.n, err)
		}
		if got != tc.want {
			t.Fatalf("PathID(%q, %d) = %d, want %d", tc.path, tc.n, got, tc.want)
		}
	}
}

func TestPathID_InvalidSegment(t *testing.T) {
	_, err := PathID("/farms/abc/zones", 2)
	if err == nil {
		t.Fatal("expected parse error")
	}
}
