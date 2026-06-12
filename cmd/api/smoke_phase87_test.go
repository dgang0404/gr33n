// Phase 87 — crop knowledge operator closure: catalog parity + Guardian multi-crop + DB registry.

package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"gr33n-api/internal/croplibrary"
	"gr33n-api/internal/farmguardian"
	db "gr33n-api/internal/db"
)

func TestPhase87_CatalogAndPickerParity(t *testing.T) {
	tok := smokeJWT(t)

	resp := authGet(t, tok, "/commons/crop-catalog")
	expectStatus(t, resp, http.StatusOK)
	catalog := decodeMap(t, resp)
	entries, _ := catalog["entries"].([]any)
	if len(entries) < 50 {
		t.Fatalf("want >= 50 commons catalog entries, got %d", len(entries))
	}

	resp = authGet(t, tok, "/farms/1/crop-library/picker")
	expectStatus(t, resp, http.StatusOK)
	picker := decodeMap(t, resp)
	counts, _ := picker["counts"].(map[string]any)
	withTargets, _ := counts["with_targets"].(float64)
	if withTargets < 46 {
		t.Fatalf("want >= 46 picker crops with targets, got %v", withTargets)
	}
}

func TestPhase87_GuardianCompareCropsFromDB(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	q := db.New(testPool)
	block, err := farmguardian.LookupCropTargets(ctx, q, 1, "Compare cannabis vs tomato EC targets for early veg", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(block, "multi-crop") {
		t.Fatalf("expected multi-crop block, got: %s", block)
	}
	if !strings.Contains(block, "cannabis") || !strings.Contains(block, "tomato") {
		t.Fatalf("expected both crops in block: %s", block)
	}
	if !strings.Contains(block, "mS/cm") {
		t.Fatalf("expected mS/cm EC in compare block: %s", block)
	}
}

func TestPhase87_DBCatalogRegistryAlias(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	croplibrary.SetRuntimeCatalogQuerier(db.New(testPool))
	reg, err := croplibrary.DefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	r := croplibrary.NewRegistry(reg)
	m, ok := r.ResolveTerm("aubergine")
	if !ok || m.Key != "eggplant" {
		t.Fatalf("expected aubergine → eggplant from DB registry, got %+v ok=%v", m, ok)
	}
}
