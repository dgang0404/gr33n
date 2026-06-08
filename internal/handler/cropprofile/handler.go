package cropprofile

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

type Handler struct{ q *db.Queries }

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

type profileWithStages struct {
	db.Gr33ncropsCropProfile
	Stages []db.Gr33ncropsCropProfileStage `json:"stages"`
}

// List — GET /farms/{id}/crop-profiles
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListCropProfilesForFarm(r.Context(), &farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncropsCropProfile{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Get — GET /crop-profiles/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop profile id")
		return
	}
	profile, err := h.q.GetCropProfile(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop profile not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if profile.FarmID != nil {
		if !farmauthz.RequireFarmMember(w, r, h.q, *profile.FarmID) {
			return
		}
	}
	stages, err := h.q.ListCropProfileStages(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if stages == nil {
		stages = []db.Gr33ncropsCropProfileStage{}
	}
	httputil.WriteJSON(w, http.StatusOK, profileWithStages{Gr33ncropsCropProfile: profile, Stages: stages})
}

// Clone — POST /crop-profiles/{id}/clone?farm_id=
func (h *Handler) Clone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop profile id")
		return
	}
	farmID, err := strconv.ParseInt(r.URL.Query().Get("farm_id"), 10, 64)
	if err != nil || farmID <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "farm_id query required")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	source, err := h.q.GetCropProfile(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop profile not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !source.IsBuiltin {
		httputil.WriteError(w, http.StatusBadRequest, "only built-in profiles can be cloned")
		return
	}
	stages, err := h.q.ListCropProfileStages(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	farmIDPtr := farmID
	cropKey := fmt.Sprintf("%s-farm-%d", source.CropKey, farmID)
	displayName := source.DisplayName + " (my copy)"
	sourceNote := ""
	if source.Source != nil {
		sourceNote = *source.Source + " — "
	}
	sourceNote += "cloned from built-in"
	created, err := h.q.CreateCropProfile(r.Context(), db.CreateCropProfileParams{
		FarmID:      &farmIDPtr,
		CropKey:     cropKey,
		DisplayName: displayName,
		Category:    source.Category,
		Source:      &sourceNote,
		Version:     1,
		IsBuiltin:   false,
		Meta:        source.Meta,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, st := range stages {
		if _, err := h.q.CreateCropProfileStage(r.Context(), db.CreateCropProfileStageParams{
			CropProfileID:  created.ID,
			Stage:          st.Stage,
			EcMin:          st.EcMin,
			EcTarget:       st.EcTarget,
			EcMax:          st.EcMax,
			PhMin:          st.PhMin,
			PhMax:          st.PhMax,
			VpdMinKpa:      st.VpdMinKpa,
			VpdMaxKpa:      st.VpdMaxKpa,
			TempMinC:       st.TempMinC,
			TempMaxC:       st.TempMaxC,
			RhMinPct:       st.RhMinPct,
			RhMaxPct:       st.RhMaxPct,
			DliTarget:      st.DliTarget,
			PhotoperiodHrs: st.PhotoperiodHrs,
			Notes:          st.Notes,
		}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	outStages, _ := h.q.ListCropProfileStages(r.Context(), created.ID)
	httputil.WriteJSON(w, http.StatusCreated, profileWithStages{Gr33ncropsCropProfile: created, Stages: outStages})
}

// Export — GET /crop-profiles/{id}/export
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop profile id")
		return
	}
	profile, err := h.q.GetCropProfile(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop profile not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if profile.FarmID != nil {
		if !farmauthz.RequireFarmMember(w, r, h.q, *profile.FarmID) {
			return
		}
	}
	stages, err := h.q.ListCropProfileStages(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", profile.CropKey+".json"))
	httputil.WriteJSON(w, http.StatusOK, profileWithStages{Gr33ncropsCropProfile: profile, Stages: stages})
}

// Import — POST /farms/{id}/crop-profiles/import
func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farmID) {
		return
	}
	var body struct {
		CropKey     string                          `json:"crop_key"`
		DisplayName string                          `json:"display_name"`
		Category    *string                         `json:"category"`
		Source      *string                         `json:"source"`
		Stages      []db.Gr33ncropsCropProfileStage `json:"stages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	key := strings.TrimSpace(body.CropKey)
	name := strings.TrimSpace(body.DisplayName)
	if key == "" || name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "crop_key and display_name required")
		return
	}
	farmIDPtr := farmID
	src := "imported JSON"
	if body.Source != nil && strings.TrimSpace(*body.Source) != "" {
		src = strings.TrimSpace(*body.Source)
	}
	created, err := h.q.CreateCropProfile(r.Context(), db.CreateCropProfileParams{
		FarmID:      &farmIDPtr,
		CropKey:     key,
		DisplayName: name,
		Category:    body.Category,
		Source:      &src,
		Version:     1,
		IsBuiltin:   false,
		Meta:        []byte("{}"),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, st := range body.Stages {
		if _, err := h.q.CreateCropProfileStage(r.Context(), db.CreateCropProfileStageParams{
			CropProfileID:  created.ID,
			Stage:          st.Stage,
			EcMin:          st.EcMin,
			EcTarget:       st.EcTarget,
			EcMax:          st.EcMax,
			PhMin:          st.PhMin,
			PhMax:          st.PhMax,
			VpdMinKpa:      st.VpdMinKpa,
			VpdMaxKpa:      st.VpdMaxKpa,
			TempMinC:       st.TempMinC,
			TempMaxC:       st.TempMaxC,
			RhMinPct:       st.RhMinPct,
			RhMaxPct:       st.RhMaxPct,
			DliTarget:      st.DliTarget,
			PhotoperiodHrs: st.PhotoperiodHrs,
			Notes:          st.Notes,
		}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	outStages, _ := h.q.ListCropProfileStages(r.Context(), created.ID)
	httputil.WriteJSON(w, http.StatusCreated, profileWithStages{Gr33ncropsCropProfile: created, Stages: outStages})
}
