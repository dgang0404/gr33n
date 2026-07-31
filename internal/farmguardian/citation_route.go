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
		return resolveFertigationProgramCitationRoute(ctx, q, farmID, sourceID)
	case "task":
		return resolveTaskCitationRoute(ctx, q, farmID, sourceID)
	case "schedule":
		return resolveScheduleCitationRoute(ctx, q, farmID, sourceID)
	case "alert_notification":
		return resolveAlertCitationRoute(ctx, q, farmID, sourceID)
	case "field_guide", "platform_doc", "symptom_guide":
		return resolveDocCitationRoute(ctx, q, farmID, sourceType, sourceID)
	case "input_batch":
		return resolveInputBatchCitationRoute(ctx, q, farmID, sourceID)
	case "input_definition":
		return resolveInputDefinitionCitationRoute(ctx, q, farmID, sourceID)
	case "cost_transaction":
		return resolveCostTransactionCitationRoute(ctx, q, farmID, sourceID)
	case "automation_rule":
		return resolveAutomationRuleCitationRoute(ctx, q, farmID, sourceID)
	case "executable_action":
		return resolveExecutableActionCitationRoute(ctx, q, farmID, sourceID)
	default:
		return "", false
	}
}

func resolveFertigationProgramCitationRoute(ctx context.Context, q *db.Queries, farmID, programID int64) (string, bool) {
	p, err := q.GetFertigationProgramByID(ctx, programID)
	if err != nil || p.FarmID != farmID {
		return "", false
	}
	if p.TargetZoneID != nil && *p.TargetZoneID > 0 {
		return zonePath(*p.TargetZoneID, "water", ""), true
	}
	// Recipe-pack / unassigned programs — Feed & water Programs hub.
	return "/feed-water?tab=programs", true
}

func resolveTaskCitationRoute(ctx context.Context, q *db.Queries, farmID, taskID int64) (string, bool) {
	t, err := q.GetTaskByID(ctx, taskID)
	if err != nil || t.FarmID != farmID {
		return "", false
	}
	if t.ZoneID != nil && *t.ZoneID > 0 {
		return zonePath(*t.ZoneID, "ops", "tasks"), true
	}
	return "/zones", true
}

func resolveInputBatchCitationRoute(ctx context.Context, q *db.Queries, farmID, batchID int64) (string, bool) {
	b, err := q.GetInputBatchByID(ctx, batchID)
	if err != nil || b.FarmID != farmID {
		return "", false
	}
	return "/money?tab=supplies&batch_id=" + strconv.FormatInt(batchID, 10), true
}

func resolveInputDefinitionCitationRoute(ctx context.Context, q *db.Queries, farmID, defID int64) (string, bool) {
	d, err := q.GetInputDefinitionByID(ctx, defID)
	if err != nil || d.FarmID != farmID {
		return "", false
	}
	return "/natural-farming?tab=batch", true
}

func resolveCostTransactionCitationRoute(ctx context.Context, q *db.Queries, farmID, txID int64) (string, bool) {
	tx, err := q.GetCostTransactionByID(ctx, txID)
	if err != nil || tx.FarmID != farmID {
		return "", false
	}
	if tx.CropCycleID != nil && *tx.CropCycleID > 0 {
		return "/crop-cycles/" + strconv.FormatInt(*tx.CropCycleID, 10) + "/summary", true
	}
	return "/money?tab=ledger", true
}

func resolveAutomationRuleCitationRoute(ctx context.Context, q *db.Queries, farmID, ruleID int64) (string, bool) {
	rule, err := q.GetAutomationRuleByID(ctx, ruleID)
	if err != nil || rule.FarmID != farmID {
		return "", false
	}
	if z := ruleZoneIDFromConfig(rule.TriggerConfiguration); z != nil && *z > 0 {
		return zonePath(*z, "ops", "automations"), true
	}
	if z := zoneIDFromRuleConditions(ctx, q, rule); z != nil && *z > 0 {
		return zonePath(*z, "ops", "automations"), true
	}
	return "/comfort-targets?tab=automations", true
}

