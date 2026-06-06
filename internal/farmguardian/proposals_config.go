package farmguardian

import (
	"regexp"
	"strings"
)

var (
	createTaskIntent = regexp.MustCompile(`(?i)\b(create|add|make)\b.*\btask\b|\btask\b.*\b(create|add)\b`)
	createTaskFromAlertIntent = regexp.MustCompile(`(?i)\b(task\s+from\s+alert|create\s+task\s+from|turn\s+.*alert\s+into\s+task)\b`)
	updateStageIntent = regexp.MustCompile(`(?i)\b(update|change|move|set)\b.*\b(stage|growth)\b|\b(veg|flower|harvest|seedling)\b.*\bstage\b`)
)

func matchConfigToolIntent(question string, snap Snapshot) (toolID string, args map[string]any, summary string, ok bool) {
	q := strings.TrimSpace(question)
	lower := strings.ToLower(q)

	if createTaskFromAlertIntent.MatchString(q) || (strings.Contains(lower, "task") && strings.Contains(lower, "alert")) {
		if len(snap.UnreadAlertDetails) == 0 {
			return "", nil, "", false
		}
		alert := pickAlertForIntent(question, snap.UnreadAlertDetails)
		title := taskTitleFromQuestion(question, alert)
		return "create_task_from_alert", map[string]any{
			"alert_id":          alert.ID,
			"title":             title,
			"alert_subject":     alert.Subject,
			"alert_source_type": alert.SourceType,
		}, "Create task: " + title, true
	}

	if createTaskIntent.MatchString(q) || (strings.Contains(lower, "check") && strings.Contains(lower, "humidity")) {
		title := taskTitleFromQuestion(question, UnreadAlertDetail{})
		args := map[string]any{"title": title}
		if zid, zname := pickZoneForIntent(question, snap); zid > 0 {
			args["zone_id"] = zid
			if !strings.Contains(strings.ToLower(title), strings.ToLower(zname)) && zname != "" {
				title = title + " — " + zname
				args["title"] = title
			}
		}
		return "create_task", args, "Create task: " + title, true
	}

	if updateStageIntent.MatchString(q) {
		cycle, stage, ok := pickCycleStageForIntent(question, snap)
		if !ok {
			return "", nil, "", false
		}
		return "update_cycle_stage", map[string]any{
			"crop_cycle_id": cycle.ID,
			"current_stage": stage,
		}, "Update stage → " + stage + " (" + cycle.Name + ")", true
	}

	return "", nil, "", false
}

func taskTitleFromQuestion(question string, alert UnreadAlertDetail) string {
	lower := strings.ToLower(question)
	if alert.SourceType == "inventory_low_stock" {
		if name := lowStockInputFromSubject(alert.Subject); name != "" {
			if strings.Contains(lower, "refill") || strings.Contains(lower, "restock") ||
				strings.Contains(lower, "reorder") || strings.Contains(lower, "low stock") ||
				strings.Contains(lower, "supplies") {
				return "Refill " + name
			}
			return "Refill supplies: " + name
		}
	}
	switch {
	case strings.Contains(lower, "humidity"):
		if alert.Subject != "" && strings.Contains(strings.ToLower(alert.Subject), "humidity") {
			return "Check: " + alert.Subject
		}
		return "Check humidity in grow room"
	case strings.Contains(lower, "inspect"), strings.Contains(lower, "check"):
		if alert.Subject != "" {
			return "Check: " + alert.Subject
		}
		return "Inspect and follow up"
	default:
		if alert.Subject != "" {
			return "Task: " + alert.Subject
		}
		return "Follow up from Guardian chat"
	}
}

func pickZoneForIntent(question string, snap Snapshot) (zoneID int64, zoneName string) {
	lower := strings.ToLower(question)
	for _, c := range snap.ActiveCycles {
		if c.ZoneName != "" && strings.Contains(lower, strings.ToLower(c.ZoneName)) {
			// zone id not in ActiveCycle — match by name in ZoneNames only
			zoneName = c.ZoneName
			break
		}
	}
	for _, name := range snap.ZoneNames {
		if strings.Contains(lower, strings.ToLower(name)) {
			zoneName = name
			break
		}
	}
	_ = zoneID
	return 0, zoneName
}

func pickCycleStageForIntent(question string, snap Snapshot) (ActiveCycle, string, bool) {
	if len(snap.ActiveCycles) == 0 {
		return ActiveCycle{}, "", false
	}
	lower := strings.ToLower(question)
	stage := inferStageKeyword(lower)
	cycle := snap.ActiveCycles[0]
	for _, c := range snap.ActiveCycles {
		if c.ZoneName != "" && strings.Contains(lower, strings.ToLower(c.ZoneName)) {
			cycle = c
			break
		}
		if c.Name != "" && strings.Contains(lower, strings.ToLower(c.Name)) {
			cycle = c
			break
		}
	}
	if stage == "" {
		return ActiveCycle{}, "", false
	}
	return cycle, stage, true
}

func inferStageKeyword(lower string) string {
	// Values must be valid gr33nfertigation.growth_stage_enum members so the
	// resulting advance-stage proposal applies cleanly on Confirm.
	stages := []struct{ kw, stage string }{
		{"flower", "early_flower"},
		{"bloom", "early_flower"},
		{"veg", "early_veg"},
		{"vegetative", "early_veg"},
		{"harvest", "harvest"},
		{"seedling", "seedling"},
		{"clone", "clone"},
		{"dry", "dry_cure"},
	}
	for _, s := range stages {
		if strings.Contains(lower, s.kw) {
			return s.stage
		}
	}
	return ""
}

// lowStockInputFromSubject parses "Inventory low: OHN at …" subjects from the worker.
func lowStockInputFromSubject(subject string) string {
	const prefix = "Inventory low:"
	if !strings.Contains(subject, prefix) {
		return ""
	}
	rest := strings.TrimSpace(strings.SplitN(subject, prefix, 2)[1])
	if at := strings.Index(rest, " at "); at > 0 {
		return strings.TrimSpace(rest[:at])
	}
	return strings.TrimSpace(rest)
}
