package farmguardian

import (
	"fmt"
	"strings"
	"testing"
	"time"
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

func TestSnapshot_RenderZonePhotos(t *testing.T) {
	s := Snapshot{
		ZoneCount: 1,
		ZoneNames: []string{"Flower Room"},
		ZonePhotoHints: []ZonePhotoHint{{
			ZoneName: "Flower Room", PhotoCount: 2, LatestAttachmentID: 99,
		}},
	}
	got := s.Render()
	for _, want := range []string{
		"Zone reference photos on file:",
		"Flower Room (2 photos)",
		"attachment #99",
		"walkthrough photos",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("render missing %q:\n%s", want, got)
		}
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

func TestSnapshot_RendersAlertDetailLines(t *testing.T) {
	defer withFrozenTimeNow(t, time.Date(2026, 5, 19, 18, 0, 0, 0, time.UTC))()

	s := Snapshot{
		UnreadAlerts: 2,
		UnreadAlertDetails: []UnreadAlertDetail{
			{
				ID:          11,
				Severity:    "high",
				Subject:     "Humidity threshold breach — Flower Room",
				Message:     "Humidity is 72.5% (threshold 65%) for sensor RH-Flower.",
				SourceType:  "sensor_reading",
				SourceID:    42,
				TriggeredAt: time.Date(2026, 5, 19, 14, 0, 0, 0, time.UTC), // 4h ago
			},
			{
				ID:          12,
				Severity:    "medium",
				Subject:     "Reservoir refill due",
				SourceType:  "automation_rule",
				SourceID:    7,
				TriggeredAt: time.Date(2026, 5, 19, 17, 50, 0, 0, time.UTC), // 10m ago
			},
		},
	}
	got := s.Render()
	for _, want := range []string{
		"Unread alerts: 2",
		"[high] Humidity threshold breach — Flower Room (sensor_reading #42, 4h ago)",
		"detail: Humidity is 72.5% (threshold 65%) for sensor RH-Flower.",
		"[medium] Reservoir refill due (automation_rule #7, 10m ago)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in:\n%s", want, got)
		}
	}
}

func TestSnapshot_TruncatesAlertDetails(t *testing.T) {
	defer withFrozenTimeNow(t, time.Date(2026, 5, 19, 18, 0, 0, 0, time.UTC))()

	// UnreadAlerts (total) is bigger than the detail slice — this is
	// the real-world case where the SQL LIMIT clipped the listing but
	// the total count keeps the full population. The renderer must
	// emit "+ N more" with N = total - rendered.
	const total = 25
	details := make([]UnreadAlertDetail, SnapshotMaxAlertDetails)
	for i := range details {
		details[i] = UnreadAlertDetail{
			ID:          int64(i + 1),
			Severity:    "high",
			Subject:     "Alert",
			TriggeredAt: time.Date(2026, 5, 19, 17, 30, 0, 0, time.UTC),
		}
	}
	s := Snapshot{UnreadAlerts: total, UnreadAlertDetails: details}
	got := s.Render()
	wantExtra := total - SnapshotMaxAlertDetails
	want := fmt.Sprintf("(+ %d more unread alerts)", wantExtra)
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q, got:\n%s", want, got)
	}
}

func TestSnapshot_AlertWithMessageOnlyTrimmedAndCapped(t *testing.T) {
	defer withFrozenTimeNow(t, time.Date(2026, 5, 19, 18, 0, 0, 0, time.UTC))()

	long := strings.Repeat("abcdefghij ", 30) // 330 chars, ~30 tokens
	s := Snapshot{
		UnreadAlerts: 1,
		UnreadAlertDetails: []UnreadAlertDetail{{
			ID:          1,
			Subject:     "Long alert",
			Message:     long,
			TriggeredAt: time.Date(2026, 5, 19, 17, 0, 0, 0, time.UTC),
		}},
	}
	got := s.Render()
	// Find the line beginning "    detail: " and assert it ends with the
	// ellipsis (cap was applied) and is no longer than AlertMessageSnippetMax + ellipsis.
	idx := strings.Index(got, "detail: ")
	if idx == -1 {
		t.Fatalf("expected detail line in:\n%s", got)
	}
	line := strings.TrimSpace(got[idx+len("detail: "):])
	if !strings.HasSuffix(line, "…") {
		t.Fatalf("expected truncation ellipsis on long message: %q", line)
	}
	// Snippet body without ellipsis must equal AlertMessageSnippetMax.
	runes := []rune(line)
	body := string(runes[:len(runes)-1])
	if len(body) != AlertMessageSnippetMax {
		t.Fatalf("expected body length %d, got %d", AlertMessageSnippetMax, len(body))
	}
}

func TestHumanizeAge(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "just now"},
		{30 * time.Second, "just now"},
		{2 * time.Minute, "2m ago"},
		{75 * time.Minute, "1h ago"},
		{5 * time.Hour, "5h ago"},
		{36 * time.Hour, "1d ago"},
		{72 * time.Hour, "3d ago"},
	}
	for _, c := range cases {
		if got := humanizeAge(c.d); got != c.want {
			t.Errorf("humanizeAge(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

// withFrozenTimeNow swaps out nowFunc (which timeSince reads through)
// so the snapshot renderer produces deterministic "Xh ago" strings.
// Returns a closure tests use as a defer to restore.
func withFrozenTimeNow(t *testing.T, now time.Time) func() {
	t.Helper()
	prev := nowFunc
	nowFunc = func() time.Time { return now }
	return func() { nowFunc = prev }
}
