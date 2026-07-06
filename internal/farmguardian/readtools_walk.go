// Phase 60 WS1 — morning walkthrough read tool (walk_farm).

package farmguardian

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
)

const (
	walkFarmMaxAlertFindings   = 5
	walkFarmMaxLowStockLines   = 4
	walkFarmMaxComfortFindings = 5
	walkFarmMaxFeedHighlights  = 3
)

// WalkFarmPersonaRule guides Guardian during morning walkthrough turns.
const WalkFarmPersonaRule = `Morning walkthrough (Phase 60): Use walk_farm for structured daily checks. Report only categories with findings — skip empty sections. Use zone names, not IDs. Plain farmer language; no schema terms. Read-only — never propose changes during the walkthrough unless the operator asks.`

var walkFarmIntent = regexp.MustCompile(`(?i)\b(morning (check|walkthrough|walk)|walk_farm|walk the farm|what needs attention|urgent issue|what should i (check|do) (this )?morning|daily check|farm walk|anything urgent)\b`)

type walkFinding struct {
	Category    string
	Severity    string // warn | ok
	PlainText   string
	ActionRoute string
	AlertID     int64 // alerts category — Phase 132 walkthrough proposals
}

func shouldRunWalkFarmReadIntent(question string, ref *ContextRef) bool {
	if ref != nil && strings.EqualFold(strings.TrimSpace(ref.GuardianMode), "morning_walkthrough") {
		return true
	}
	q := strings.TrimSpace(question)
	if q == "" {
		return false
	}
	return walkFarmIntent.MatchString(q)
}

func renderWalkFarm(ctx context.Context, q db.Querier, farmID int64) (string, error) {
	farmLabel := fmt.Sprintf("farm #%d", farmID)
	if farm, err := q.GetFarmByID(ctx, farmID); err == nil {
		if name := strings.TrimSpace(farm.Name); name != "" {
			farmLabel = name
		}
	}

	findings, err := collectWalkFarmFindings(ctx, q, farmID)
	if err != nil {
		return "", err
	}

	zones, _ := q.ListZonesByFarm(ctx, farmID)
	zoneName := map[int64]string{}
	for _, z := range zones {
		zoneName[z.ID] = strings.TrimSpace(z.Name)
	}
	feedOK, _, err := walkFarmFeedFindings(ctx, q, farmID, zoneName)
	if err != nil {
		return "", err
	}
	okHighlights := append([]string(nil), feedOK...)

	var b strings.Builder
	b.WriteString("walk_farm — " + farmLabel)
	b.WriteString("\nAreas checked: alerts, feeds, devices, comfort, supplies")

	warns := filterWalkFindings(findings, "warn")
	if len(warns) == 0 {
		b.WriteString("\nFarm looks good this morning.")
		if len(okHighlights) > 0 {
			b.WriteString(" " + strings.Join(okHighlights, " "))
		} else {
			b.WriteString(" No urgent issues flagged.")
		}
		return b.String(), nil
	}

	b.WriteString(fmt.Sprintf("\nFindings (%d need attention):", len(warns)))
	for i, f := range warns {
		b.WriteString(fmt.Sprintf("\n%d. [%s/%s] %s", i+1, f.Severity, f.Category, f.PlainText))
		if f.ActionRoute != "" {
			b.WriteString(" → " + f.ActionRoute)
		}
	}
	if len(okHighlights) > 0 {
		b.WriteString("\nAlso OK: " + strings.Join(okHighlights, "; "))
	}
	if len(warns) == 1 {
		b.WriteString("\nSummary: One thing needs attention — start there.")
	} else {
		b.WriteString(fmt.Sprintf("\nSummary: %d things need attention. Start with the first finding.", len(warns)))
	}
	return b.String(), nil
}

func filterWalkFindings(in []walkFinding, severity string) []walkFinding {
	out := make([]walkFinding, 0, len(in))
	for _, f := range in {
		if f.Severity == severity {
			out = append(out, f)
		}
	}
	return out
}

