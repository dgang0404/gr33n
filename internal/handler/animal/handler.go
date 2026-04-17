// Package animal implements Phase 20.8 WS2 CRUD + lifecycle-event
// handlers for gr33nanimals.animal_groups and
// gr33nanimals.animal_lifecycle_events. The group row is the
// *head-count + timeline anchor*; the hardware layer (sensors,
// actuators, rules, tasks) is still used for feeding, watering, and
// climate — see docs/workflow-guide.md §12.
package animal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct{ q *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// ── animal_groups ──────────────────────────────────────────────────────────

// ListGroups — GET /farms/{id}/animal-groups
func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListAnimalGroupsByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nanimalsAnimalGroup{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// GetGroup — GET /animal-groups/{id}
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid animal group id")
		return
	}
	row, err := h.q.GetAnimalGroupByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "animal group not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, row.FarmID) {
		return
	}
	// Return the group plus the summed lifecycle deltas so the UI can
	// surface "stored count vs events delta" without a second round-trip.
	delta, _ := h.q.SumLifecycleDeltasByGroup(r.Context(), id)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"group":        row,
		"delta_total":  delta,
	})
}

type groupCreateReq struct {
	Label         string          `json:"label"`
	Species       *string         `json:"species"`
	Count         *int32          `json:"count"`
	PrimaryZoneID *int64          `json:"primary_zone_id"`
	Meta          json.RawMessage `json:"meta"`
}

