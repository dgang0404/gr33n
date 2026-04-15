package alert

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct{ pool *pgxpool.Pool }

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

func (h *Handler) ListByFarm(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	q := db.New(h.pool)
	if !farmauthz.RequireFarmMember(w, r, q, farmID) {
		return
	}
	rows, err := q.ListAlertsByFarm(r.Context(), db.ListAlertsByFarmParams{
		FarmID: farmID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreAlertsNotification{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

func (h *Handler) CountUnread(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireFarmMember(w, r, q, farmID) {
		return
	}
	count, err := q.CountUnreadAlertsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]int64{"unread_count": count})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid alert id")
		return
	}
	q := db.New(h.pool)
	a0, err := q.GetAlertNotificationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "alert not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, q, a0.FarmID) {
		return
	}
	alert, err := q.MarkAlertRead(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, alert)
}

func (h *Handler) MarkAcknowledged(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid alert id")
		return
	}
	q := db.New(h.pool)
	a0, err := q.GetAlertNotificationByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "alert not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, q, a0.FarmID) {
		return
	}
	alert, err := q.MarkAlertAcknowledged(r.Context(), db.MarkAlertAcknowledgedParams{
		ID:                   id,
		AcknowledgedByUserID: pgtype.UUID{},
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, alert)
}
