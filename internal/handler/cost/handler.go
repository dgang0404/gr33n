package cost

import (
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

// Create — POST /farms/{id}/costs
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
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
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if _, err := h.q.GetCostTransactionByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "transaction not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
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
	row, err := h.q.UpdateCostTransaction(r.Context(), db.UpdateCostTransactionParams{
		ID:              id,
		TransactionDate: td,
		Category:        commontypes.CostCategoryEnum(body.Category),
		Subcategory:     body.Subcategory,
		Amount:          amt,
		Currency:        cur,
		Description:     body.Description,
		IsIncome:        body.IsIncome,
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
	if _, err := h.q.GetCostTransactionByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "transaction not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.q.DeleteCostTransaction(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
