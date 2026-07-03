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
	"gr33n-api/internal/fileattachutil"
	"gr33n-api/internal/filestorage"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/zonephotos"
)

const maxZonePhotoUpload = 8 << 20 // 8 MiB

var zonePhotoMimeOK = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

// UploadZonePhoto — POST /zones/{id}/photos (multipart: file, optional description)
func (h *Handler) UploadZonePhoto(w http.ResponseWriter, r *http.Request) {
	zoneID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || zoneID < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid zone id")
		return
	}
	z, err := h.q.GetZoneByID(r.Context(), zoneID)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "zone not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, z.FarmID) {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxZonePhotoUpload+512*1024)
	if err := r.ParseMultipartForm(maxZonePhotoUpload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "file field required")
		return
	}
	defer file.Close()
	_ = hdr

	detected, body, err := fileattachutil.SniffAndValidate(file, zonePhotoMimeOK)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	mime := detected

	ext := filestorage.ExtForMime(mime)
	key := "farm-" + strconv.FormatInt(z.FarmID, 10) + "/zones/" + strconv.FormatInt(zoneID, 10) + "/" + uuid.New().String() + ext
	n, err := h.store.Put(r.Context(), key, body, maxZonePhotoUpload)
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
		FarmID:              z.FarmID,
		RelatedModuleSchema: "gr33ncore",
		RelatedTableName:    "zones",
		RelatedRecordID:     strconv.FormatInt(zoneID, 10),
		FileName:            hdr.Filename,
		FileType:            "zone_photo",
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

	meta, extra, err := zonephotos.ParseMeta(z.MetaData)
	if err != nil {
		_ = h.store.Delete(r.Context(), key)
		_ = fileattachutil.DeleteZonePhotoIfUnreferenced(r.Context(), h.pool, h.store, att.ID)
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := zonephotos.AppendPhotoID(&meta, att.ID); err != nil {
		_ = h.store.Delete(r.Context(), key)
		_ = fileattachutil.DeleteZonePhotoIfUnreferenced(r.Context(), h.pool, h.store, att.ID)
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	metaBytes, err := zonephotos.MarshalMeta(meta, extra)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, err := h.q.UpdateZone(r.Context(), db.UpdateZoneParams{
		ID:              z.ID,
		Name:            z.Name,
		Description:     z.Description,
		ZoneType:        z.ZoneType,
		AreaSqm:         z.AreaSqm,
		MetaData:        metaBytes,
		UpdatedByUserID: uid,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logZonePhotoAudit(r, z.FarmID, zoneID, att.ID, "zone_photo_uploaded", map[string]any{
		"file_attachment_id": att.ID,
	})
	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"file_attachment": att,
		"zone":            updated,
	})
}

// ListZonePhotos — GET /zones/{id}/photos
func (h *Handler) ListZonePhotos(w http.ResponseWriter, r *http.Request) {
	zoneID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || zoneID < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid zone id")
		return
	}
	z, err := h.q.GetZoneByID(r.Context(), zoneID)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "zone not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmMember(w, r, h.q, z.FarmID) {
		return
	}
	meta, _, err := zonephotos.ParseMeta(z.MetaData)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	photos := make([]map[string]any, 0, len(meta.PhotoAttachmentIDs))
	for _, id := range meta.PhotoAttachmentIDs {
		att, err := h.q.GetFileAttachmentByID(r.Context(), id)
		if err != nil {
			continue
		}
		photos = append(photos, zonePhotoJSON(att))
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"zone_id": zoneID,
		"photos":  photos,
	})
}

// DeleteZonePhoto — DELETE /zones/{id}/photos/{attachment_id}
func (h *Handler) DeleteZonePhoto(w http.ResponseWriter, r *http.Request) {
	zoneID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || zoneID < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid zone id")
		return
	}
	attID, err := strconv.ParseInt(r.PathValue("attachment_id"), 10, 64)
	if err != nil || attID < 1 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid attachment id")
		return
	}
	z, err := h.q.GetZoneByID(r.Context(), zoneID)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "zone not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireFarmOperate(w, r, h.q, z.FarmID) {
		return
	}
	meta, extra, err := zonephotos.ParseMeta(z.MetaData)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !zonephotos.RemovePhotoID(&meta, attID) {
		httputil.WriteError(w, http.StatusNotFound, "photo not linked to this zone")
		return
	}
	metaBytes, err := zonephotos.MarshalMeta(meta, extra)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var uid pgtype.UUID
	if u, ok := authctx.UserID(r.Context()); ok {
		uid = pgtype.UUID{Bytes: u, Valid: true}
	}
	updated, err := h.q.UpdateZone(r.Context(), db.UpdateZoneParams{
		ID:              z.ID,
		Name:            z.Name,
		Description:     z.Description,
		ZoneType:        z.ZoneType,
		AreaSqm:         z.AreaSqm,
		MetaData:        metaBytes,
		UpdatedByUserID: uid,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := fileattachutil.DeleteZonePhotoIfUnreferenced(r.Context(), h.pool, h.store, attID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.logZonePhotoAudit(r, z.FarmID, zoneID, attID, "zone_photo_deleted", nil)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"zone": updated})
}

func zonePhotoJSON(att db.Gr33ncoreFileAttachment) map[string]any {
	out := map[string]any{
		"id":         att.ID,
		"file_name":  att.FileName,
		"file_type":  att.FileType,
		"created_at": att.CreatedAt,
		"content_url": "/file-attachments/" + strconv.FormatInt(att.ID, 10) + "/content",
		"download_url": "/file-attachments/" + strconv.FormatInt(att.ID, 10) + "/download",
	}
	if att.MimeType != nil {
		out["mime_type"] = *att.MimeType
	}
	if att.Description != nil {
		out["description"] = *att.Description
	}
	if att.FileSizeBytes != nil {
		out["file_size_bytes"] = *att.FileSizeBytes
	}
	return out
}

func (h *Handler) logZonePhotoAudit(r *http.Request, farmID, zoneID, attachmentID int64, kind string, extra map[string]any) {
	details := map[string]any{"kind": kind, "zone_id": zoneID, "file_attachment_id": attachmentID}
	for k, v := range extra {
		details[k] = v
	}
	mod := "gr33ncore"
	tbl := "zones"
	rid := strconv.FormatInt(zoneID, 10)
	auditlog.Submit(r.Context(), h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumUpdateRecord,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details:        details,
	})
}
