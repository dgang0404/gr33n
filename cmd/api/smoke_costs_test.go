// Phase 20.95 WS5 — split out of cmd/api/smoke_test.go with zero behaviour
// change. Shared globals (testPool / testServer / testWorker / testNotifier)
// and helpers live in smoke_helpers_test.go; TestMain stays in smoke_test.go.

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestPhase2095CostEnergyColumns(t *testing.T) {
	tok := smokeJWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ── input_definitions.unit_cost / currency / unit_id round-trip ──────
	defName := uniqueName("ws2_def")
	resp := authPost(t, tok, "/farms/1/naturalfarming/inputs", map[string]any{
		"name":               defName,
		"category":           "animal_feed",
		"unit_cost":          12.3456,
		"unit_cost_currency": "eur",
		"unit_cost_unit_id":  1,
	})
	expectStatus(t, resp, 201)
	def := decodeMap(t, resp)
	defID := int64(def["id"].(float64))
	if got, _ := def["unit_cost_currency"].(string); got != "EUR" {
		t.Fatalf("expected unit_cost_currency=EUR, got %q", got)
	}

	// ── input_batches.low_stock_threshold round-trip ────────────────────
	resp = authPost(t, tok, "/farms/1/naturalfarming/batches", map[string]any{
		"input_definition_id": defID,
		"status":              "ready_for_use",
		"creation_start_date": "2026-02-01",
		"low_stock_threshold": 5.5,
	})
	expectStatus(t, resp, 201)
	batch := decodeMap(t, resp)
	if got, _ := batch["low_stock_threshold"].(float64); got != 5.5 {
		t.Fatalf("expected low_stock_threshold=5.5, got %v", batch["low_stock_threshold"])
	}

	// ── actuators.watts DEFAULT 0 ───────────────────────────────────────
	var watts *float64
	if err := testPool.QueryRow(ctx,
		`SELECT watts FROM gr33ncore.actuators WHERE farm_id = 1 LIMIT 1`,
	).Scan(&watts); err == nil { // only assert if an actuator exists
		if watts == nil || *watts != 0 {
			got := "nil"
			if watts != nil {
				got = fmt.Sprintf("%v", *watts)
			}
			t.Fatalf("expected actuators.watts default=0, got %s", got)
		}
	}

	// ── farm_energy_prices full CRUD ────────────────────────────────────
	resp = authPost(t, tok, "/farms/1/energy-prices", map[string]any{
		"effective_from": "2026-01-01",
		"price_per_kwh":  0.18,
		"currency":       "eur",
		"notes":          "baseline tariff",
	})
	expectStatus(t, resp, 201)
	price := decodeMap(t, resp)
	priceID := int64(price["id"].(float64))

	resp = authGet(t, tok, "/farms/1/energy-prices")
	expectStatus(t, resp, 200)
	rows := decodeSlice(t, resp)
	if len(rows) < 1 {
		t.Fatalf("expected at least 1 energy price row")
	}

	resp = authPut(t, tok, fmt.Sprintf("/energy-prices/%d", priceID), map[string]any{
		"effective_from": "2026-01-01",
		"effective_to":   "2026-06-30",
		"price_per_kwh":  0.21,
		"currency":       "EUR",
	})
	expectStatus(t, resp, 200)

	resp = authDelete(t, tok, fmt.Sprintf("/energy-prices/%d", priceID))
	expectStatus(t, resp, 204)

	// ── cost_transactions.crop_cycle_id round-trip ──────────────────────
	cycleName := uniqueName("ws2_cycle")
	resp = authPost(t, tok, "/farms/1/crop-cycles", map[string]any{
		"zone_id":       1,
		"name":          cycleName,
		"current_stage": "early_veg",
		"started_at":    "2026-01-15",
		"is_active":     false,
	})
	expectStatus(t, resp, 201)
	cycle := decodeMap(t, resp)
	cycleID := int64(cycle["id"].(float64))

	resp = authPost(t, tok, "/farms/1/costs", map[string]any{
		"transaction_date": "2026-02-01",
		"category":         "miscellaneous",
		"amount":           42.50,
		"currency":         "EUR",
		"is_income":        false,
		"crop_cycle_id":    cycleID,
	})
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		t.Fatalf("cost create with crop_cycle_id: status=%d", resp.StatusCode)
	}
	cost := decodeMap(t, resp)
	if got, _ := cost["crop_cycle_id"].(float64); int64(got) != cycleID {
		t.Fatalf("expected cost.crop_cycle_id=%d, got %v", cycleID, cost["crop_cycle_id"])
	}
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
