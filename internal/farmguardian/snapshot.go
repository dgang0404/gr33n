package farmguardian

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	db "gr33n-api/internal/db"
)

// SnapshotMaxZones caps how many zone names get rendered into the prompt.
// Larger farms still work — extras are summarised with a trailing count.
const SnapshotMaxZones = 12

// SnapshotMaxCycles caps how many active crop cycles render in full.
const SnapshotMaxCycles = 8

// SnapshotMaxAlertDetails caps how many unread alerts get full detail
// (severity, subject, source, age) rendered into the prompt. The
// surviving alerts beyond the cap are still represented by the
// UnreadAlerts count so the LLM knows there are more. Phase 28 WS4.
const SnapshotMaxAlertDetails = 3

// AlertMessageSnippetMax caps how many characters of an alert's
// rendered message body get folded into the prompt. The message can be
// arbitrarily long (operator templates support markdown) so we trim to
// keep the snapshot's token budget predictable.
const AlertMessageSnippetMax = 160

// Snapshot is the live farm-state block injected into the Farm Guardian
// system prompt on grounded turns. It is intentionally tiny — operators
// want orientation cues ("3 zones, A B C; active cycle TomatoVeg in zone B;
// 2 unread alerts"), not a data dump.
type Snapshot struct {
	FarmID             int64
	ZoneCount          int
	ZoneNames          []string
	ActiveCycles       []ActiveCycle
	UnreadAlerts       int64
	UnreadAlertDetails []UnreadAlertDetail
}

// UnreadAlertDetail is the prompt-ready projection of a single unread
// alert. Populated by BuildSnapshot for the first
// SnapshotMaxAlertDetails alerts ordered by severity DESC, created_at
// DESC. Optional fields are kept as empty strings rather than pointers
// so the renderer stays branch-light. Phase 28 WS4.
type UnreadAlertDetail struct {
	ID          int64
	Severity    string // "low" / "medium" / "high" / "critical" / "" if NULL
	Subject     string // subject_rendered (trimmed)
	Message     string // message_text_rendered (trimmed, capped at AlertMessageSnippetMax)
	SourceType  string // e.g. "sensor_reading", "automation_rule", "automation_program"
	SourceID    int64  // 0 when source ID is NULL
	TriggeredAt time.Time
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
		details := s.UnreadAlertDetails
		// Defensive trim — BuildSnapshot already LIMITs to
		// SnapshotMaxAlertDetails at the SQL layer, but a caller could
		// construct a Snapshot directly (tests, future plumbing).
		if len(details) > SnapshotMaxAlertDetails {
			details = details[:SnapshotMaxAlertDetails]
		}
		// "+ N more unread alerts" reflects the gap between the total
		// unread count and how many details we rendered — not the
		// (possibly trimmed) detail slice. With UnreadAlerts=28000 and
		// details=3 the operator must see that the rest exist.
		extra := int(s.UnreadAlerts) - len(details)
		if extra < 0 {
			extra = 0
		}
		for _, a := range details {
			b.WriteString("  - ")
			if a.Severity != "" {
				b.WriteString("[")
				b.WriteString(a.Severity)
				b.WriteString("] ")
			}
			if a.Subject != "" {
				b.WriteString(a.Subject)
			} else {
				b.WriteString(fmt.Sprintf("alert #%d", a.ID))
			}
			meta := []string{}
			if a.SourceType != "" {
				if a.SourceID > 0 {
					meta = append(meta, fmt.Sprintf("%s #%d", a.SourceType, a.SourceID))
				} else {
					meta = append(meta, a.SourceType)
				}
			}
			if !a.TriggeredAt.IsZero() {
				meta = append(meta, humanizeAge(timeSince(a.TriggeredAt)))
			}
			if len(meta) > 0 {
				b.WriteString(" (")
				b.WriteString(strings.Join(meta, ", "))
				b.WriteString(")")
			}
			b.WriteString("\n")
			if a.Message != "" {
				// Defensive cap — toUnreadAlertDetail already trims,
				// but callers constructing the struct directly (tests,
				// future plumbing) shouldn't be able to blow the
				// prompt budget by stuffing a 5K-char Message in.
				msg := a.Message
				if r := []rune(msg); len(r) > AlertMessageSnippetMax {
					msg = string(r[:AlertMessageSnippetMax]) + "…"
				}
				b.WriteString("    detail: ")
				b.WriteString(msg)
				b.WriteString("\n")
			}
		}
		if extra > 0 {
			b.WriteString(fmt.Sprintf("  - (+ %d more unread alerts)\n", extra))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// timeSince is a small indirection so tests can freeze the clock the
// same way nowFunc lets them stub the analytics duration math.
var timeSince = func(t time.Time) time.Duration { return nowFunc().Sub(t) }

// humanizeAge renders a duration in a way operators read fluently in an
// alert summary: "12m ago", "4h ago", "3d ago". The LLM uses this to
// say "triggered 4h ago" without doing time math from a raw timestamp.
func humanizeAge(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d/time.Minute))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d/time.Hour))
	}
	return fmt.Sprintf("%dd ago", int(d/(24*time.Hour)))
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

	// Phase 28 WS4 — pull the top N unread alerts with enough detail
	// for Guardian to explain them. Best-effort: if the query fails,
	// we still keep the count from CountUnreadAlertsByFarm so the
	// snapshot is not silently impoverished.
	if s.UnreadAlerts > 0 {
		alerts, alErr := q.ListRecentUnreadAlertsByFarm(ctx, farmID, int32(SnapshotMaxAlertDetails))
		if alErr != nil {
			slog.Warn("farm guardian unread alert details failed",
				"farm_id", farmID, "err", alErr)
		} else {
			s.UnreadAlertDetails = make([]UnreadAlertDetail, 0, len(alerts))
			for _, a := range alerts {
				s.UnreadAlertDetails = append(s.UnreadAlertDetails, toUnreadAlertDetail(a))
			}
		}
	}

	return s, nil
}

// toUnreadAlertDetail projects a DB row into the prompt-ready struct.
// Trims whitespace, caps the message snippet at AlertMessageSnippetMax,
// and converts the enum + nullable IDs into safer plain types.
func toUnreadAlertDetail(a db.RecentUnreadAlertSummary) UnreadAlertDetail {
	out := UnreadAlertDetail{
		ID:          a.ID,
		TriggeredAt: a.CreatedAt,
	}
	if a.Severity.Valid {
		out.Severity = string(a.Severity.Gr33ncoreNotificationPriorityEnum)
	}
	if a.SubjectRendered != nil {
		out.Subject = strings.TrimSpace(*a.SubjectRendered)
	}
	if a.MessageTextRendered != nil {
		msg := strings.TrimSpace(*a.MessageTextRendered)
		// Collapse interior whitespace (newlines, tabs) so the
		// prompt block stays single-line per alert.
		msg = strings.Join(strings.Fields(msg), " ")
		if len(msg) > AlertMessageSnippetMax {
			msg = msg[:AlertMessageSnippetMax] + "…"
		}
		out.Message = msg
	}
	if a.TriggeringEventSourceType != nil {
		out.SourceType = strings.TrimSpace(*a.TriggeringEventSourceType)
	}
	if a.TriggeringEventSourceID != nil {
		out.SourceID = *a.TriggeringEventSourceID
	}
	return out
}
