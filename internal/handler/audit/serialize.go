package audit

import (
	"encoding/json"

	"github.com/google/uuid"

	db "gr33n-api/internal/db"
)

func activityRowsToJSON(rows []db.Gr33ncoreUserActivityLog) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		var details any = map[string]any{}
		if len(row.Details) > 0 {
			if err := json.Unmarshal(row.Details, &details); err != nil {
				details = json.RawMessage(row.Details)
			}
		}
		item := map[string]any{
			"id":            row.ID,
			"activity_time": row.ActivityTime,
			"action_type":   row.ActionType,
			"details":       details,
			"created_at":    row.CreatedAt,
		}
		if row.UserID.Valid {
			item["user_id"] = uuid.UUID(row.UserID.Bytes).String()
		}
		if row.FarmID != nil {
			item["farm_id"] = *row.FarmID
		}
		if row.TargetModuleSchema != nil {
			item["target_module_schema"] = *row.TargetModuleSchema
		}
		if row.TargetTableName != nil {
			item["target_table_name"] = *row.TargetTableName
		}
		if row.TargetRecordID != nil {
			item["target_record_id"] = *row.TargetRecordID
		}
		if row.TargetRecordDescription != nil {
			item["target_record_description"] = *row.TargetRecordDescription
		}
		if row.UserAgent != nil {
			item["user_agent"] = *row.UserAgent
		}
		if row.Status != nil {
			item["status"] = *row.Status
		}
		if row.FailureReason != nil {
			item["failure_reason"] = *row.FailureReason
		}
		if row.IpAddress != nil {
			item["ip_address"] = row.IpAddress.String()
		}
		out = append(out, item)
	}
	return out
}