func walkFarmAlertFindings(ctx context.Context, q db.Querier, farmID int64) ([]walkFinding, error) {
	alerts, err := q.ListAlertsByFarm(ctx, db.ListAlertsByFarmParams{
		FarmID: farmID,
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}
	var out []walkFinding
	for _, a := range alerts {
		if a.IsAcknowledged != nil && *a.IsAcknowledged {
			continue
		}
		subject := strings.TrimSpace(ptrStr(a.SubjectRendered))
		if subject == "" {
			subject = strings.TrimSpace(ptrStr(a.MessageTextRendered))
		}
		if subject == "" {
			subject = "Alert"
		}
		age := ""
		if !a.CreatedAt.IsZero() {
			age = humanizeAge(timeSince(a.CreatedAt)) + " ago"
		}
		line := subject
		if age != "" {
			line += " since " + age
		}
		if a.Severity != nil && strings.TrimSpace(string(*a.Severity)) != "" {
			line = "[" + strings.TrimSpace(string(*a.Severity)) + "] " + line
		}
		out = append(out, walkFinding{
			Category:    "alerts",
			Severity:    "warn",
			PlainText:   line,
			ActionRoute: "/alerts",
			AlertID:     a.ID,
		})
		if len(out) >= walkFarmMaxAlertFindings {
			break
		}
	}
	return out, nil
}

func walkFarmFeedFindings(ctx context.Context, q db.Querier, farmID int64, zoneName map[int64]string) (ok []string, warn []walkFinding, err error) {
	schedules, err := q.ListSchedulesByFarm(ctx, farmID)
	if err != nil {
		return nil, nil, err
	}
	now := nowFunc()
	today := now.Format("2006-01-02")
	for _, s := range schedules {
		if !s.IsActive {
			continue
		}
		label := strings.TrimSpace(s.Name)
		if label == "" {
			label = "Feed schedule"
		}
		when := walkFarmScheduleWhen(s, now)
		if when == "" {
			continue
		}
		line := fmt.Sprintf("%s — %s", label, when)
		if s.NextExpectedTriggerTime.Valid {
			nextDay := s.NextExpectedTriggerTime.Time.Format("2006-01-02")
			if nextDay == today || s.NextExpectedTriggerTime.Time.Before(now.Add(24*time.Hour)) {
				ok = append(ok, line)
			}
		} else if strings.Contains(strings.ToLower(when), "today") || strings.Contains(strings.ToLower(when), "every day") {
			ok = append(ok, line)
		}
		if len(ok) >= walkFarmMaxFeedHighlights {
			break
		}
	}
	if len(ok) == 0 {
		programs, perr := q.ListProgramsByFarm(ctx, farmID)
		if perr == nil {
			active := 0
			for _, p := range programs {
				if p.IsActive {
					active++
				}
			}
			if active > 0 {
				ok = append(ok, fmt.Sprintf("%d active feeding plan(s) on file", active))
			}
		}
	}
	return ok, warn, nil
}

func walkFarmScheduleWhen(s db.Gr33ncoreSchedule, now time.Time) string {
	if s.NextExpectedTriggerTime.Valid {
		delta := s.NextExpectedTriggerTime.Time.Sub(now)
		if delta >= 0 && delta < 24*time.Hour {
			return "next run " + humanizeAge(delta) + " from now"
		}
	}
	cron := strings.TrimSpace(s.CronExpression)
	if cron == "" {
		return ""
	}
	parts := strings.Fields(cron)
	if len(parts) < 5 {
		return "scheduled"
	}
	hourField := parts[1]
	minField := parts[0]
	if hourField != "*" && (parts[2] == "*" || parts[2] == "?") {
		if h, err := parseHourField(hourField); err == nil {
			min := 0
			if minField != "*" && minField != "?" {
				fmt.Sscanf(minField, "%d", &min)
			}
			return fmt.Sprintf("runs today around %d:%02d", h, min)
		}
	}
	return "active schedule"
}

func parseHourField(field string) (int, error) {
	field = strings.TrimSpace(field)
	var h int
	_, err := fmt.Sscanf(field, "%d", &h)
	return h, err
}

func walkFarmDeviceFindings(ctx context.Context, q db.Querier, farmID int64) ([]walkFinding, error) {
	devices, err := q.ListDevicesByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	var out []walkFinding
	for _, d := range devices {
		if d.DeletedAt.Valid {
			continue
		}
		offline := false
		if string(d.Status) != "online" {
			offline = true
		}
		if d.LastHeartbeat.Valid {
			if timeSince(d.LastHeartbeat.Time) > deviceHealthOfflineAfter {
				offline = true
			}
		} else {
			offline = true
		}
		if !offline {
			continue
		}
		age := "never"
		if d.LastHeartbeat.Valid {
			age = humanizeAge(timeSince(d.LastHeartbeat.Time)) + " ago"
		}
		out = append(out, walkFinding{
			Category:    "devices",
			Severity:    "warn",
			PlainText:   fmt.Sprintf("%s last seen %s — may need reconnect", strings.TrimSpace(d.Name), age),
			ActionRoute: "/",
		})
	}
	return out, nil
}

func walkFarmComfortFindings(ctx context.Context, q db.Querier, farmID int64, zoneName map[int64]string) ([]walkFinding, error) {
	setpoints, err := q.ListSetpointsByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	byZone := map[int64][]db.Gr33ncoreZoneSetpoint{}
	for _, sp := range setpoints {
		if sp.ZoneID == nil || sp.CropCycleID != nil {
			continue
		}
		byZone[*sp.ZoneID] = append(byZone[*sp.ZoneID], sp)
	}

	var out []walkFinding
	for zoneID, sps := range byZone {
		for _, sp := range sps {
			if !walkFarmClimateSensorType(sp.SensorType) {
				continue
			}
			if !sp.MinValue.Valid && !sp.MaxValue.Valid && !sp.IdealValue.Valid {
				continue
			}
			reading, val, have := latestZoneSensorReading(ctx, q, zoneID, sp.SensorType)
			if !have || val == nil {
				continue
			}
			if breach := comfortValueBreach(sp, *val); breach != "" {
				zn := zoneName[zoneID]
				if zn == "" {
					zn = fmt.Sprintf("zone #%d", zoneID)
				}
				out = append(out, walkFinding{
					Category:    "comfort",
					Severity:    "warn",
					PlainText:   fmt.Sprintf("%s in %s is %.1f — %s", comfortSensorLabel(sp.SensorType), zn, *val, breach),
					ActionRoute: fmt.Sprintf("/comfort-targets?zone_id=%d", zoneID),
				})
			}
			_ = reading
			if len(out) >= walkFarmMaxComfortFindings {
				return out, nil
			}
		}
	}
	return out, nil
}

func walkFarmClimateSensorType(sensorType string) bool {
	switch strings.ToLower(strings.TrimSpace(sensorType)) {
	case "temperature", "humidity", "co2", "dew_point", "vpd":
		return true
	default:
		return false
	}
}

func comfortSensorLabel(sensorType string) string {
	switch strings.ToLower(strings.TrimSpace(sensorType)) {
	case "temperature":
		return "Temp"
	case "humidity":
		return "Humidity"
	default:
		return sensorType
	}
}

func comfortValueBreach(sp db.Gr33ncoreZoneSetpoint, val float64) string {
	minV := numericOptional(sp.MinValue)
	maxV := numericOptional(sp.MaxValue)
	if minV != nil && val < *minV {
		if maxV != nil {
			return fmt.Sprintf("below your %.0f–%.0f band", *minV, *maxV)
		}
		return fmt.Sprintf("below your %.0f minimum", *minV)
	}
	if maxV != nil && val > *maxV {
		if minV != nil {
			return fmt.Sprintf("above your %.0f–%.0f band", *minV, *maxV)
		}
		return fmt.Sprintf("above your %.0f maximum", *maxV)
	}
	return ""
}

func numericOptional(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	v := numericToFloat64(n)
	return &v
}

func latestZoneSensorReading(ctx context.Context, q db.Querier, zoneID int64, sensorType string) (db.Gr33ncoreSensorReading, *float64, bool) {
	zid := zoneID
	reading, err := q.GetLatestReadingForZoneSensorType(ctx, db.GetLatestReadingForZoneSensorTypeParams{
		ZoneID:     &zid,
		SensorType: sensorType,
	})
	if err != nil {
		return db.Gr33ncoreSensorReading{}, nil, false
	}
	val := numericToFloat64(reading.ValueRaw)
	if !reading.ValueRaw.Valid {
		return reading, nil, false
	}
	return reading, &val, true
}

func walkFarmLowStockFindings(ctx context.Context, q db.Querier, farmID int64) ([]walkFinding, error) {
	rows, err := q.ListLowStockBatchesByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	var out []walkFinding
	for i, row := range rows {
		if i >= walkFarmMaxLowStockLines {
			break
		}
		rem := numericToFloat64(row.CurrentQuantityRemaining)
		thr := numericToFloat64(row.LowStockThreshold)
		out = append(out, walkFinding{
			Category:    "supplies",
			Severity:    "warn",
			PlainText:   fmt.Sprintf("%s batch has %.2f left (below %.2f threshold)", strings.TrimSpace(row.InputName), rem, thr),
			ActionRoute: "/operations/supplies",
		})
	}
	return out, nil
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func morningWalkthroughContextBlock(farmLabel string) string {
	name := strings.TrimSpace(farmLabel)
	if name == "" {
		name = "this farm"
	}
	return strings.TrimSpace(fmt.Sprintf(`Operator focus — morning walkthrough for %s.
You are doing a morning walkthrough. Report only what needs attention — skip categories with nothing to flag.
Use plain language. No schema terms. Cite zone name, not zone_id.
Prefer walk_farm read-tool results for ordering. Read-only unless the operator asks for a specific action.`, name))
}
