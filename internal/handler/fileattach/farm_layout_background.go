package fileattach

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/farmlayout"
	"gr33n-api/internal/fileattachutil"
	"gr33n-api/internal/filestorage"
	"gr33n-api/internal/httputil"
)

const maxFarmLayoutBackgroundUpload = 8 << 20 // 8 MiB

var farmLayoutBackgroundMimeOK = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

// UploadFarmLayoutBackground — POST /farms/{id}/layout-background (multipart: file)
func (h *Handler) UploadFarmLayoutBackground(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	farm, err := h.q.GetFarmByID(r.Context(), farmID)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "farm not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farm.ID) {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxFarmLayoutBackgroundUpload+512*1024)
	if err := r.ParseMultipartForm(maxFarmLayoutBackgroundUpload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "file field required")
		return
	}
	defer file.Close()

	detected, body, err := fileattachutil.SniffAndValidate(file, farmLayoutBackgroundMimeOK)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	mime := detected

	ext := filestorage.ExtForMime(mime)
	key := "farm-" + strconv.FormatInt(farmID, 10) + "/layout-background/" + uuid.New().String() + ext
	n, err := h.store.Put(r.Context(), key, body, maxFarmLayoutBackgroundUpload)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var uid pgtype.UUID
	if u, ok := authctx.UserID(r.Context()); ok {
		uid = pgtype.UUID{Bytes: u, Valid: true}
	}
	sz := n
	att, err := h.q.CreateFileAttachment(r.Context(), db.CreateFileAttachmentParams{
		FarmID:              farmID,
		RelatedModuleSchema: "gr33ncore",
		RelatedTableName:    "farms",
		RelatedRecordID:     strconv.FormatInt(farmID, 10),
		FileName:            hdr.Filename,
		FileType:            "farm_layout_background",
		FileSizeBytes:       &sz,
		StoragePath:         key,
		MimeType:            &mime,
		UploadedByUserID:    uid,
	})
	if err != nil {
		_ = h.store.Delete(r.Context(), key)
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	meta, extra, err := farmlayout.ParseMeta(farm.MetaData)
	if err != nil {
		_ = h.store.Delete(r.Context(), key)
		_ = fileattachutil.DeleteFarmLayoutBackgroundIfUnreferenced(r.Context(), h.pool, h.store, att.ID)
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	prevID := meta.LayoutBackgroundAttachmentID
	if err := farmlayout.SetLayoutBackgroundID(&meta, att.ID); err != nil {
		_ = h.store.Delete(r.Context(), key)
		_ = fileattachutil.DeleteFarmLayoutBackgroundIfUnreferenced(r.Context(), h.pool, h.store, att.ID)
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	metaBytes, err := farmlayout.MarshalMeta(meta, extra)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := h.q.SetFarmLayoutBackgroundAttachment(r.Context(), db.SetFarmLayoutBackgroundAttachmentParams{
		ID:       farmID,
		MetaData: metaBytes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if prevID != nil && *prevID != att.ID {
		_ = fileattachutil.DeleteFarmLayoutBackgroundIfUnreferenced(r.Context(), h.pool, h.store, *prevID)
	}

	h.logFarmLayoutBackgroundAudit(r, farmID, att.ID, "farm_layout_background_uploaded", nil)
	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"file_attachment": att,
		"farm":            updated,
		"attachment_id":   att.ID,
		"content_url":     "/file-attachments/" + strconv.FormatInt(att.ID, 10) + "/content",
	})
}

// GetFarmLayoutBackground — GET /farms/{id}/layout-background
func (h *Handler) GetFarmLayoutBackground(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	farm, err := h.q.GetFarmByID(r.Context(), farmID)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "farm not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, farm.ID) {
		return
	}
	meta, _, err := farmlayout.ParseMeta(farm.MetaData)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if meta.LayoutBackgroundAttachmentID == nil || *meta.LayoutBackgroundAttachmentID < 1 {
		httputil.WriteError(w, http.StatusNotFound, "no layout background set")
		return
	}
	attID := *meta.LayoutBackgroundAttachmentID
	att, err := h.q.GetFileAttachmentByID(r.Context(), attID)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "layout background not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"farm_id":       farmID,
		"attachment_id": attID,
		"file_name":     att.FileName,
		"content_url":   "/file-attachments/" + strconv.FormatInt(attID, 10) + "/content",
		"download_url":  "/file-attachments/" + strconv.FormatInt(attID, 10) + "/download",
	})
}

// DeleteFarmLayoutBackground — DELETE /farms/{id}/layout-background
func (h *Handler) DeleteFarmLayoutBackground(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || farmID < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	farm, err := h.q.GetFarmByID(r.Context(), farmID)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "farm not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, farm.ID) {
		return
	}
	meta, extra, err := farmlayout.ParseMeta(farm.MetaData)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if meta.LayoutBackgroundAttachmentID == nil {
		httputil.WriteError(w, http.StatusNotFound, "no layout background set")
		return
	}
	attID := *meta.LayoutBackgroundAttachmentID
	farmlayout.ClearLayoutBackgroundID(&meta)
	metaBytes, err := farmlayout.MarshalMeta(meta, extra)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := h.q.ClearFarmLayoutBackgroundAttachment(r.Context(), db.ClearFarmLayoutBackgroundAttachmentParams{
		ID:       farmID,
		MetaData: metaBytes,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := fileattachutil.DeleteFarmLayoutBackgroundIfUnreferenced(r.Context(), h.pool, h.store, attID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.logFarmLayoutBackgroundAudit(r, farmID, attID, "farm_layout_background_deleted", nil)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"farm": updated})
}

func (h *Handler) logFarmLayoutBackgroundAudit(r *http.Request, farmID, attachmentID int64, kind string, extra map[string]any) {
	details := map[string]any{"kind": kind, "file_attachment_id": attachmentID}
	for k, v := range extra {
		details[k] = v
	}
	mod := "gr33ncore"
	tbl := "farms"
	rid := strconv.FormatInt(farmID, 10)
	auditlog.Submit(r.Context(), h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumUpdateRecord,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details:        details,
	})
}
