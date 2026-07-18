package audit

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

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
	limit, offset, err := httputil.ParseLimitOffsetStrict(r, 50, 200)
	if err != nil {
		if errors.Is(err, httputil.ErrInvalidLimit) {
			httputil.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		httputil.WriteError(w, http.StatusBadRequest, "invalid offset")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	rows, err := q.ListUserActivityLogByFarm(ctx, db.ListUserActivityLogByFarmParams{
		FarmID: &farmID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list audit events")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreUserActivityLog{}
	}

	httputil.WriteJSON(w, http.StatusOK, activityRowsToJSON(rows))
}

// ListByOrganization — GET /organizations/{id}/audit-events
func (h *Handler) ListByOrganization(w http.ResponseWriter, r *http.Request) {
	orgID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	q := db.New(h.pool)
	if !farmauthz.RequireOrgAdmin(w, r, q, orgID) {
		return
	}
	limit, offset, err := httputil.ParseLimitOffsetStrict(r, 50, 200)
	if err != nil {
		if errors.Is(err, httputil.ErrInvalidLimit) {
			httputil.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		httputil.WriteError(w, http.StatusBadRequest, "invalid offset")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	oid := orgID
	rows, err := q.ListUserActivityLogForOrganization(ctx, db.ListUserActivityLogForOrganizationParams{
		OrganizationID: &oid,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list audit events")
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreUserActivityLog{}
	}
	httputil.WriteJSON(w, http.StatusOK, activityRowsToJSON(rows))
}
