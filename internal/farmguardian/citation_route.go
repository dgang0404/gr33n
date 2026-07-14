// Phase 152 WS2 + Phase 159 WS2b — citation deep links.

package farmguardian

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	db "gr33n-api/internal/db"
)

// ResolveCitationRoute maps a citation to a UI path, or ok=false when the
// source type has no route yet, the row can't be found, or it belongs to a
// different farm (defense in depth beyond the already farm-scoped RAG
// retrieval — a citation should never be able to route a click into another
// farm's data).
func ResolveCitationRoute(ctx context.Context, q *db.Queries, farmID int64, sourceType string, sourceID int64) (string, bool) {
	if q == nil || farmID <= 0 || sourceID <= 0 {
		return "", false
	}
	switch sourceType {
	case "crop_cycle":
		c, err := q.GetCropCycleByID(ctx, sourceID)
		if err != nil || c.FarmID != farmID {
			return "", false
		}
		return "/crop-cycles/" + strconv.FormatInt(sourceID, 10) + "/summary", true
	case "fertigation_program":
		p, err := q.GetFertigationProgramByID(ctx, sourceID)
		if err != nil || p.FarmID != farmID || p.TargetZoneID == nil || *p.TargetZoneID <= 0 {
			return "", false
		}
		return zonePath(*p.TargetZoneID, "water", ""), true
	case "task":
		t, err := q.GetTaskByID(ctx, sourceID)
		if err != nil || t.FarmID != farmID || t.ZoneID == nil || *t.ZoneID <= 0 {
			return "", false
		}
		return zonePath(*t.ZoneID, "", ""), true
	case "schedule":
		return resolveScheduleCitationRoute(ctx, q, farmID, sourceID)
	case "alert_notification":
		return resolveAlertCitationRoute(ctx, q, farmID, sourceID)
	case "field_guide", "platform_doc":
		return resolveDocCitationRoute(ctx, q, farmID, sourceType, sourceID)
	default:
		return "", false
	}
}

func resolveScheduleCitationRoute(ctx context.Context, q *db.Queries, farmID, scheduleID int64) (string, bool) {
	s, err := q.GetScheduleByID(ctx, scheduleID)
	if err != nil || s.FarmID != farmID {
		return "", false
	}
	isLighting := strings.EqualFold(strings.TrimSpace(s.ScheduleType), "lighting")

	if zonePtr, err := q.GetFertigationProgramZoneBySchedule(ctx, db.GetFertigationProgramZoneByScheduleParams{
		FarmID:     farmID,
		ScheduleID: &scheduleID,
	}); err == nil && zonePtr != nil && *zonePtr > 0 {
		return zonePath(*zonePtr, "water", ""), true
	}
	if zoneID, err := q.GetLightingProgramZoneBySchedule(ctx, db.GetLightingProgramZoneByScheduleParams{
		FarmID:     farmID,
		ScheduleID: &scheduleID,
	}); err == nil && zoneID > 0 {
		return zonePath(zoneID, "light", ""), true
	}
	if zonePtr, err := q.GetActuatorZoneBySchedule(ctx, db.GetActuatorZoneByScheduleParams{
		ScheduleID: &scheduleID,
		FarmID:     farmID,
	}); err == nil && zonePtr != nil && *zonePtr > 0 {
		if isLighting {
			return zonePath(*zonePtr, "light", ""), true
		}
		return zonePath(*zonePtr, "ops", "automations"), true
	}
	if zoneID, ok := zoneFromScheduleNameHint(ctx, q, farmID, s); ok {
		if isLighting {
			return zonePath(zoneID, "light", ""), true
		}
		return zonePath(zoneID, "water", ""), true
	}
	return "", false
}

// zoneFromScheduleNameHint resolves legacy orphan schedules (bootstrap lighting
// pairs without lighting_programs or executable_actions) by matching the
// schedule name/description to a zone label on the farm.
func zoneFromScheduleNameHint(ctx context.Context, q *db.Queries, farmID int64, s db.Gr33ncoreSchedule) (int64, bool) {
	zones, err := q.ListZonesByFarm(ctx, farmID)
	if err != nil || len(zones) == 0 {
		return 0, false
	}
	nameLower := strings.ToLower(strings.TrimSpace(s.Name))
	var bestID int64
	bestScore := 0
	for _, z := range zones {
		if scheduleDescribesZone(s, z.Name) {
			return z.ID, true
		}
		zoneName := strings.TrimSpace(z.Name)
		if zoneName == "" {
			continue
		}
		zoneLower := strings.ToLower(zoneName)
		score := 0
		if strings.Contains(nameLower, zoneLower) {
			score = len(zoneLower)
		} else {
			// "Light ON 12/12 Flower" ↔ zone "Flower Room"
			for _, word := range strings.Fields(zoneLower) {
				if len(word) < 3 {
					continue
				}
				if strings.Contains(nameLower, word) && len(word) > score {
					score = len(word)
				}
			}
		}
		if score > bestScore {
			bestScore = score
			bestID = z.ID
		}
	}
	if bestScore > 0 {
		return bestID, true
	}
	return 0, false
}

