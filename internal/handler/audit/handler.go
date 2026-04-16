package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

// ListByFarm — GET /farms/{id}/audit-events
func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireFarmAdmin(w, r, q, farmID) {
		return
	}
	limit := int32(50)
	if s := r.URL.Query().Get("limit"); s != "" {
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil || n < 1 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		if n > 200 {
			n = 200
		}
		limit = int32(n)
	}
	offset := int32(0)
	if s := r.URL.Query().Get("offset"); s != "" {
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil || n < 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = int32(n)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	rows, err := q.ListUserActivityLogByFarm(ctx, db.ListUserActivityLogByFarmParams{
		FarmID: &farmID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list audit events")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreUserActivityLog{}
	}

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
	httputil.WriteJSON(w, http.StatusOK, out)
}
