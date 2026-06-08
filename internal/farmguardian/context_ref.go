package farmguardian

import (
	"context"
	"fmt"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/zonephotos"
)

// ContextRef is the UI "Ask Guardian" anchor — which alert, cycle, zone, or
// dashboard route the operator opened the drawer from (Phase 29 WS6, route Phase 32 WS1).
type ContextRef struct {
	Type             string `json:"type"`
	ID               int64  `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	Path             string `json:"path,omitempty"`
	Tab              string `json:"tab,omitempty"`
	CropCycleID      int64  `json:"crop_cycle_id,omitempty"`
	PriorCropCycleID int64  `json:"prior_crop_cycle_id,omitempty"`
}

// BuildContextRefBlock loads the referenced row and renders a focused prompt
// block. Best-effort: empty string when lookup fails or farm scope mismatches.
// Route refs need no DB row — only path (and optional name hint).
// history is the ordered list of recently-visited route refs (most recent first).
func BuildContextRefBlock(ctx context.Context, q *db.Queries, farmID int64, ref ContextRef, history ...[]ContextRef) string {
	var nav []ContextRef
	if len(history) > 0 {
		nav = history[0]
	}
	refType := strings.ToLower(strings.TrimSpace(ref.Type))
	switch refType {
	case "route":
		return renderRouteContext(ref.Path, ref.Name, nav)
	}
	if q == nil || farmID <= 0 || ref.ID <= 0 {
		return ""
	}
	switch refType {
	case "alert":
		return renderAlertContext(ctx, q, farmID, ref.ID)
	case "crop_cycle", "cycle":
		return renderCropCycleContext(ctx, q, farmID, ref.ID)
	case "zone":
		return renderZoneContext(ctx, q, farmID, ref)
	default:
		return ""
	}
}

func renderRouteContext(path, nameHint string, history []ContextRef) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	label := strings.TrimSpace(nameHint)
	if label == "" {
		label = routeLabelFromPath(path)
	}
	var b strings.Builder
	b.WriteString("Operator UI context — currently viewing: " + label)
	b.WriteString("\nRoute path: " + path)

	// Page-specific hints so the Guardian adopts the right framing without
	// the operator needing to say which screen they are on.
	switch {
	case path == "/feeding":
		b.WriteString("\nFeed & water hub — prefer feeding plan language (next feed, last feed, reservoir). Use summarize_zone_fertigation when a room is in scope.")
	case path == "/operations/supplies" || path == "/inventory":
		b.WriteString("\nSupplies — on-hand batches and low-stock. Cite input names and quantities; use restock_priority or summarize_farm_low_stock for grounded stock answers. Do not promise Guardian can change stock levels — restock stays in Supplies UI (+ Add qty).")
	case path == "/operations/feeding":
		b.WriteString("\nFeeding (details) — farm-wide programs, nutrient tanks, EC targets. Not the daily Feed & water hub (/feeding).")
	case path == "/operations/money" || path == "/costs":
		b.WriteString("\nMoney — spend summary and receipts. Use summarize_farm_spending for month-by-category answers; summarize_cycle_cost when a grow is in scope. Plain language; hide GL/COA unless the operator is on the full costs editor (/costs).")
	case path == "/sensors":
		b.WriteString("\nSensors list — operator can see all sensor cards with wiring and reading status. Use summarize_device_health for GPIO/source and reading freshness — no need to ask the operator to read back config.")
	case path == "/actuators":
		b.WriteString("\nActuators list — relay and output wiring panel. Use summarize_device_health for relay channel assignments and Pi config sync status.")
	case path == "/schedules":
		b.WriteString("\nSchedules — automation rules that trigger actuators. Focus on schedule gaps, overlaps, or next-run times.")
	case path == "/comfort-targets" || path == "/setpoints":
		b.WriteString("\nTargets & schedules — comfort band editor. Focus on whether current readings are within the set bands.")
	case path == "/tasks":
		b.WriteString("\nTasks — operator task list. Focus on overdue items, upcoming deadlines, or unassigned work.")
	case path == "/alerts":
		b.WriteString("\nAlerts — unread alert inbox. Lead with highest-severity unread alerts and recommended actions.")
	case path == "/fertigation":
		b.WriteString("\nFertigation (technical) — EC/pH mixing programs. Use precise nutrient and dosing language.")
	case path == "/plants":
		b.WriteString("\nPlants — strain catalog and varieties. Distinguish catalog plants from active grow runs (summarize_active_grows). Start a new grow from zone Overview or Plants — not silent chat writes.")
	case path == "/pi-setup":
		b.WriteString("\nPi + HAT setup guide — operator is configuring Raspberry Pi hardware with Sequent Microsystems stacking relay HATs. Use summarize_device_health for live GPIO/relay assignments. Offer procedure wire-pi-relay-light.")
	case path == "/" || path == "":
		b.WriteString("\nDashboard — farm overview. Prefer high-level summaries; offer to drill down into specific zones or alerts.")
	case strings.HasPrefix(path, "/zones/"):
		b.WriteString("\nZone detail — single grow room view. Zone-scoped answers preferred; sensor readings and crop cycle context are available.")
	case strings.HasPrefix(path, "/sensors/"):
		b.WriteString("\nSensor detail — single sensor config and history. Use summarize_device_health for wiring pin, source, and reading staleness on this sensor's Pi.")
	case strings.HasPrefix(path, "/farms/") && strings.HasSuffix(path, "/setup"):
		b.WriteString("\nFarm setup wizard — guide the operator through adding zones, connecting a device, and setting comfort targets in that order. Prefer wizard actions over free-form config advice.")
	case strings.HasPrefix(path, "/farms/") && strings.HasSuffix(path, "/zones/new"):
		b.WriteString("\nAdd grow room wizard — zone creation happens in the wizard UI, not chat. Guide through name, type, and comfort targets.")
	case strings.HasPrefix(path, "/farms/") && strings.HasSuffix(path, "/devices/new"):
		b.WriteString("\nEdge device wizard — device registration and Pi config. Use summarize_device_health after registration for heartbeat and wiring status. Offer procedure wire-pi-relay-light.")
	}

	// Navigation breadcrumb — show where the operator came from so the Guardian
	// understands the intent journey (e.g. Dashboard → Farm setup → Add zone).
	if len(history) > 0 {
		var trail []string
		for _, h := range history {
			if h.Path == "" {
				continue
			}
			hl := strings.TrimSpace(h.Name)
			if hl == "" {
				hl = routeLabelFromPath(h.Path)
			}
			trail = append(trail, hl)
		}
		if len(trail) > 0 {
			b.WriteString("\nNavigation trail (most recent first): " + strings.Join(trail, " → "))
		}
	}

	b.WriteString("\nUse this page context to answer without asking the operator to describe their screen.")
	return b.String()
}

func routeLabelFromPath(path string) string {
	if label, ok := knownRouteLabels[path]; ok {
		return label
	}
	if label := setupWizardRouteLabel(path); label != "" {
		return label
	}
	switch {
	case strings.HasPrefix(path, "/zones/"):
		return "Zone detail"
	case strings.HasPrefix(path, "/sensors/"):
		return "Sensor detail"
	case strings.Contains(path, "/crop-cycles/") && strings.HasSuffix(path, "/summary"):
		return "Crop cycle summary"
	case strings.Contains(path, "/crop-cycles/compare"):
		return "Crop cycle compare"
	default:
		return path
	}
}

var knownRouteLabels = map[string]string{
	"/":                  "Dashboard",
	"/zones":             "Zones",
	"/sensors":           "Sensors",
	"/actuators":         "Actuators",
	"/schedules":         "Schedules",
	"/automation":        "Automation",
	"/feeding":              "Feed & water",
	"/operations/supplies":  "Supplies",
	"/operations/feeding":   "Feeding (details)",
	"/operations/money":     "Money",
	"/fertigation":          "Feeding (technical)",
	"/comfort-targets":      "Targets & schedules",
	"/setpoints":            "Comfort bands",
	"/tasks":                "Tasks",
	"/inventory":            "Supplies (full editor)",
	"/costs":                "Money (full editor)",
	"/alerts":            "Alerts",
	"/plants":            "Plants",
	"/animals":           "Animals",
	"/aquaponics":        "Aquaponics",
	"/catalog":           "Commons catalog",
	"/farm-knowledge":    "Farm knowledge",
	"/chat":              "Farm Guardian chat",
	"/guardian/requests": "Guardian change requests",
	"/settings":          "Settings",
	"/operator-guide":    "Operator guide",
	"/pi-setup":          "Pi + HAT setup guide",
}

func setupWizardRouteLabel(path string) string {
	switch {
	case strings.HasPrefix(path, "/farms/") && strings.HasSuffix(path, "/setup"):
		return "Farm setup"
	case strings.HasPrefix(path, "/farms/") && strings.HasSuffix(path, "/zones/new"):
		return "Add grow room"
	case strings.HasPrefix(path, "/farms/") && strings.HasSuffix(path, "/devices/new"):
		return "Connect edge device"
	default:
		return ""
	}
}

func renderAlertContext(ctx context.Context, q *db.Queries, farmID, alertID int64) string {
	a, err := q.GetAlertNotificationByID(ctx, alertID)
	if err != nil || a.FarmID != farmID {
		return ""
	}
	detail := toUnreadAlertDetail(db.ListRecentUnreadAlertsByFarmRow{
		ID:                        a.ID,
		Severity:                  a.Severity,
		SubjectRendered:           a.SubjectRendered,
		MessageTextRendered:       a.MessageTextRendered,
		TriggeringEventSourceType: a.TriggeringEventSourceType,
		TriggeringEventSourceID:   a.TriggeringEventSourceID,
		CreatedAt:                 a.CreatedAt,
	})
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Operator focus — alert #%d", detail.ID))
	if detail.Severity != "" {
		b.WriteString(fmt.Sprintf(" (%s severity)", detail.Severity))
	}
	b.WriteByte('\n')
	if detail.Subject != "" {
		b.WriteString("Subject: " + detail.Subject + "\n")
	}
	if detail.Message != "" {
		b.WriteString("Message: " + detail.Message + "\n")
	}
	if detail.SourceType != "" {
		b.WriteString(fmt.Sprintf("Source: %s", detail.SourceType))
		if detail.SourceID > 0 {
			b.WriteString(fmt.Sprintf(" #%d", detail.SourceID))
		}
		b.WriteByte('\n')
	}
	read := "unread"
	if a.IsRead != nil && *a.IsRead {
		read = "read"
	}
	ack := "unacknowledged"
	if a.IsAcknowledged != nil && *a.IsAcknowledged {
		ack = "acknowledged"
	}
	b.WriteString(fmt.Sprintf("Status: %s, %s", read, ack))
	return b.String()
}

func renderCropCycleContext(ctx context.Context, q *db.Queries, farmID, cycleID int64) string {
	c, err := q.GetCropCycleByID(ctx, cycleID)
	if err != nil || c.FarmID != farmID {
		return ""
	}
	ac := ActiveCycle{ID: c.ID, Name: c.Name}
	if z, zerr := q.GetZoneByID(ctx, c.ZoneID); zerr == nil {
		ac.ZoneName = z.Name
	}
	if c.StrainOrVariety != nil {
		ac.Strain = *c.StrainOrVariety
	}
	if c.CurrentStage != nil {
		ac.Stage = string(*c.CurrentStage)
	}
	if a, aerr := fetchCycleAnalytics(ctx, q, c); aerr == nil {
		ac.Analytics = a
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Operator focus — crop cycle #%d: %s", c.ID, c.Name))
	if ac.ZoneName != "" {
		b.WriteString(fmt.Sprintf(" (zone %s)", ac.ZoneName))
	}
	b.WriteByte('\n')
	if ac.Strain != "" {
		b.WriteString("Strain: " + ac.Strain + "\n")
	}
	if ac.Stage != "" {
		b.WriteString("Stage: " + ac.Stage + "\n")
	}
	if !c.IsActive {
		b.WriteString("Status: harvested/inactive\n")
	} else {
		b.WriteString("Status: active\n")
	}
	if line := ac.Analytics.renderLine(); line != "" {
		b.WriteString("Metrics: " + line + "\n")
	}
	if hint := growAdvisorBriefLine(ctx, q, farmID, c.ZoneID, c.ID); hint != "" {
		b.WriteString(hint + "\n")
	}
	if costLine := cycleCostSummaryLine(ctx, q, c.ID); costLine != "" {
		b.WriteString(costLine + "\n")
	}
	if c.IsActive {
		b.WriteString("Prefer grow_advisor and lookup_crop_targets for VPD, DLI, and stage transition advice.")
	}
	return strings.TrimSpace(b.String())
}

// zoneTabConnectionPipelineHint documents the interactive sidebar pipeline on zone tabs (Phase 54).
func zoneTabConnectionPipelineHint(tab string) string {
	switch strings.TrimSpace(tab) {
	case "water":
		return "\nConnection chain (zone Water tab): sensor reading → target band → automation or feed timing → pump/light/fan → edge device."
	case "climate", "air":
		return "\nConnection chain (zone Climate tab): sensor reading → target band → automation → pump/light/fan → edge device."
	case "light":
		return "\nConnection chain (zone Light tab): sensor reading → target band → automation → grow light → edge device."
	default:
		return ""
	}
}

func renderZoneContext(ctx context.Context, q *db.Queries, farmID int64, ref ContextRef) string {
	z, err := q.GetZoneByID(ctx, ref.ID)
	if err != nil || z.FarmID != farmID {
		return ""
	}
	name := z.Name
	if name == "" {
		name = strings.TrimSpace(ref.Name)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Operator focus — zone #%d: %s", z.ID, name))
	if tab := strings.TrimSpace(ref.Tab); tab == "water" {
		b.WriteString(" (Water / feeding plan tab)")
		b.WriteString("\nPrefer feeding plan language — next feed, last feed, reservoir, water-only. Use summarize_zone_fertigation for program details.")
		b.WriteString(zoneTabConnectionPipelineHint(tab))
	} else if tab != "" {
		b.WriteString(fmt.Sprintf(" (%s tab)", tab))
	}
	if z.ZoneType != nil && *z.ZoneType != "" {
		b.WriteString(fmt.Sprintf(" (%s)", *z.ZoneType))
	}
	b.WriteByte('\n')
	if z.Description != nil && strings.TrimSpace(*z.Description) != "" {
		b.WriteString("Description: " + strings.TrimSpace(*z.Description) + "\n")
	}
	// Phase 33 WS2: carry the same latest sensor readings summarize_zone would
	// render, so this focus block is the single, complete zone block for the
	// turn (EnrichPromptBlock skips summarize_zone for this zone).
	if readings, rerr := renderZoneSensorReadings(ctx, q, z.ID); rerr == nil && readings != "" {
		b.WriteString(readings)
	}
	cycleID := ref.CropCycleID
	if cycleID <= 0 {
		if c, ok := activeCycleForZoneID(ctx, q, farmID, z.ID); ok {
			cycleID = c.ID
		}
	}
	if cycleID > 0 {
		if hint := growAdvisorBriefLine(ctx, q, farmID, z.ID, cycleID); hint != "" {
			b.WriteByte('\n')
			b.WriteString(hint)
		}
		if costLine := cycleCostSummaryLine(ctx, q, cycleID); costLine != "" {
			b.WriteByte('\n')
			b.WriteString(costLine)
		}
		b.WriteString("\nPrefer grow_advisor, lookup_crop_targets, summarize_cycle_cost, and cycle summary over generic farm snapshot when answering grow questions.")
	}
	if meta, _, err := zonephotos.ParseMeta(z.MetaData); err == nil && len(meta.PhotoAttachmentIDs) > 0 {
		b.WriteByte('\n')
		b.WriteString(fmt.Sprintf("Reference photos on file: %d", len(meta.PhotoAttachmentIDs)))
		if latest := zonephotos.LatestID(meta); latest > 0 {
			b.WriteString(fmt.Sprintf(" (latest attachment #%d)", latest))
		}
		b.WriteString(". Ask about this zone's walkthrough photos; pixel-level analysis requires a vision-capable model.")
	}
	return strings.TrimSpace(b.String())
}

// ContextRefPromptBlock wraps BuildContextRefBlock for the chat system prompt.
// navHistory (optional) is passed through to renderRouteContext so the Guardian
// receives the operator's breadcrumb trail alongside the current-page context.
func ContextRefPromptBlock(ctx context.Context, q *db.Queries, farmID int64, ref ContextRef, navHistory ...[]ContextRef) string {
	body := BuildContextRefBlock(ctx, q, farmID, ref, navHistory...)
	if body == "" {
		return ""
	}
	return "Contextual focus (background — do not cite as [n]):\n" + body
}
