package farmguardian

import (
	"context"
	"fmt"
	"strings"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/zonephotos"
)

// ContextRef is the UI "Ask Guardian" anchor — which alert, cycle, or zone
// the operator opened the drawer from (Phase 29 WS6).
type ContextRef struct {
	Type string `json:"type"`
	ID   int64  `json:"id"`
	Name string `json:"name,omitempty"`
}

// BuildContextRefBlock loads the referenced row and renders a focused prompt
// block. Best-effort: empty string when lookup fails or farm scope mismatches.
func BuildContextRefBlock(ctx context.Context, q *db.Queries, farmID int64, ref ContextRef) string {
	if q == nil || farmID <= 0 || ref.ID <= 0 {
		return ""
	}
	refType := strings.ToLower(strings.TrimSpace(ref.Type))
	switch refType {
	case "alert":
		return renderAlertContext(ctx, q, farmID, ref.ID)
	case "crop_cycle", "cycle":
		return renderCropCycleContext(ctx, q, farmID, ref.ID)
	case "zone":
		return renderZoneContext(ctx, q, farmID, ref.ID, ref.Name)
	default:
		return ""
	}
}

func renderAlertContext(ctx context.Context, q *db.Queries, farmID, alertID int64) string {
	a, err := q.GetAlertNotificationByID(ctx, alertID)
	if err != nil || a.FarmID != farmID {
		return ""
	}
	detail := toUnreadAlertDetail(db.RecentUnreadAlertSummary{
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
	if c.CurrentStage.Valid {
		ac.Stage = string(c.CurrentStage.Gr33nfertigationGrowthStageEnum)
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
	return strings.TrimSpace(b.String())
}

func renderZoneContext(ctx context.Context, q *db.Queries, farmID, zoneID int64, nameHint string) string {
	z, err := q.GetZoneByID(ctx, zoneID)
	if err != nil || z.FarmID != farmID {
		return ""
	}
	name := z.Name
	if name == "" {
		name = strings.TrimSpace(nameHint)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Operator focus — zone #%d: %s", z.ID, name))
	if z.ZoneType != nil && *z.ZoneType != "" {
		b.WriteString(fmt.Sprintf(" (%s)", *z.ZoneType))
	}
	b.WriteByte('\n')
	if z.Description != nil && strings.TrimSpace(*z.Description) != "" {
		b.WriteString("Description: " + strings.TrimSpace(*z.Description) + "\n")
	}
	sensors, _ := q.ListSensorsByZone(ctx, &zoneID)
	if len(sensors) > 0 {
		b.WriteString(fmt.Sprintf("Sensors in zone: %d", len(sensors)))
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
func ContextRefPromptBlock(ctx context.Context, q *db.Queries, farmID int64, ref ContextRef) string {
	body := BuildContextRefBlock(ctx, q, farmID, ref)
	if body == "" {
		return ""
	}
	return "Contextual focus (background — do not cite as [n]):\n" + body
}
