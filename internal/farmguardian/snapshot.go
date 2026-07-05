package farmguardian

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/zonephotos"
)

// SnapshotMaxZones caps how many zone names get rendered into the prompt.
// Larger farms still work — extras are summarised with a trailing count.
const SnapshotMaxZones = 12

// SnapshotMaxCycles caps how many active crop cycles render in full.
const SnapshotMaxCycles = 8

// SnapshotMaxPlantNames caps plant display names in the live snapshot (Phase 32 WS1).
const SnapshotMaxPlantNames = 8

// SnapshotMaxProgramsPerZone caps active program names listed per zone in the snapshot.
const SnapshotMaxProgramsPerZone = 4

// SnapshotMaxProgramZones caps how many zones get a programs-by-zone line.
const SnapshotMaxProgramZones = 8

// SnapshotMaxOfflineDeviceNames caps offline device names in the snapshot.
const SnapshotMaxOfflineDeviceNames = 2

// SnapshotMaxUnscheduledProgramNames caps manual-only program names listed.
const SnapshotMaxUnscheduledProgramNames = 3
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
	PlantCount         int
	PlantNames         []string
	ProgramsByZone     []ZoneProgramsSummary
	ActiveCycles       []ActiveCycle
	UnreadAlerts       int64
	UnreadAlertDetails []UnreadAlertDetail
	ZonePhotoHints     []ZonePhotoHint
	Devices            DeviceSummary
	FertigationSchedule FertigationScheduleSummary
}

// ZoneProgramsSummary lists active fertigation program names targeting a zone.
type ZoneProgramsSummary struct {
	ZoneID   int64
	ZoneName string
	Programs []string
}

// DeviceSummary is a compact edge-device posture line for grounded chat (Phase 127).
type DeviceSummary struct {
	Total        int
	Online       int
	Offline      int
	OfflineNames []string
}

// FertigationScheduleSummary counts how active programs are scheduled vs manual-only.
type FertigationScheduleSummary struct {
	ScheduledActive   int
	UnscheduledActive int
	UnscheduledNames  []string
}

// ZonePhotoHint tells the model which zones have operator reference photos on file.
type ZonePhotoHint struct {
	ZoneID             int64
	ZoneName           string
	PhotoCount         int
	LatestAttachmentID int64
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
	BatchLabel string
	Stage     string
	Analytics CycleAnalytics
}

