package farmguardian

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	db "gr33n-api/internal/db"
)

// SnapshotMaxZones caps how many zone names get rendered into the prompt.
// Larger farms still work — extras are summarised with a trailing count.
const SnapshotMaxZones = 12

// SnapshotMaxCycles caps how many active crop cycles render in full.
const SnapshotMaxCycles = 8

// Snapshot is the live farm-state block injected into the Farm Guardian
// system prompt on grounded turns. It is intentionally tiny — operators
// want orientation cues ("3 zones, A B C; active cycle TomatoVeg in zone B;
// 2 unread alerts"), not a data dump.
type Snapshot struct {
	FarmID       int64
	ZoneCount    int
	ZoneNames    []string
	ActiveCycles []ActiveCycle
	UnreadAlerts int64
}

// ActiveCycle is a single in-flight grow cycle entry. Analytics is
// optional — populated by BuildSnapshot for the first
// SnapshotMaxAnalyticsCycles cycles (Phase 28 WS3) and left zero for the
// rest so the prompt stays bounded on farms running many cycles in
// parallel. ID is exposed so future UI deep-links (e.g. "open the
// CropCycleSummary view for cycle 42") have a stable target.
type ActiveCycle struct {
	ID        int64
	Name      string
	ZoneName  string
	Strain    string
	Stage     string
	Analytics CycleAnalytics
}

// IsEmpty returns true when there's nothing useful to render (avoids a noisy
// "Current farm snapshot:" header with zero bullets).
func (s Snapshot) IsEmpty() bool {
	return s.ZoneCount == 0 && len(s.ActiveCycles) == 0 && s.UnreadAlerts == 0
}

// Render returns the prompt-ready text block (no trailing newline). Empty
// snapshots return "" so the caller can omit the section entirely.
func (s Snapshot) Render() string {
	if s.IsEmpty() {
		return ""
	}
	var b strings.Builder
	if s.ZoneCount > 0 {
		b.WriteString(fmt.Sprintf("- Zones (%d):", s.ZoneCount))
		names := s.ZoneNames
		extra := 0
		if len(names) > SnapshotMaxZones {
			extra = len(names) - SnapshotMaxZones
			names = names[:SnapshotMaxZones]
		}
		if len(names) > 0 {
			b.WriteString(" ")
			b.WriteString(strings.Join(names, ", "))
		}
		if extra > 0 {
			b.WriteString(fmt.Sprintf(" (+ %d more)", extra))
		}
		b.WriteString("\n")
	}
	if len(s.ActiveCycles) > 0 {
		cycles := s.ActiveCycles
		extra := 0
		if len(cycles) > SnapshotMaxCycles {
			extra = len(cycles) - SnapshotMaxCycles
			cycles = cycles[:SnapshotMaxCycles]
		}
		b.WriteString(fmt.Sprintf("- Active cycles (%d):\n", len(s.ActiveCycles)))
		for _, c := range cycles {
			b.WriteString("  - ")
			b.WriteString(c.Name)
			if c.ZoneName != "" {
				b.WriteString(" — zone ")
				b.WriteString(c.ZoneName)
			}
			details := []string{}
			if c.Strain != "" {
				details = append(details, c.Strain)
			}
			if c.Stage != "" {
				details = append(details, "stage: "+c.Stage)
			}
			if len(details) > 0 {
				b.WriteString(" (")
				b.WriteString(strings.Join(details, "; "))
				b.WriteString(")")
			}
			b.WriteString("\n")
			if line := c.Analytics.renderLine(); line != "" {
				b.WriteString("    metrics: ")
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
		if extra > 0 {
			b.WriteString(fmt.Sprintf("  - (+ %d more active cycles)\n", extra))
		}
	}
	if s.UnreadAlerts > 0 {
		b.WriteString(fmt.Sprintf("- Unread alerts: %d\n", s.UnreadAlerts))
	}
	return strings.TrimRight(b.String(), "\n")
}

// PromptBlock wraps Render with a fixed header used by the chat handler.
// The header tells the model these facts are background context (no citation
// requirement) — important when the synthesis system prompt mandates [n] for
// every claim from the sources list.
func (s Snapshot) PromptBlock() string {
	body := s.Render()
	if body == "" {
		return ""
	}
	return "Current farm snapshot (background context — do not cite as [n]):\n" + body
}

// BuildSnapshot runs the three queries needed to populate Snapshot. Each
// query is non-fatal — if one fails we still return whatever else we got
// rather than blocking the chat turn.
func BuildSnapshot(ctx context.Context, q *db.Queries, farmID int64) (Snapshot, error) {
	s := Snapshot{FarmID: farmID}
	if q == nil {
		return s, nil
	}

	zones, zErr := q.ListZonesByFarm(ctx, farmID)
	if zErr == nil {
		s.ZoneCount = len(zones)
		s.ZoneNames = make([]string, 0, len(zones))
		for _, z := range zones {
			s.ZoneNames = append(s.ZoneNames, z.Name)
		}
		sort.Strings(s.ZoneNames)
	}

	cycles, cErr := q.ListCropCyclesByFarm(ctx, farmID)
	if cErr == nil {
		zoneByID := make(map[int64]string, len(zones))
		for _, z := range zones {
			zoneByID[z.ID] = z.Name
		}
		// Collect the active subset first so we can decide which ones
		// get analytics. We bound the analytics population at
		// SnapshotMaxAnalyticsCycles to keep prompt cost predictable.
		actives := make([]db.Gr33nfertigationCropCycle, 0, len(cycles))
		for _, c := range cycles {
			if c.IsActive {
				actives = append(actives, c)
			}
		}
		analyticsBudget := SnapshotMaxAnalyticsCycles
		for _, c := range actives {
			ac := ActiveCycle{ID: c.ID, Name: c.Name}
			if name, ok := zoneByID[c.ZoneID]; ok {
				ac.ZoneName = name
			}
			if c.StrainOrVariety != nil {
				ac.Strain = *c.StrainOrVariety
			}
			if c.CurrentStage.Valid {
				ac.Stage = string(c.CurrentStage.Gr33nfertigationGrowthStageEnum)
			}
			if analyticsBudget > 0 {
				if a, aerr := fetchCycleAnalytics(ctx, q, c); aerr == nil {
					ac.Analytics = a
				} else {
					slog.Warn("farm guardian cycle analytics failed",
						"farm_id", farmID, "cycle_id", c.ID, "err", aerr)
				}
				analyticsBudget--
			}
			s.ActiveCycles = append(s.ActiveCycles, ac)
		}
	}

	if cnt, aErr := q.CountUnreadAlertsByFarm(ctx, farmID); aErr == nil {
		s.UnreadAlerts = cnt
	}

	return s, nil
}
