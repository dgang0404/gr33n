package fileattach

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
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
	q     *db.Queries
	store *filestorage.Local
}

func NewHandler(pool *pgxpool.Pool, store *filestorage.Local) *Handler {
	return &Handler{q: db.New(pool), store: store}
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
	n, err := h.store.Put(key, file, maxReceiptUpload)
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
		httputil.WriteJSON(w, http.StatusCreated, map[string]any{
			"file_attachment":  att,
			"cost_transaction": row,
		})
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, att)
}

// Download — GET /file-attachments/{id}/content
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid attachment id")
		return
	}
	att, err := h.q.GetFileAttachmentByID(r.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			httputil.WriteError(w, http.StatusNotFound, "attachment not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if att.RelatedTableName != "cost_transactions" {
		httputil.WriteError(w, http.StatusNotFound, "attachment not found")
		return
	}
	if !farmauthz.RequireCostRead(w, r, h.q, att.FarmID) {
		return
	}
	rc, err := h.store.Open(att.StoragePath)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "stored file missing")
		return
	}
	defer rc.Close()

	mt := "application/octet-stream"
	if att.MimeType != nil && *att.MimeType != "" {
		mt = *att.MimeType
	}
	disposition := "inline"
	if mt == "application/pdf" {
		disposition = "inline"
	}
	w.Header().Set("Content-Type", mt)
	w.Header().Set("Content-Disposition", disposition+`; filename="`+strings.ReplaceAll(att.FileName, `"`, ``)+`"`)
	if att.FileSizeBytes != nil {
		w.Header().Set("Content-Length", strconv.FormatInt(*att.FileSizeBytes, 10))
	}
	_, _ = io.Copy(w, rc)
}
