package fileattach

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/fileattachutil"
	"gr33n-api/internal/filestorage"
	"gr33n-api/internal/httputil"
)

const maxReceiptUpload = 5 << 20 // 5 MiB

var receiptMimeOK = map[string]struct{}{
	"application/pdf": {},
	"image/jpeg":      {},
	"image/png":       {},
	"image/webp":      {},
}

type Handler struct {
	pool           *pgxpool.Pool
	q              *db.Queries
	store          filestorage.Store
	downloadURLTTL time.Duration
}

func NewHandler(pool *pgxpool.Pool, store filestorage.Store, downloadURLTTL time.Duration) *Handler {
	if downloadURLTTL <= 0 {
		downloadURLTTL = 5 * time.Minute
	}
	return &Handler{pool: pool, q: db.New(pool), store: store, downloadURLTTL: downloadURLTTL}
}

// UploadCostReceipt — POST /farms/{id}/cost-receipts (multipart: file, optional cost_transaction_id)
func (h *Handler) UploadCostReceipt(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, farmID) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxReceiptUpload+512*1024)
	if err := r.ParseMultipartForm(maxReceiptUpload); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "file field required")
		return
	}
	defer file.Close()

	mime := strings.ToLower(strings.TrimSpace(hdr.Header.Get("Content-Type")))
	if _, ok := receiptMimeOK[mime]; !ok {
		httputil.WriteError(w, http.StatusBadRequest, "unsupported file type (use PDF, JPEG, PNG, or WebP)")
		return
	}

	ext := filestorage.ExtForMime(mime)
	key := "farm-" + strconv.FormatInt(farmID, 10) + "/" + uuid.New().String() + ext
	n, err := h.store.Put(r.Context(), key, file, maxReceiptUpload)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var uid pgtype.UUID
	if u, ok := authctx.UserID(r.Context()); ok {
		uid = pgtype.UUID{Bytes: u, Valid: true}
	}

	relatedID := "draft"
	var costID *int64
	if v := strings.TrimSpace(r.FormValue("cost_transaction_id")); v != "" {
		cid, err := strconv.ParseInt(v, 10, 64)
		if err != nil || cid < 1 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid cost_transaction_id")
			return
		}
		tx, err := h.q.GetCostTransactionByID(r.Context(), cid)
		if err != nil {
			if err == pgx.ErrNoRows {
				httputil.WriteError(w, http.StatusNotFound, "cost transaction not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if tx.FarmID != farmID {
			httputil.WriteError(w, http.StatusForbidden, "cost belongs to another farm")
			return
		}
		relatedID = strconv.FormatInt(cid, 10)
		costID = &cid
	}

	sz := n
	att, err := h.q.CreateFileAttachment(r.Context(), db.CreateFileAttachmentParams{
		FarmID:              farmID,
		RelatedModuleSchema: "gr33ncore",
		RelatedTableName:    "cost_transactions",
		RelatedRecordID:     relatedID,
		FileName:            hdr.Filename,
		FileType:            "cost_receipt",
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

	if costID != nil {
		tx, err := h.q.GetCostTransactionByID(r.Context(), *costID)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		rid := att.ID
		oldReceiptID := tx.ReceiptFileID
		row, err := h.q.UpdateCostTransaction(r.Context(), db.UpdateCostTransactionParams{
			ID:              tx.ID,
			TransactionDate: tx.TransactionDate,
			Category:        tx.Category,
			Subcategory:     tx.Subcategory,
			Amount:          tx.Amount,
			Currency:        tx.Currency,
			Description:     tx.Description,
			IsIncome:        tx.IsIncome,
			ReceiptFileID:   &rid,
		})
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		replaced := oldReceiptID != nil
		h.logReceiptAudit(r, farmID, att.ID, "cost_receipt_uploaded", map[string]any{
			"cost_transaction_id": *costID,
			"replaced_receipt":    replaced,
		})
		h.cleanupReplacedReceipt(r.Context(), oldReceiptID, rid)
		httputil.WriteJSON(w, http.StatusCreated, map[string]any{
			"file_attachment":  att,
			"cost_transaction": row,
		})
		return
	}

	h.logReceiptAudit(r, farmID, att.ID, "cost_receipt_uploaded", map[string]any{"draft": true})
	httputil.WriteJSON(w, http.StatusCreated, att)
}

func (h *Handler) logReceiptAudit(r *http.Request, farmID, attachmentID int64, kind string, extra map[string]any) {
	details := map[string]any{"kind": kind, "file_attachment_id": attachmentID}
	for k, v := range extra {
		details[k] = v
	}
	mod := "gr33ncore"
	tbl := "file_attachments"
	aid := strconv.FormatInt(attachmentID, 10)
	auditlog.Submit(r.Context(), h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumCreateRecord,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &aid,
		Details:        details,
	})
}

func (h *Handler) logReceiptAccess(r *http.Request, att db.Gr33ncoreFileAttachment, endpoint string) {
	mod := "gr33ncore"
	tbl := "file_attachments"
	aid := strconv.FormatInt(att.ID, 10)
	auditlog.Submit(r.Context(), h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(att.FarmID),
		Action:         db.Gr33ncoreUserActionTypeEnumExportData,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &aid,
		Details: map[string]any{
			"kind":          "cost_receipt_access",
			"endpoint":      endpoint,
			"file_type":     att.FileType,
			"related_table": att.RelatedTableName,
		},
	})
}

