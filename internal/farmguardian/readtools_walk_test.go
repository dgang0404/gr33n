package farmguardian

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/ai"
)

func TestShouldRunWalkFarmReadIntent(t *testing.T) {
	if !shouldRunWalkFarmReadIntent("run my morning check", nil) {
		t.Fatal("expected morning check intent")
	}
	if !shouldRunWalkFarmReadIntent("", &ContextRef{GuardianMode: "morning_walkthrough"}) {
		t.Fatal("expected guardian_mode to trigger")
	}
	if !shouldRunWalkFarmReadIntent("what needs attention today", nil) {
		t.Fatal("expected attention intent")
	}
	if shouldRunWalkFarmReadIntent("hello", nil) {
		t.Fatal("expected no match for greeting")
	}
}

func TestReadToolIDs_IncludesWalkFarm(t *testing.T) {
	for _, id := range ReadToolIDs() {
		if id == "walk_farm" {
			return
		}
	}
	t.Fatal("walk_farm missing from ReadToolIDs")
}

func TestPlatformContextBlock_IncludesWalkFarmRule(t *testing.T) {
	block := PlatformContextBlock(ai.Config{Enabled: true}, true, ReadToolIDs())
	if !strings.Contains(block, "walk_farm") || !strings.Contains(block, "Morning walkthrough") {
		t.Fatalf("platform context missing walk farm rule")
	}
}

func TestComfortValueBreach(t *testing.T) {
	sp := db.Gr33ncoreZoneSetpoint{
		MinValue: mustNumeric(20),
		MaxValue: mustNumeric(26),
	}
	if comfortValueBreach(sp, 29) == "" {
		t.Fatal("expected above-max breach")
	}
	if comfortValueBreach(sp, 22) != "" {
		t.Fatal("expected in-range")
	}
}

func mustNumeric(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.1f", v))
	return n
}
