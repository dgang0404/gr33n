// Phase 33 WS1 — read-tool hardening smokes (EnrichPromptBlock with seeded farm).
package main

import (
	"context"
	"strings"
	"testing"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

func TestPhase33WS1_EnrichSummarizeZoneHumidity(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}
	block := farmguardian.EnrichPromptBlock(ctx, q, 1, "what's the humidity in Flower Room?", snap, nil)
	if block == "" {
		t.Fatal("expected read-tool enrichment block for humidity question")
	}
	if !strings.Contains(block, "summarize_zone") {
		t.Fatalf("block missing summarize_zone:\n%s", block)
	}
	if !strings.Contains(block, "Flower Room") {
		t.Fatalf("block missing zone name:\n%s", block)
	}
}

func TestPhase33WS1_AckIntentSkipsSummarizeZone(t *testing.T) {
	if testPool == nil {
		t.Skip("testPool unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q := db.New(testPool)
	snap, err := farmguardian.BuildSnapshot(ctx, q, 1)
	if err != nil {
		t.Fatalf("BuildSnapshot: %v", err)
	}
	question := "Please acknowledge the humidity alert in Flower Room"
	block := farmguardian.EnrichPromptBlock(ctx, q, 1, question, snap, nil)
	if strings.Contains(block, "summarize_zone") {
		t.Fatalf("ack intent must not inject summarize_zone:\n%s", block)
	}
}
