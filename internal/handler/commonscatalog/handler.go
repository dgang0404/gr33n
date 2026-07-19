package commonscatalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	catalogpack "gr33n-api/internal/commonscatalog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

const (
	defaultListLimit = 50
	maxListLimit     = 100
	maxPublishBody   = 512 << 10 // 512 KiB
)

type Handler struct {
	q db.Querier
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{q: db.New(pool)}
}

func catalogEntryJSON(row db.Gr33ncoreCommonsCatalogEntry) map[string]any {
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
		"published":           row.Published,
		"created_at":          row.CreatedAt,
		"updated_at":          row.UpdatedAt,
	}
	if row.ContributorUri != nil {
		out["contributor_uri"] = *row.ContributorUri
	}
	if row.LicenseNotes != nil {
		out["license_notes"] = *row.LicenseNotes
	}
	if row.PublishedByUserID.Valid {
		out["published_by_user_id"] = uuid.UUID(row.PublishedByUserID.Bytes).String()
	}
	if row.SourceFarmID != nil {
		out["source_farm_id"] = *row.SourceFarmID
	}
	return out
}

// List — GET /commons/catalog
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit, offset := httputil.ParseLimitOffset(r, defaultListLimit, maxListLimit)

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
	out := map[string]any{
		"id":                  row.ID,
		"slug":                row.Slug,
		"title":               row.Title,
		"summary":             row.Summary,
		"body":                jsonRawToAny(row.Body),
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
	if row.PublishedByUserID.Valid {
		out["published_by_user_id"] = uuid.UUID(row.PublishedByUserID.Bytes).String()
	}
	if row.SourceFarmID != nil {
		out["source_farm_id"] = *row.SourceFarmID
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

// Publish — POST /commons/catalog
func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	uid, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req struct {
		Slug                string          `json:"slug"`
		Title               string          `json:"title"`
		Summary             string          `json:"summary"`
		Body                json.RawMessage `json:"body"`
		ContributorDisplay  string          `json:"contributor_display"`
		ContributorURI      *string         `json:"contributor_uri"`
		LicenseSPDX         string          `json:"license_spdx"`
		LicenseNotes        *string         `json:"license_notes"`
		Tags                []string        `json:"tags"`
		Published           *bool           `json:"published"`
		SourceFarmID        *int64          `json:"source_farm_id"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, maxPublishBody)).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	slug, err := catalogpack.NormalizeSlug(req.Slug)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		httputil.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(req.Body) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "body is required")
		return
	}

	packBody, err := catalogpack.ParsePackBody(req.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := catalogpack.ValidatePublishBody(packBody); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if packBody.Kind == catalogpack.KindFertigationRecipePack {
		crops, err := h.q.ListCropCatalogEntries(r.Context())
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to load crop catalog")
			return
		}
		if err := catalogpack.ValidateRecipeCropKeys(packBody.Programs, crops); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	if req.SourceFarmID != nil && *req.SourceFarmID > 0 {
		if !farmauthz.RequireFarmAdmin(w, r, h.q, *req.SourceFarmID) {
			return
		}
	}

	published := true
	if req.Published != nil {
		published = *req.Published
	}
	license := strings.TrimSpace(req.LicenseSPDX)
	if license == "" {
		license = "CC-BY-4.0"
	}
	contributor := strings.TrimSpace(req.ContributorDisplay)
	if contributor == "" {
		contributor = "gr33n operator"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var pubUUID pgtype.UUID
	if err := pubUUID.Scan(uid.String()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "invalid user id")
		return
	}

	row, err := h.q.InsertCommonsCatalogEntry(ctx, db.InsertCommonsCatalogEntryParams{
		Slug:               slug,
		Title:              title,
		Summary:            strings.TrimSpace(req.Summary),
		Body:               req.Body,
		ContributorDisplay: contributor,
		ContributorUri:     req.ContributorURI,
		LicenseSpdx:        license,
		LicenseNotes:       req.LicenseNotes,
		Tags:               req.Tags,
		Published:          published,
		SortOrder:          100,
		PublishedByUserID:  pubUUID,
		SourceFarmID:       req.SourceFarmID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			httputil.WriteError(w, http.StatusConflict, "catalog slug already exists")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to publish catalog entry")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, catalogEntryJSON(row))
}

// ExportRecipePack — POST /farms/{id}/commons/catalog-export/recipe-pack
func (h *Handler) ExportRecipePack(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Slug               string  `json:"slug"`
		Title              string  `json:"title"`
		Summary            string  `json:"summary"`
		ReadmeMD           string  `json:"readme_md"`
		ContributorDisplay string  `json:"contributor_display"`
		LicenseSPDX        string  `json:"license_spdx"`
		Tags               []string `json:"tags"`
		Publish            *bool   `json:"publish"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	slug, err := catalogpack.NormalizeSlug(req.Slug)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		httputil.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	programs, err := h.q.ListProgramsByFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list programs")
		return
	}
	if len(programs) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "farm has no fertigation programs to export")
		return
	}

	readme := strings.TrimSpace(req.ReadmeMD)
	if readme == "" {
		readme = fmt.Sprintf("# %s\n\nExported from farm %d. Programs import as **inactive** — review before enabling automation.\n", title, farmID)
	}
	bodyBytes, err := catalogpack.BuildRecipePackBody(programs, readme)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to build recipe pack")
		return
	}

	publish := true
	if req.Publish != nil {
		publish = *req.Publish
	}
	license := strings.TrimSpace(req.LicenseSPDX)
	if license == "" {
		license = "CC-BY-4.0"
	}
	contributor := strings.TrimSpace(req.ContributorDisplay)
	if contributor == "" {
		contributor = "gr33n operator"
	}

	var pubUUID pgtype.UUID
	if err := pubUUID.Scan(uid.String()); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "invalid user id")
		return
	}
	farmCopy := farmID
	row, err := h.q.InsertCommonsCatalogEntry(ctx, db.InsertCommonsCatalogEntryParams{
		Slug:               slug,
		Title:              title,
		Summary:            strings.TrimSpace(req.Summary),
		Body:               bodyBytes,
		ContributorDisplay: contributor,
		LicenseSpdx:        license,
		Tags:               req.Tags,
		Published:          publish,
		SortOrder:          100,
		PublishedByUserID:  pubUUID,
		SourceFarmID:       &farmCopy,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			httputil.WriteError(w, http.StatusConflict, "catalog slug already exists")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to publish exported pack")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"catalog_entry": catalogEntryJSON(row),
		"programs_exported": len(programs),
		"message": fmt.Sprintf("Published recipe pack with %d programs. Other farms can import from Help → Catalog.", len(programs)),
	})
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

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
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

	applyResult, applyErr := catalogpack.ApplyPack(ctx, h.q, farmID, entry.Body)
	if applyErr != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"import": map[string]any{
				"id":               row.ID,
				"farm_id":          row.FarmID,
				"catalog_entry_id": row.CatalogEntryID,
				"imported_at":      row.ImportedAt,
				"note":             row.Note,
			},
			"catalog_entry": map[string]any{
				"id":    entry.ID,
				"slug":  entry.Slug,
				"title": entry.Title,
			},
			"apply": applyResult,
			"error": applyErr.Error(),
		})
		return
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
			"body":                jsonRawToAny(entry.Body),
			"contributor_display": entry.ContributorDisplay,
			"license_spdx":        entry.LicenseSpdx,
			"tags":                entry.Tags,
		},
		"apply": applyResult,
	})
}

func jsonRawToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	_ = json.Unmarshal(raw, &v)
	return v
}
