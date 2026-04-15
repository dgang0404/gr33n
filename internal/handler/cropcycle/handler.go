package cropcycle

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

type Handler struct{ q *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

func numericFromFloat64(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(v, 'f', -1, 64))
	return n, err
}

func parseGrowthStage(s string) db.NullGr33nfertigationGrowthStageEnum {
	if strings.TrimSpace(s) == "" {
		return db.NullGr33nfertigationGrowthStageEnum{
			Gr33nfertigationGrowthStageEnum: db.Gr33nfertigationGrowthStageEnumSeedling,
			Valid:                           true,
		}
	}
	return db.NullGr33nfertigationGrowthStageEnum{
		Gr33nfertigationGrowthStageEnum: db.Gr33nfertigationGrowthStageEnum(strings.TrimSpace(s)),
		Valid:                           true,
	}
}

func parseDate(s string) (pgtype.Date, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Date{}, errors.New("empty date")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// List — GET /farms/{id}/crop-cycles
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	rows, err := h.q.ListCropCyclesByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33nfertigationCropCycle{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Get — GET /crop-cycles/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	row, err := h.q.GetCropCycleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Create — POST /farms/{id}/crop-cycles
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	var body struct {
		ZoneID           int64   `json:"zone_id"`
		Name             string  `json:"name"`
		StrainOrVariety  *string `json:"strain_or_variety"`
		CurrentStage     string  `json:"current_stage"`
		IsActive         *bool   `json:"is_active"`
		StartedAt        string  `json:"started_at"`
		CycleNotes       *string `json:"cycle_notes"`
		PrimaryProgramID *int64  `json:"primary_program_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || body.ZoneID == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "name and zone_id required")
		return
	}
	z, err := h.q.GetZoneByID(r.Context(), body.ZoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "zone not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if z.FarmID != farmID {
		httputil.WriteError(w, http.StatusBadRequest, "zone does not belong to this farm")
		return
	}
	started, err := parseDate(body.StartedAt)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid started_at (use YYYY-MM-DD)")
		return
	}
	active := true
	if body.IsActive != nil {
		active = *body.IsActive
	}
	row, err := h.q.CreateCropCycle(r.Context(), db.CreateCropCycleParams{
		FarmID:           farmID,
		ZoneID:           body.ZoneID,
		Name:             name,
		StrainOrVariety:  body.StrainOrVariety,
		CurrentStage:     parseGrowthStage(body.CurrentStage),
		IsActive:         active,
		StartedAt:        started,
		CycleNotes:       body.CycleNotes,
		PrimaryProgramID: body.PrimaryProgramID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httputil.WriteError(w, http.StatusConflict, "only one active crop cycle per zone is allowed")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// Update — PUT /crop-cycles/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	var body struct {
		Name             string   `json:"name"`
		StrainOrVariety  *string  `json:"strain_or_variety"`
		ZoneID           int64    `json:"zone_id"`
		IsActive         bool     `json:"is_active"`
		CycleNotes       *string  `json:"cycle_notes"`
		HarvestedAt      *string  `json:"harvested_at"`
		YieldGrams       *float64 `json:"yield_grams"`
		YieldNotes       *string  `json:"yield_notes"`
		PrimaryProgramID *int64   `json:"primary_program_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || body.ZoneID == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "name and zone_id required")
		return
	}
	existing, err := h.q.GetCropCycleByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	z, err := h.q.GetZoneByID(r.Context(), body.ZoneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusBadRequest, "zone not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if z.FarmID != existing.FarmID {
		httputil.WriteError(w, http.StatusBadRequest, "zone does not belong to this farm")
		return
	}
	harvested := existing.HarvestedAt
	if body.HarvestedAt != nil {
		if strings.TrimSpace(*body.HarvestedAt) == "" {
			harvested = pgtype.Date{}
		} else {
			harvested, err = parseDate(*body.HarvestedAt)
			if err != nil {
				httputil.WriteError(w, http.StatusBadRequest, "invalid harvested_at")
				return
			}
		}
	}
	yield := existing.YieldGrams
	if body.YieldGrams != nil {
		yield, err = numericFromFloat64(*body.YieldGrams)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid yield_grams")
			return
		}
	}
	row, err := h.q.UpdateCropCycle(r.Context(), db.UpdateCropCycleParams{
		ID:               id,
		Name:             name,
		StrainOrVariety:  body.StrainOrVariety,
		ZoneID:           body.ZoneID,
		IsActive:         body.IsActive,
		CycleNotes:       body.CycleNotes,
		HarvestedAt:      harvested,
		YieldGrams:       yield,
		YieldNotes:       body.YieldNotes,
		PrimaryProgramID: body.PrimaryProgramID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			httputil.WriteError(w, http.StatusConflict, "only one active crop cycle per zone is allowed")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// UpdateStage — PATCH /crop-cycles/{id}/stage
func (h *Handler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	var body struct {
		CurrentStage string `json:"current_stage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(body.CurrentStage) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "current_stage required")
		return
	}
	if _, err := h.q.GetCropCycleByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	row, err := h.q.UpdateCropCycleStage(r.Context(), db.UpdateCropCycleStageParams{
		ID:           id,
		CurrentStage: parseGrowthStage(body.CurrentStage),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Delete — DELETE /crop-cycles/{id} (soft: is_active = false)
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop cycle id")
		return
	}
	if _, err := h.q.GetCropCycleByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop cycle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.q.SoftDeleteCropCycle(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