// Download — GET /file-attachments/{id}/content
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	att, err := h.loadReadableAttachment(w, r)
	if err != nil {
		return
	}
	h.logReceiptAccess(r, att, "content")
	rc, err := h.store.Open(r.Context(), att.StoragePath)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "stored file missing")
		return
	}
	defer rc.Close()

	mt := contentType(att)
	w.Header().Set("Content-Type", mt)
	w.Header().Set("Content-Disposition", inlineDisposition(att.FileName))
	if att.FileSizeBytes != nil {
		w.Header().Set("Content-Length", strconv.FormatInt(*att.FileSizeBytes, 10))
	}
	_, _ = io.Copy(w, rc)
}

// DownloadTarget — GET /file-attachments/{id}/download
func (h *Handler) DownloadTarget(w http.ResponseWriter, r *http.Request) {
	att, err := h.loadReadableAttachment(w, r)
	if err != nil {
		return
	}
	h.logReceiptAccess(r, att, "download")
	url, err := h.store.DownloadURL(r.Context(), att.StoragePath, att.FileName, contentType(att), h.downloadURLTTL)
	if err != nil {
		// Local storage and any non-presigning backends continue to use the proxied content endpoint.
		if errors.Is(err, filestorage.ErrDownloadURLNotSupported) {
			httputil.WriteJSON(w, http.StatusOK, map[string]any{
				"url":        "/file-attachments/" + strconv.FormatInt(att.ID, 10) + "/content",
				"backend":    h.store.Backend(),
				"proxied":    true,
				"expires_at": nil,
				"file_name":  att.FileName,
			})
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"url":        url,
		"backend":    h.store.Backend(),
		"proxied":    false,
		"expires_at": time.Now().Add(h.downloadURLTTL).UTC().Format(time.RFC3339),
		"file_name":  att.FileName,
	})
}

func (h *Handler) loadReadableAttachment(w http.ResponseWriter, r *http.Request) (db.Gr33ncoreFileAttachment, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid attachment id")
		return db.Gr33ncoreFileAttachment{}, err
	}
	att, err := h.q.GetFileAttachmentByID(r.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "attachment not found")
			return db.Gr33ncoreFileAttachment{}, err
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return db.Gr33ncoreFileAttachment{}, err
	}
	if att.RelatedTableName != "cost_transactions" {
		httputil.WriteError(w, http.StatusNotFound, "attachment not found")
		return db.Gr33ncoreFileAttachment{}, errors.New("unsupported attachment table")
	}
	if !farmauthz.RequireCostRead(w, r, h.q, att.FarmID) {
		return db.Gr33ncoreFileAttachment{}, errors.New("forbidden")
	}
	return att, nil
}

func contentType(att db.Gr33ncoreFileAttachment) string {
	mt := "application/octet-stream"
	if att.MimeType != nil && *att.MimeType != "" {
		mt = *att.MimeType
	}
	return mt
}

func inlineDisposition(fileName string) string {
	return `inline; filename="` + strings.ReplaceAll(fileName, `"`, ``) + `"`
}

func (h *Handler) cleanupReplacedReceipt(ctx context.Context, oldReceiptID *int64, newReceiptID int64) {
	if oldReceiptID == nil || *oldReceiptID == newReceiptID {
		return
	}
	if err := fileattachutil.DeleteAttachmentIfUnreferenced(ctx, h.pool, h.store, *oldReceiptID); err != nil {
		log.Printf("receipt cleanup old attachment %d: %v", *oldReceiptID, err)
	}
}