func resolveExecutableActionCitationRoute(ctx context.Context, q *db.Queries, farmID, actionID int64) (string, bool) {
	a, err := q.GetExecutableActionByID(ctx, actionID)
	if err != nil {
		return "", false
	}
	farmOwned := false
	if a.ScheduleID != nil && *a.ScheduleID > 0 {
		if s, err := q.GetScheduleByID(ctx, *a.ScheduleID); err == nil && s.FarmID == farmID {
			farmOwned = true
			if route, ok := resolveScheduleCitationRoute(ctx, q, farmID, *a.ScheduleID); ok {
				return route, true
			}
		}
	}
	if a.RuleID != nil && *a.RuleID > 0 {
		if rule, err := q.GetAutomationRuleByID(ctx, *a.RuleID); err == nil && rule.FarmID == farmID {
			farmOwned = true
			if route, ok := resolveAutomationRuleCitationRoute(ctx, q, farmID, *a.RuleID); ok {
				return route, true
			}
		}
	}
	if a.ProgramID != nil && *a.ProgramID > 0 {
		if p, err := q.GetFertigationProgramByID(ctx, *a.ProgramID); err == nil && p.FarmID == farmID {
			farmOwned = true
			if route, ok := resolveFertigationProgramCitationRoute(ctx, q, farmID, *a.ProgramID); ok {
				return route, true
			}
		}
	}
	if !farmOwned {
		return "", false
	}
	return "/comfort-targets?tab=schedules", true
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
	// Farm-owned schedule with no zone hop — Comfort schedules hub.
	return "/comfort-targets?tab=schedules", true
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
	if ok && zoneID > 0 {
		return zonePath(zoneID, "ops", "alerts"), true
	}
	// Farm-wide inbox when the trigger sensor/rule has no zone (common for gate/test sensors).
	return "/alerts", true
}

func zoneIDFromAlertTrigger(ctx context.Context, q *db.Queries, alert db.Gr33ncoreAlertsNotification) (int64, bool) {
	if alert.TriggeringEventSourceType == nil || alert.TriggeringEventSourceID == nil {
		return 0, false
	}
	srcType := strings.TrimSpace(*alert.TriggeringEventSourceType)
	srcID := *alert.TriggeringEventSourceID
	switch srcType {
	// Handler stores the sensor id under type sensor_reading (not a reading row id).
	case "sensor_reading", "sensor":
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
		return "/operator-guide?tab=symptoms&crop_key=" + url.QueryEscape(cropKey), true
	}
	if docPath != "" {
		switch sourceType {
		case "platform_doc":
			return docCitationRoute("library", docPath, "platform_doc", "guide"), true
		case "symptom_guide":
			return docCitationRoute("symptoms", docPath, "symptom_guide", ""), true
		default:
			return docCitationRoute("knowledge", docPath, "field_guide", ""), true
		}
	}
	return landingDocRoute(sourceType, cropKey)
}

func docCitationRoute(tab, docPath, citedType string, section string) string {
	v := url.Values{}
	v.Set("tab", tab)
	if section != "" {
		v.Set("section", section)
	}
	v.Set("cited_doc", docPath)
	v.Set("cited_type", citedType)
	return "/operator-guide?" + v.Encode()
}

func landingDocRoute(sourceType, cropKey string) (string, bool) {
	if sourceType == "field_guide" {
		if cropKey != "" {
			return "/operator-guide?tab=symptoms&crop_key=" + url.QueryEscape(cropKey), true
		}
		return "/operator-guide?tab=knowledge", true
	}
	if sourceType == "platform_doc" {
		return "/operator-guide?tab=library&section=guide", true
	}
	if sourceType == "symptom_guide" {
		return "/operator-guide?tab=symptoms", true
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
