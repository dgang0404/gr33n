package commonscropcatalog

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	q db.Querier
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

type catalogListResponse struct {
	CatalogVersion int32                       `json:"catalog_version"`
	Count          int                         `json:"count"`
	Entries        []db.Gr33ncropsCropCatalogEntry `json:"entries"`
	Aliases        map[string]string           `json:"aliases"`
}

type catalogEntryDetail struct {
	db.Gr33ncropsCropCatalogEntry
	Aliases       []string `json:"aliases"`
	CropProfileID *int64   `json:"crop_profile_id,omitempty"`
}

type fieldGuideSummary struct {
	ID             int64   `json:"id"`
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	CropKey        *string `json:"crop_key"`
	GuideKind      string  `json:"guide_kind"`
	Domain         *string `json:"domain"`
	SafetyTier     string  `json:"safety_tier"`
	CatalogVersion int32   `json:"catalog_version"`
	SortOrder      int32   `json:"sort_order"`
}

type fieldGuideDetail struct {
	fieldGuideSummary
	BodyMd string `json:"body_md"`
}

// ListCropCatalog — GET /commons/crop-catalog
func (h *Handler) ListCropCatalog(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	entries, err := h.q.ListCropCatalogEntries(ctx)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list crop catalog")
		return
	}
	if supportedOnly := parseBoolQuery(r, "supported"); supportedOnly != nil {
		filtered := entries[:0]
		for _, e := range entries {
			if e.Supported == *supportedOnly {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	aliasRows, err := h.q.ListCropCatalogAliases(ctx)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list crop aliases")
		return
	}
	aliases := make(map[string]string, len(aliasRows))
	version := int32(1)
	for _, e := range entries {
		if e.CatalogVersion > version {
			version = e.CatalogVersion
		}
	}
	for _, a := range aliasRows {
		aliases[a.Alias] = a.CropKey
		if version == 1 {
			// aliases don't carry version; entries loop above sets version
		}
	}

	httputil.WriteJSON(w, http.StatusOK, catalogListResponse{
		CatalogVersion: version,
		Count:          len(entries),
		Entries:        entries,
		Aliases:        aliases,
	})
}

// GetCropCatalogEntry — GET /commons/crop-catalog/{crop_key}
func (h *Handler) GetCropCatalogEntry(w http.ResponseWriter, r *http.Request) {
	cropKey := strings.ToLower(strings.TrimSpace(r.PathValue("crop_key")))
	if cropKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "invalid crop_key")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	entry, err := h.q.GetCropCatalogEntry(ctx, cropKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "crop catalog entry not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load crop catalog entry")
		return
	}

	aliasRows, err := h.q.ListCropCatalogAliases(ctx)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list crop aliases")
		return
	}
	var cropAliases []string
	for _, a := range aliasRows {
		if a.CropKey == cropKey && a.Alias != cropKey {
			cropAliases = append(cropAliases, a.Alias)
		}
	}

	out := catalogEntryDetail{
		Gr33ncropsCropCatalogEntry: entry,
		Aliases:                    cropAliases,
	}
	if entry.Supported {
		if id, err := h.q.GetBuiltinCropProfileIDByCropKey(ctx, cropKey); err == nil {
			out.CropProfileID = &id
		} else if !errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to resolve crop profile")
			return
		}
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

// ListFieldGuides — GET /commons/agronomy-field-guides
func (h *Handler) ListFieldGuides(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	guides, err := h.q.ListAgronomyFieldGuides(ctx)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list field guides")
		return
	}
	cropKey := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("crop_key")))
	kind := strings.TrimSpace(r.URL.Query().Get("guide_kind"))

	var out []fieldGuideSummary
	for _, g := range guides {
		if cropKey != "" {
			if g.CropKey == nil || strings.ToLower(*g.CropKey) != cropKey {
				continue
			}
		}
		if kind != "" && !strings.EqualFold(g.GuideKind, kind) {
			continue
		}
		out = append(out, toFieldGuideSummary(g))
	}
	if out == nil {
		out = []fieldGuideSummary{}
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

// GetFieldGuide — GET /commons/agronomy-field-guides/{slug}
func (h *Handler) GetFieldGuide(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		httputil.WriteError(w, http.StatusBadRequest, "invalid slug")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	guide, err := h.q.GetPublishedAgronomyFieldGuideBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "field guide not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load field guide")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, fieldGuideDetail{
		fieldGuideSummary: toFieldGuideSummary(guide),
		BodyMd:            guide.BodyMd,
	})
}

func toFieldGuideSummary(g db.Gr33ncropsAgronomyFieldGuide) fieldGuideSummary {
	return fieldGuideSummary{
		ID:             g.ID,
		Slug:           g.Slug,
		Title:          g.Title,
		CropKey:        g.CropKey,
		GuideKind:      g.GuideKind,
		Domain:         g.Domain,
		SafetyTier:     g.SafetyTier,
		CatalogVersion: g.CatalogVersion,
		SortOrder:      g.SortOrder,
	}
}

func parseBoolQuery(r *http.Request, key string) *bool {
	s := strings.TrimSpace(r.URL.Query().Get(key))
	if s == "" {
		return nil
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return nil
	}
	return &v
}
