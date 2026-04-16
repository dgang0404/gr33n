package commonscatalog

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

const (
	defaultListLimit = 50
	maxListLimit     = 100
)

type Handler struct {
	q db.Querier
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

// List — GET /commons/catalog
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := defaultListLimit
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
			if limit > maxListLimit {
				limit = maxListLimit
			}
		}
	}
	offset := 0
	if s := strings.TrimSpace(r.URL.Query().Get("offset")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			offset = n
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.q.ListPublishedCommonsCatalogEntries(ctx, db.ListPublishedCommonsCatalogEntriesParams{
		Column1: q,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list catalog")
		return
	}
	if rows == nil {
		rows = []db.ListPublishedCommonsCatalogEntriesRow{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// GetBySlug — GET /commons/catalog/{slug}
func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	if slug == "" {
		httputil.WriteError(w, http.StatusBadRequest, "invalid slug")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	row, err := h.q.GetPublishedCommonsCatalogEntryBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "catalog entry not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load catalog entry")
		return
	}
	var body any
	if len(row.Body) > 0 {
		_ = json.Unmarshal(row.Body, &body)
	}
	out := map[string]any{
		"id":                  row.ID,
		"slug":                row.Slug,
		"title":               row.Title,
		"summary":             row.Summary,
		"body":                body,
		"contributor_display": row.ContributorDisplay,
		"license_spdx":        row.LicenseSpdx,
		"tags":                row.Tags,
		"sort_order":          row.SortOrder,
		"created_at":          row.CreatedAt,
		"updated_at":          row.UpdatedAt,
	}
	if row.ContributorUri != nil {
		out["contributor_uri"] = *row.ContributorUri
	}
	if row.LicenseNotes != nil {
		out["license_notes"] = *row.LicenseNotes
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

// ListFarmImports — GET /farms/{id}/commons/catalog-imports
func (h *Handler) ListFarmImports(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farmID) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.q.ListFarmCommonsCatalogImports(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list imports")
		return
	}
	if rows == nil {
		rows = []db.ListFarmCommonsCatalogImportsRow{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Import — POST /farms/{id}/commons/catalog-imports
func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var body struct {
		Slug string  `json:"slug"`
		Note *string `json:"note"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	slug := strings.TrimSpace(body.Slug)
	if slug == "" {
		httputil.WriteError(w, http.StatusBadRequest, "slug is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	entry, err := h.q.GetPublishedCommonsCatalogEntryBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "catalog entry not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load catalog entry")
		return
	}

	var notePtr *string
	if body.Note != nil && strings.TrimSpace(*body.Note) != "" {
		n := strings.TrimSpace(*body.Note)
		notePtr = &n
	}

	row, err := h.q.UpsertFarmCommonsCatalogImport(ctx, db.UpsertFarmCommonsCatalogImportParams{
		FarmID:         farmID,
		CatalogEntryID: entry.ID,
		ImportedBy:     uid,
		Note:           notePtr,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record import")
		return
	}

	var payload any
	if len(entry.Body) > 0 {
		_ = json.Unmarshal(entry.Body, &payload)
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"import": map[string]any{
			"id":               row.ID,
			"farm_id":          row.FarmID,
			"catalog_entry_id": row.CatalogEntryID,
			"imported_at":      row.ImportedAt,
			"note":             row.Note,
		},
		"catalog_entry": map[string]any{
			"id":                  entry.ID,
			"slug":                entry.Slug,
			"title":               entry.Title,
			"summary":             entry.Summary,
			"body":                payload,
			"contributor_display": entry.ContributorDisplay,
			"license_spdx":        entry.LicenseSpdx,
			"tags":                entry.Tags,
		},
	})
}
