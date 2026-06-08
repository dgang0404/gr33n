// Phase 61 WS1 — rule-based proactive Guardian nudge (one per farm, priority order).

package farmguardian

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "gr33n-api/internal/db"
)

const (
	nudgeAlertMinAge       = 15 * time.Minute
	nudgeComfortBreachMin  = 30 * time.Minute
	nudgePiStaleAfter      = 2 * time.Hour
	nudgeFeedMissedGrace   = 30 * time.Minute
)

// NudgePayload is returned by GET /farms/{id}/guardian-nudge.
type NudgePayload struct {
	Category    string `json:"category"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	ActionRoute string `json:"action_route"`
	NudgeID     string `json:"nudge_id"`
}

// ComputeGuardianNudge returns the highest-priority nudge for a farm, or nil when none.
func ComputeGuardianNudge(ctx context.Context, q db.Querier, farmID int64) (*NudgePayload, error) {
	if q == nil || farmID <= 0 {
		return nil, nil
	}
	zones, _ := q.ListZonesByFarm(ctx, farmID)
	zoneName := map[int64]string{}
	for _, z := range zones {
		zoneName[z.ID] = strings.TrimSpace(z.Name)
	}

	if n, err := nudgeCriticalAlert(ctx, q, farmID); err != nil {
		return nil, err
	} else if n != nil {
		return n, nil
	}
	if n, err := nudgeFeedMissed(ctx, q, farmID); err != nil {
		return nil, err
	} else if n != nil {
		return n, nil
	}
	if n, err := nudgeComfortBreach(ctx, q, farmID, zoneName); err != nil {
		return nil, err
	} else if n != nil {
		return n, nil
	}
	if n, err := nudgePiStale(ctx, q, farmID); err != nil {
		return nil, err
	} else if n != nil {
		return n, nil
	}
	return nudgeLowStock(ctx, q, farmID)
}

func nudgeCriticalAlert(ctx context.Context, q db.Querier, farmID int64) (*NudgePayload, error) {
	alerts, err := q.ListAlertsByFarm(ctx, db.ListAlertsByFarmParams{FarmID: farmID, Limit: 30, Offset: 0})
	if err != nil {
		return nil, err
	}
	now := nowFunc()
	for _, a := range alerts {
		if a.IsAcknowledged != nil && *a.IsAcknowledged {
			continue
		}
		if !alertSeverityWarnOrHigher(a.Severity) {
			continue
		}
		if now.Sub(a.CreatedAt) < nudgeAlertMinAge {
			continue
		}
		subject := strings.TrimSpace(ptrStr(a.SubjectRendered))
		if subject == "" {
			subject = strings.TrimSpace(ptrStr(a.MessageTextRendered))
		}
		if subject == "" {
			subject = "Alert needs review"
		}
		return &NudgePayload{
			Category:    "critical_alert",
			Message:     subject + " — tap to review",
			Severity:    "warn",
			ActionRoute: "/alerts",
			NudgeID:     fmt.Sprintf("alert-%d", a.ID),
		}, nil
	}
	return nil, nil
}

func alertSeverityWarnOrHigher(sev *db.Gr33ncoreNotificationPriorityEnum) bool {
	if sev == nil {
		return true
	}
	switch *sev {
	case db.Gr33ncoreNotificationPriorityEnumLow:
		return false
	default:
		return true
	}
}

func nudgeFeedMissed(ctx context.Context, q db.Querier, farmID int64) (*NudgePayload, error) {
	schedules, err := q.ListSchedulesByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	now := nowFunc()
	for _, s := range schedules {
		if !s.IsActive || !s.NextExpectedTriggerTime.Valid {
			continue
		}
		missedAt := s.NextExpectedTriggerTime.Time
		if now.Sub(missedAt) < nudgeFeedMissedGrace {
			continue
		}
		if s.LastTriggeredTime.Valid && !s.LastTriggeredTime.Time.Before(missedAt) {
			continue
		}
		label := strings.TrimSpace(s.Name)
		if label == "" {
			label = "Scheduled feed"
		}
		when := humanizeAge(now.Sub(missedAt)) + " ago"
		return &NudgePayload{
			Category:    "feed_missed",
			Message:     fmt.Sprintf("%s hasn't run (due %s) — is that intentional?", label, when),
			Severity:    "warn",
			ActionRoute: "/feeding",
			NudgeID:     fmt.Sprintf("schedule-%d", s.ID),
		}, nil
	}
	return nil, nil
}

func nudgeComfortBreach(ctx context.Context, q db.Querier, farmID int64, zoneName map[int64]string) (*NudgePayload, error) {
	setpoints, err := q.ListSetpointsByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	sensors, err := q.ListSensorsByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	sensorByZoneType := map[int64]map[string]db.Gr33ncoreSensor{}
	for _, s := range sensors {
		if s.ZoneID == nil || s.DeletedAt.Valid {
			continue
		}
		if !walkFarmClimateSensorType(s.SensorType) {
			continue
		}
		zid := *s.ZoneID
		if sensorByZoneType[zid] == nil {
			sensorByZoneType[zid] = make(map[string]db.Gr33ncoreSensor)
		}
		sensorByZoneType[zid][s.SensorType] = s
	}

	now := nowFunc()
	for _, sp := range setpoints {
		if sp.ZoneID == nil || sp.CropCycleID != nil || !walkFarmClimateSensorType(sp.SensorType) {
			continue
		}
		if !sp.MinValue.Valid && !sp.MaxValue.Valid {
			continue
		}
		zid := *sp.ZoneID
		sensor, ok := sensorByZoneType[zid][sp.SensorType]
		if !ok {
			continue
		}
		if !sensor.AlertBreachStartedAt.Valid {
			continue
		}
		if now.Sub(sensor.AlertBreachStartedAt.Time) < nudgeComfortBreachMin {
			continue
		}
		reading, val, have := latestZoneSensorReading(ctx, q, zid, sp.SensorType)
		if !have || val == nil {
			continue
		}
		breach := comfortValueBreach(sp, *val)
		if breach == "" {
			continue
		}
		_ = reading
		zn := zoneName[zid]
		if zn == "" {
			zn = fmt.Sprintf("zone #%d", zid)
		}
		mins := int(now.Sub(sensor.AlertBreachStartedAt.Time).Minutes())
		return &NudgePayload{
			Category:    "comfort_breach",
			Message:     fmt.Sprintf("%s in %s — %s for %d minutes", comfortSensorLabel(sp.SensorType), zn, breach, mins),
			Severity:    "warn",
			ActionRoute: fmt.Sprintf("/comfort-targets?zone_id=%d", zid),
			NudgeID:     fmt.Sprintf("comfort-%d-%s", zid, sp.SensorType),
		}, nil
	}
	return nil, nil
}

func nudgePiStale(ctx context.Context, q db.Querier, farmID int64) (*NudgePayload, error) {
	devices, err := q.ListDevicesByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	for _, d := range devices {
		if d.DeletedAt.Valid || d.ConfigVersion <= 0 {
			continue
		}
		age, ok := deviceConfigSyncAge(d.Config)
		if !ok || age < nudgePiStaleAfter {
			continue
		}
		name := strings.TrimSpace(d.Name)
		if name == "" {
			name = "Pi device"
		}
		return &NudgePayload{
			Category:    "pi_stale",
			Message:     fmt.Sprintf("%s hasn't checked in — worth a look", name),
			Severity:    "warn",
			ActionRoute: "/",
			NudgeID:     fmt.Sprintf("device-%d", d.ID),
		}, nil
	}
	return nil, nil
}

func nudgeLowStock(ctx context.Context, q db.Querier, farmID int64) (*NudgePayload, error) {
	rows, err := q.ListLowStockBatchesByFarm(ctx, farmID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	name := strings.TrimSpace(row.InputName)
	if name == "" {
		name = "Supply"
	}
	return &NudgePayload{
		Category:    "low_stock",
		Message:     fmt.Sprintf("%s is almost out — create a refill task?", name),
		Severity:    "warn",
		ActionRoute: "/operations/supplies",
		NudgeID:     fmt.Sprintf("stock-%d", row.ID),
	}, nil
}

// NudgeContextBlock frames a Guardian turn when the operator tapped Review on a nudge.
func NudgeContextBlock(ref ContextRef) string {
	cat := strings.TrimSpace(ref.NudgeCategory)
	if cat == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("Operator focus — reviewing a Guardian proactive nudge")
	if cat != "" {
		b.WriteString(" (" + cat + ")")
	}
	if id := strings.TrimSpace(ref.NudgeID); id != "" {
		b.WriteString(" nudge_id=" + id)
	}
	b.WriteString(".\nSkip pleasantries — address the specific issue immediately. Read-only unless the operator asks for a change request.")
	return b.String()
}
