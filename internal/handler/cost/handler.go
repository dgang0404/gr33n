package cost

import (
	"encoding/csv"
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
	"gr33n-api/internal/platform/commontypes"
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

func numericToFloat64(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
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

// List — GET /farms/{id}/costs
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostRead(w, r, h.q, farmID) {
		return
	}
	limit := int32(50)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n < 1 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		if n > 500 {
			n = 500
		}
		limit = int32(n)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n < 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = int32(n)
	}
	rows, err := h.q.ListCostTransactionsByFarm(r.Context(), db.ListCostTransactionsByFarmParams{
		FarmID: farmID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == nil {
		rows = []db.Gr33ncoreCostTransaction{}
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// Summary — GET /farms/{id}/costs/summary
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostRead(w, r, h.q, farmID) {
		return
	}
	row, err := h.q.GetCostSummaryByFarm(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]float64{
		"total_income":   numericToFloat64(row.TotalIncome),
		"total_expenses": numericToFloat64(row.TotalExpenses),
		"net":            numericToFloat64(row.Net),
	})
}

// Export — GET /farms/{id}/costs/export?format=csv
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostRead(w, r, h.q, farmID) {
		return
	}
	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format != "" && format != "csv" {
		httputil.WriteError(w, http.StatusBadRequest, "format must be csv")
		return
	}
	rows, err := h.q.ListCostTransactionsByFarmExport(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="farm-costs-`+strconv.FormatInt(farmID, 10)+`.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"date", "category", "amount", "currency", "is_income", "description"})
	for _, row := range rows {
		desc := ""
		if row.Description != nil {
			desc = *row.Description
		}
		sub := ""
		if row.Subcategory != nil {
			sub = *row.Subcategory
		}
		cat := string(row.Category)
		if sub != "" {
			cat = cat + " / " + sub
		}
		amt := ""
		if f, err := row.Amount.Float64Value(); err == nil && f.Valid {
			amt = strconv.FormatFloat(f.Float64, 'f', -1, 64)
		}
		dateStr := ""
		if row.TransactionDate.Valid {
			dateStr = row.TransactionDate.Time.Format("2006-01-02")
		}
		_ = cw.Write([]string{
			dateStr,
			cat,
			amt,
			row.Currency,
			strconv.FormatBool(row.IsIncome),
			desc,
		})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// Create — POST /farms/{id}/costs
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, farmID) {
		return
	}
	var body struct {
		TransactionDate string  `json:"transaction_date"`
		Category        string  `json:"category"`
		Subcategory     *string `json:"subcategory"`
		Amount          float64 `json:"amount"`
		Currency        string  `json:"currency"`
		Description     *string `json:"description"`
		IsIncome        bool    `json:"is_income"`
		ReceiptFileID   *int64  `json:"receipt_file_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	td, err := parseDate(body.TransactionDate)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid transaction_date (YYYY-MM-DD)")
		return
	}
	if strings.TrimSpace(body.Category) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "category required")
		return
	}
	cur := strings.TrimSpace(strings.ToUpper(body.Currency))
	if len(cur) != 3 {
		httputil.WriteError(w, http.StatusBadRequest, "currency must be a 3-letter ISO code")
		return
	}
	amt, err := numericFromFloat64(body.Amount)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid amount")
		return
	}
	if body.ReceiptFileID != nil {
		att, err := h.q.GetFileAttachmentByID(r.Context(), *body.ReceiptFileID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "receipt_file_id not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if att.FarmID != farmID || att.RelatedTableName != "cost_transactions" {
			httputil.WriteError(w, http.StatusBadRequest, "invalid receipt_file_id")
			return
		}
	}
	var createdBy pgtype.UUID
	if uid, ok := authctx.UserID(r.Context()); ok {
		createdBy = pgtype.UUID{Bytes: uid, Valid: true}
	}
	row, err := h.q.CreateCostTransaction(r.Context(), db.CreateCostTransactionParams{
		FarmID:          farmID,
		TransactionDate: td,
		Category:        commontypes.CostCategoryEnum(body.Category),
		Subcategory:     body.Subcategory,
		Amount:          amt,
		Currency:        cur,
		Description:     body.Description,
		IsIncome:        body.IsIncome,
		CreatedByUserID: createdBy,
		ReceiptFileID:   body.ReceiptFileID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, row)
}

// Update — PUT /costs/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid cost id")
		return
	}
	var body struct {
		TransactionDate string  `json:"transaction_date"`
		Category        string  `json:"category"`
		Subcategory     *string `json:"subcategory"`
		Amount          float64 `json:"amount"`
		Currency        string  `json:"currency"`
		Description     *string `json:"description"`
		IsIncome        bool    `json:"is_income"`
		ReceiptFileID   *int64  `json:"receipt_file_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	existing, err := h.q.GetCostTransactionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "transaction not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, existing.FarmID) {
		return
	}
	td, err := parseDate(body.TransactionDate)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid transaction_date")
		return
	}
	if strings.TrimSpace(body.Category) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "category required")
		return
	}
	cur := strings.TrimSpace(strings.ToUpper(body.Currency))
	if len(cur) != 3 {
		httputil.WriteError(w, http.StatusBadRequest, "currency must be a 3-letter ISO code")
		return
	}
	amt, err := numericFromFloat64(body.Amount)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid amount")
		return
	}
	receiptID := existing.ReceiptFileID
	if body.ReceiptFileID != nil {
		att, err := h.q.GetFileAttachmentByID(r.Context(), *body.ReceiptFileID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusBadRequest, "receipt_file_id not found")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if att.FarmID != existing.FarmID || att.RelatedTableName != "cost_transactions" {
			httputil.WriteError(w, http.StatusBadRequest, "invalid receipt_file_id")
			return
		}
		receiptID = body.ReceiptFileID
	}
	row, err := h.q.UpdateCostTransaction(r.Context(), db.UpdateCostTransactionParams{
		ID:              id,
		TransactionDate: td,
		Category:        commontypes.CostCategoryEnum(body.Category),
		Subcategory:     body.Subcategory,
		Amount:          amt,
		Currency:        cur,
		Description:     body.Description,
		IsIncome:        body.IsIncome,
		ReceiptFileID:   receiptID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "transaction not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, row)
}

// Delete — DELETE /costs/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid cost id")
		return
	}
	existing, err := h.q.GetCostTransactionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "transaction not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, existing.FarmID) {
		return
	}
	if err := h.q.DeleteCostTransaction(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