// IsEmpty returns true when there's nothing useful to render (avoids a noisy
// "Current farm snapshot:" header with zero bullets).
func (s Snapshot) IsEmpty() bool {
	return s.ZoneCount == 0 && len(s.ActiveCycles) == 0 && s.UnreadAlerts == 0 &&
		s.PlantCount == 0 && len(s.ProgramsByZone) == 0 && s.Devices.Total == 0
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
	if s.PlantCount > 0 {
		names := s.PlantNames
		extra := 0
		if len(names) > SnapshotMaxPlantNames {
			extra = s.PlantCount - SnapshotMaxPlantNames
			names = names[:SnapshotMaxPlantNames]
		}
		b.WriteString(fmt.Sprintf("- Plants (%d):", s.PlantCount))
		if len(names) > 0 {
			b.WriteString(" ")
			b.WriteString(strings.Join(names, ", "))
		}
		if extra > 0 {
			b.WriteString(fmt.Sprintf(" (+ %d more)", extra))
		}
		b.WriteString("\n")
	}
	if len(s.ProgramsByZone) > 0 {
		byZone := s.ProgramsByZone
		extraZones := 0
		if len(byZone) > SnapshotMaxProgramZones {
			extraZones = len(byZone) - SnapshotMaxProgramZones
			byZone = byZone[:SnapshotMaxProgramZones]
		}
		b.WriteString("- Active fertigation programs by zone:\n")
		for _, zp := range byZone {
			b.WriteString("  - ")
			b.WriteString(zp.ZoneName)
			progs := zp.Programs
			extraProgs := 0
			if len(progs) > SnapshotMaxProgramsPerZone {
				extraProgs = len(progs) - SnapshotMaxProgramsPerZone
				progs = progs[:SnapshotMaxProgramsPerZone]
			}
			if len(progs) > 0 {
				b.WriteString(": ")
				b.WriteString(strings.Join(progs, ", "))
			} else {
				b.WriteString(": (none active)")
			}
			if extraProgs > 0 {
				b.WriteString(fmt.Sprintf(" (+ %d more programs)", extraProgs))
			}
			b.WriteString("\n")
		}
		if extraZones > 0 {
			b.WriteString(fmt.Sprintf("  - (+ %d more zones with programs)\n", extraZones))
		}
	}
	if s.FertigationSchedule.ScheduledActive > 0 || s.FertigationSchedule.UnscheduledActive > 0 {
		b.WriteString(fmt.Sprintf(
			"- Fertigation programs: %d on schedule, %d manual-only (no cron)\n",
			s.FertigationSchedule.ScheduledActive,
			s.FertigationSchedule.UnscheduledActive,
		))
		names := s.FertigationSchedule.UnscheduledNames
		if len(names) > 0 {
			b.WriteString("  - Manual-only: ")
			b.WriteString(strings.Join(names, ", "))
			b.WriteString("\n")
		}
	}
	if s.Devices.Total > 0 {
		b.WriteString(fmt.Sprintf("- Edge devices: %d online, %d offline", s.Devices.Online, s.Devices.Offline))
		if len(s.Devices.OfflineNames) > 0 {
			b.WriteString(" (")
			b.WriteString(strings.Join(s.Devices.OfflineNames, ", "))
			b.WriteString(")")
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
			if c.BatchLabel != "" {
				details = append(details, c.BatchLabel)
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
	if len(s.ZonePhotoHints) > 0 {
		b.WriteString("- Zone reference photos on file:\n")
		for _, h := range s.ZonePhotoHints {
			b.WriteString("  - ")
			b.WriteString(h.ZoneName)
			b.WriteString(fmt.Sprintf(" (%d photo", h.PhotoCount))
			if h.PhotoCount != 1 {
				b.WriteString("s")
			}
			b.WriteString(")")
			if h.LatestAttachmentID > 0 {
				b.WriteString(fmt.Sprintf("; latest attachment #%d (/file-attachments/%d/content)", h.LatestAttachmentID, h.LatestAttachmentID))
			}
			b.WriteString("\n")
		}
		b.WriteString("  - Operators can ask Guardian about a zone's walkthrough photos; image analysis needs a vision model (optional WS6).\n")
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
			if meta, _, mErr := zonephotos.ParseMeta(z.MetaData); mErr == nil && len(meta.PhotoAttachmentIDs) > 0 {
				s.ZonePhotoHints = append(s.ZonePhotoHints, ZonePhotoHint{
					ZoneID:             z.ID,
					ZoneName:           z.Name,
					PhotoCount:         len(meta.PhotoAttachmentIDs),
					LatestAttachmentID: zonephotos.LatestID(meta),
				})
			}
		}
		sort.Strings(s.ZoneNames)
		sort.Slice(s.ZonePhotoHints, func(i, j int) bool {
			return s.ZonePhotoHints[i].ZoneName < s.ZonePhotoHints[j].ZoneName
		})
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
			if c.BatchLabel != nil {
				ac.BatchLabel = *c.BatchLabel
			}
			if c.CurrentStage != nil {
				ac.Stage = string(*c.CurrentStage)
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

	if plants, pErr := q.ListPlantsByFarm(ctx, farmID); pErr == nil {
		s.PlantCount = len(plants)
		s.PlantNames = make([]string, 0, len(plants))
		for _, p := range plants {
			name := strings.TrimSpace(p.DisplayName)
			if p.VarietyOrCultivar != nil && strings.TrimSpace(*p.VarietyOrCultivar) != "" {
				name += " (" + strings.TrimSpace(*p.VarietyOrCultivar) + ")"
			}
			if name != "" {
				s.PlantNames = append(s.PlantNames, name)
			}
		}
		sort.Strings(s.PlantNames)
	} else {
		slog.Warn("farm guardian plants summary failed", "farm_id", farmID, "err", pErr)
	}

	if programs, prErr := q.ListProgramsByFarm(ctx, farmID); prErr == nil {
		zoneByID := make(map[int64]string, len(zones))
		for _, z := range zones {
			zoneByID[z.ID] = z.Name
		}
		type zoneBucket struct {
			zoneID   int64
			zoneName string
			programs []string
		}
		buckets := make(map[int64]*zoneBucket)
		var unscheduled []string
		for _, p := range programs {
			if !p.IsActive {
				continue
			}
			if p.ScheduleID != nil {
				s.FertigationSchedule.ScheduledActive++
			} else {
				s.FertigationSchedule.UnscheduledActive++
				if len(unscheduled) < SnapshotMaxUnscheduledProgramNames {
					unscheduled = append(unscheduled, strings.TrimSpace(p.Name))
				}
			}
			if p.TargetZoneID == nil {
				continue
			}
			zid := *p.TargetZoneID
			b, ok := buckets[zid]
			if !ok {
				name := zoneByID[zid]
				if name == "" {
					name = fmt.Sprintf("zone #%d", zid)
				}
				b = &zoneBucket{zoneID: zid, zoneName: name}
				buckets[zid] = b
			}
			b.programs = append(b.programs, strings.TrimSpace(p.Name))
		}
		s.FertigationSchedule.UnscheduledNames = unscheduled
		for _, b := range buckets {
			sort.Strings(b.programs)
			s.ProgramsByZone = append(s.ProgramsByZone, ZoneProgramsSummary{
				ZoneID:   b.zoneID,
				ZoneName: b.zoneName,
				Programs: b.programs,
			})
		}
		sort.Slice(s.ProgramsByZone, func(i, j int) bool {
			return s.ProgramsByZone[i].ZoneName < s.ProgramsByZone[j].ZoneName
		})
	} else {
		slog.Warn("farm guardian programs-by-zone failed", "farm_id", farmID, "err", prErr)
	}

	if counts, dErr := q.CountDevicesByStatusForFarm(ctx, farmID); dErr == nil {
		for _, row := range counts {
			s.Devices.Total += int(row.Cnt)
			switch string(row.Status) {
			case "online":
				s.Devices.Online = int(row.Cnt)
			case "offline":
				s.Devices.Offline = int(row.Cnt)
			}
		}
	} else {
		slog.Warn("farm guardian device counts failed", "farm_id", farmID, "err", dErr)
	}
	if s.Devices.Offline > 0 {
		if devices, err := q.ListDevicesByFarm(ctx, farmID); err == nil {
			for _, d := range devices {
				if string(d.Status) != "online" {
					continue
				}
				if len(s.Devices.OfflineNames) >= SnapshotMaxOfflineDeviceNames {
					break
				}
				s.Devices.OfflineNames = append(s.Devices.OfflineNames, strings.TrimSpace(d.Name))
			}
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
		alerts, alErr := q.ListRecentUnreadAlertsByFarm(ctx, db.ListRecentUnreadAlertsByFarmParams{FarmID: farmID, Limit: int32(SnapshotMaxAlertDetails)})
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
func toUnreadAlertDetail(a db.ListRecentUnreadAlertsByFarmRow) UnreadAlertDetail {
	out := UnreadAlertDetail{
		ID:          a.ID,
		TriggeredAt: a.CreatedAt,
	}
	if a.Severity != nil {
		out.Severity = string(*a.Severity)
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
