// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"net/http"
	"testing"
)

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
