package farmguardian

import (
	"strings"
	"testing"
)

func TestSnapshot_IsEmpty(t *testing.T) {
	if !(Snapshot{}).IsEmpty() {
		t.Fatal("zero Snapshot must be empty")
	}
	if (Snapshot{ZoneCount: 1}).IsEmpty() {
		t.Fatal("snapshot with a zone is not empty")
	}
	if (Snapshot{UnreadAlerts: 3}).IsEmpty() {
		t.Fatal("snapshot with alerts is not empty")
	}
	if (Snapshot{ActiveCycles: []ActiveCycle{{Name: "x"}}}).IsEmpty() {
		t.Fatal("snapshot with cycle is not empty")
	}
}

func TestSnapshot_RenderEmptyReturnsEmpty(t *testing.T) {
	if got := (Snapshot{}).Render(); got != "" {
		t.Fatalf("expected empty render, got %q", got)
	}
	if got := (Snapshot{}).PromptBlock(); got != "" {
		t.Fatalf("expected empty PromptBlock, got %q", got)
	}
}

func TestSnapshot_RenderZonesAndAlerts(t *testing.T) {
	s := Snapshot{
		ZoneCount:    3,
		ZoneNames:    []string{"A", "B", "C"},
		UnreadAlerts: 2,
	}
	got := s.Render()
	for _, want := range []string{"Zones (3):", "A, B, C", "Unread alerts: 2"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered output missing %q:\n%s", want, got)
		}
	}
	// PromptBlock prepends the header so the model knows not to cite these.
	pb := s.PromptBlock()
	if !strings.HasPrefix(pb, "Current farm snapshot (background context") {
		t.Fatalf("PromptBlock missing header: %s", pb)
	}
	if !strings.Contains(pb, "do not cite as [n]") {
		t.Fatalf("PromptBlock missing citation guidance: %s", pb)
	}
}

func TestSnapshot_TruncatesZones(t *testing.T) {
	names := []string{}
	for i := 0; i < SnapshotMaxZones+5; i++ {
		names = append(names, "Z")
	}
	s := Snapshot{ZoneCount: len(names), ZoneNames: names}
	got := s.Render()
	if !strings.Contains(got, "(+ 5 more)") {
		t.Fatalf("expected truncation note, got %s", got)
	}
}

func TestSnapshot_RendersActiveCycleDetails(t *testing.T) {
	s := Snapshot{
		ActiveCycles: []ActiveCycle{
			{Name: "TomatoSpring", ZoneName: "B", Strain: "Roma", Stage: "vegetative"},
			{Name: "BasilWinter", ZoneName: "A"},
		},
	}
	got := s.Render()
	for _, want := range []string{
		"Active cycles (2):",
		"TomatoSpring — zone B (Roma; stage: vegetative)",
		"BasilWinter — zone A",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in:\n%s", want, got)
		}
	}
}

func TestSnapshot_TruncatesCycles(t *testing.T) {
	cycles := make([]ActiveCycle, SnapshotMaxCycles+3)
	for i := range cycles {
		cycles[i] = ActiveCycle{Name: "C"}
	}
	s := Snapshot{ActiveCycles: cycles}
	got := s.Render()
	if !strings.Contains(got, "(+ 3 more active cycles)") {
		t.Fatalf("expected cycle truncation note, got:\n%s", got)
	}
}