// CreateGroup — POST /farms/{id}/animal-groups
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body groupCreateReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	label := strings.TrimSpace(body.Label)
	if label == "" {
		httputil.WriteError(w, http.StatusBadRequest, "label required")
		return
	}
	if body.Count != nil && *body.Count < 0 {
		httputil.WriteError(w, http.StatusBadRequest, "count must be >= 0")
		return
	}
	if body.PrimaryZoneID != nil {
		if err := h.assertZoneInFarm(r, *body.PrimaryZoneID, farmID); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	meta := validMetaOrNil(body.Meta, w)
	if meta == nil && len(body.Meta) > 0 {
		return // error already written
	}
	row, err := h.q.CreateAnimalGroup(r.Context(), db.CreateAnimalGroupParams{
		FarmID:        farmID,
		Label:         label,
		Species:       body.Species,
		Count:         body.Count,
		PrimaryZoneID: body.PrimaryZoneID,
		Meta:          meta,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// UpdateGroup — PUT /animal-groups/{id}
func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid animal group id")
		return
	}
	existing, err := h.q.GetAnimalGroupByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "animal group not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	var body groupCreateReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	label := strings.TrimSpace(body.Label)
	if label == "" {
		httputil.WriteError(w, http.StatusBadRequest, "label required")
		return
	}
	if body.Count != nil && *body.Count < 0 {
		httputil.WriteError(w, http.StatusBadRequest, "count must be >= 0")
		return
	}
	if body.PrimaryZoneID != nil {
		if err := h.assertZoneInFarm(r, *body.PrimaryZoneID, existing.FarmID); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	meta := validMetaOrNil(body.Meta, w)
	if meta == nil && len(body.Meta) > 0 {
		return
	}
	row, err := h.q.UpdateAnimalGroup(r.Context(), db.UpdateAnimalGroupParams{
		ID:            id,
		Label:         label,
		Species:       body.Species,
		Count:         body.Count,
		PrimaryZoneID: body.PrimaryZoneID,
		Meta:          meta,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// ArchiveGroup — PATCH /animal-groups/{id}/archive
// Archiving preserves the group + lifecycle history but marks active=false.
// Distinct from soft-delete, which is for mistake-entry cleanup.
func (h *Handler) ArchiveGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid animal group id")
		return
	}
	existing, err := h.q.GetAnimalGroupByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "animal group not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	var body struct {
		Reason *string `json:"archived_reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body) // optional body
	row, err := h.q.ArchiveAnimalGroup(r.Context(), db.ArchiveAnimalGroupParams{
		ID:             id,
		ArchivedReason: body.Reason,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// DeleteGroup — DELETE /animal-groups/{id}
func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid animal group id")
		return
	}
	existing, err := h.q.GetAnimalGroupByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "animal group not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	if err := h.q.SoftDeleteAnimalGroup(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── lifecycle events ───────────────────────────────────────────────────────

// ListLifecycle — GET /animal-groups/{id}/lifecycle-events
func (h *Handler) ListLifecycle(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid animal group id")
		return
	}
	group, err := h.q.GetAnimalGroupByID(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "animal group not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, group.FarmID) {
		return
	}
	rows, err := h.q.ListLifecycleEventsByGroup(r.Context(), groupID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nanimalsAnimalLifecycleEvent{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// CreateLifecycle — POST /animal-groups/{id}/lifecycle-events
func (h *Handler) CreateLifecycle(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid animal group id")
		return
	}
	group, err := h.q.GetAnimalGroupByID(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "animal group not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, group.FarmID) {
		return
	}
	var body struct {
		EventType     string          `json:"event_type"`
		EventTime     *time.Time      `json:"event_time"`
		DeltaCount    *int32          `json:"delta_count"`
		Notes         *string         `json:"notes"`
		RelatedTaskID *int64          `json:"related_task_id"`
		Meta          json.RawMessage `json:"meta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	evType := strings.TrimSpace(body.EventType)
	if evType == "" {
		httputil.WriteError(w, http.StatusBadRequest, "event_type required")
		return
	}
	if body.RelatedTaskID != nil {
		task, err := h.q.GetTaskByID(r.Context(), *body.RelatedTaskID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "related_task_id not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if task.FarmID != group.FarmID {
			httputil.WriteError(w, http.StatusBadRequest, "related_task_id does not belong to this farm")
			return
		}
	}
	var eventTime pgtype.Timestamptz
	if body.EventTime != nil {
		eventTime = pgtype.Timestamptz{Time: *body.EventTime, Valid: true}
	}
	var recordedBy pgtype.UUID
	if uid, ok := authctx.UserID(r.Context()); ok {
		recordedBy = pgtype.UUID{Bytes: uid, Valid: true}
	}
	meta := validMetaOrNil(body.Meta, w)
	if meta == nil && len(body.Meta) > 0 {
		return
	}
	row, err := h.q.CreateLifecycleEvent(r.Context(), db.CreateLifecycleEventParams{
		FarmID:        group.FarmID,
		AnimalGroupID: groupID,
		EventType:     evType,
		EventTime:     eventTime,
		DeltaCount:    body.DeltaCount,
		Notes:         body.Notes,
		RecordedBy:    recordedBy,
		RelatedTaskID: body.RelatedTaskID,
		Meta:          meta,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// DeleteLifecycle — DELETE /lifecycle-events/{id}. Lifecycle events
// should usually be corrected by appending a compensating event, not
// deleted, but admins can hard-delete for mistake cleanup. Requires
// operate role.
func (h *Handler) DeleteLifecycle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid lifecycle event id")
		return
	}
	existing, err := h.q.GetLifecycleEventByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "lifecycle event not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, existing.FarmID) {
		return
	}
	if err := h.q.DeleteLifecycleEvent(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── helpers ────────────────────────────────────────────────────────────────

// validMetaOrNil returns the meta bytes when caller supplied a body, or
// nil when empty. On invalid JSON it writes a 400 and returns nil; the
// caller distinguishes the two by checking len(body.Meta) against 0.
func validMetaOrNil(raw json.RawMessage, w http.ResponseWriter) []byte {
	if len(raw) == 0 {
		return nil
	}
	if !json.Valid(raw) {
		httputil.WriteError(w, http.StatusBadRequest, "meta must be valid JSON")
		return nil
	}
	return raw
}

func (h *Handler) assertZoneInFarm(r *http.Request, zoneID, farmID int64) error {
	z, err := h.q.GetZoneByID(r.Context(), zoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("primary_zone_id not found")
		}
		return err
	}
	if z.FarmID != farmID {
		return errors.New("primary_zone_id does not belong to this farm")
	}
	return nil
}