func resolveAlertCitationRoute(ctx context.Context, q *db.Queries, farmID, alertID int64) (string, bool) {
	alert, err := q.GetAlertNotificationByID(ctx, alertID)
	if err != nil || alert.FarmID != farmID {
		return "", false
	}
	zoneID, ok := zoneIDFromAlertTrigger(ctx, q, alert)
	if !ok || zoneID <= 0 {
		return "", false
	}
	return zonePath(zoneID, "ops", "alerts"), true
}

func zoneIDFromAlertTrigger(ctx context.Context, q *db.Queries, alert db.Gr33ncoreAlertsNotification) (int64, bool) {
	if alert.TriggeringEventSourceType == nil || alert.TriggeringEventSourceID == nil {
		return 0, false
	}
	srcType := strings.TrimSpace(*alert.TriggeringEventSourceType)
	srcID := *alert.TriggeringEventSourceID
	switch srcType {
	case "sensor_reading":
		sensor, err := q.GetSensorByID(ctx, srcID)
		if err != nil || sensor.FarmID != alert.FarmID || sensor.ZoneID == nil {
			return 0, false
		}
		return *sensor.ZoneID, true
	case "automation_rule":
		rule, err := q.GetAutomationRuleByID(ctx, srcID)
		if err != nil || rule.FarmID != alert.FarmID {
			return 0, false
		}
		if z := ruleZoneIDFromConfig(rule.TriggerConfiguration); z != nil {
			return *z, true
		}
		if z := zoneIDFromRuleConditions(ctx, q, rule); z != nil {
			return *z, true
		}
		return 0, false
	case "automation_program":
		prog, err := q.GetFertigationProgramByID(ctx, srcID)
		if err != nil || prog.FarmID != alert.FarmID || prog.TargetZoneID == nil {
			return 0, false
		}
		return *prog.TargetZoneID, true
	default:
		return 0, false
	}
}

type ruleConditionsWire struct {
	Predicates []struct {
		SensorID int64 `json:"sensor_id"`
	} `json:"predicates"`
}

func zoneIDFromRuleConditions(ctx context.Context, q *db.Queries, rule db.Gr33ncoreAutomationRule) *int64 {
	if len(rule.ConditionsJsonb) == 0 {
		return nil
	}
	var wire ruleConditionsWire
	if err := json.Unmarshal(rule.ConditionsJsonb, &wire); err != nil {
		return nil
	}
	for _, p := range wire.Predicates {
		if p.SensorID <= 0 {
			continue
		}
		sensor, err := q.GetSensorByID(ctx, p.SensorID)
		if err != nil || sensor.FarmID != rule.FarmID || sensor.ZoneID == nil {
			continue
		}
		z := *sensor.ZoneID
		return &z
	}
	return nil
}

type ragDocCitationMeta struct {
	DocPath string `json:"doc_path"`
	CropKey string `json:"crop_key"`
}

func resolveDocCitationRoute(ctx context.Context, q *db.Queries, farmID int64, sourceType string, sourceID int64) (string, bool) {
	metaRaw, err := q.GetRagChunkMetadataByFarmSource(ctx, db.GetRagChunkMetadataByFarmSourceParams{
		FarmID:     farmID,
		SourceType: sourceType,
		SourceID:   sourceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return landingDocRoute(sourceType, "")
		}
		return "", false
	}
	var meta ragDocCitationMeta
	_ = json.Unmarshal(metaRaw, &meta)
	docPath := strings.TrimSpace(meta.DocPath)
	cropKey := strings.TrimSpace(meta.CropKey)
	if sourceType == "field_guide" && cropKey != "" {
		return "/symptom-guide?crop_key=" + url.QueryEscape(cropKey), true
	}
	if docPath != "" {
		if sourceType == "platform_doc" {
			return docCitationRoute("guide", docPath, "platform_doc"), true
		}
		return docCitationRoute("knowledge", docPath, "field_guide"), true
	}
	return landingDocRoute(sourceType, cropKey)
}

func docCitationRoute(section, docPath, citedType string) string {
	v := url.Values{}
	v.Set("tab", "library")
	v.Set("section", section)
	v.Set("cited_doc", docPath)
	v.Set("cited_type", citedType)
	return "/operator-guide?" + v.Encode()
}

func landingDocRoute(sourceType, cropKey string) (string, bool) {
	if sourceType == "field_guide" {
		if cropKey != "" {
			return "/symptom-guide?crop_key=" + url.QueryEscape(cropKey), true
		}
		return "/operator-guide?tab=library&section=knowledge", true
	}
	if sourceType == "platform_doc" {
		return "/operator-guide?tab=library&section=guide", true
	}
	return "", false
}

func zonePath(zoneID int64, tab, ops string) string {
	path := "/zones/" + strconv.FormatInt(zoneID, 10)
	q := url.Values{}
	if tab != "" {
		q.Set("tab", tab)
	}
	if ops != "" {
		q.Set("ops", ops)
	}
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	return path
}
