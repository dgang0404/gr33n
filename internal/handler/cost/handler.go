package cost

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"log"
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
	"gr33n-api/internal/fileattachutil"
	"gr33n-api/internal/filestorage"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

type glAccount struct {
	Code string
	Name string
}

type coaMappingView struct {
	Category    string `json:"category"`
	AccountCode string `json:"account_code"`
	AccountName string `json:"account_name"`
	Source      string `json:"source"` // default | override
}

var costCategoryToGL = map[commontypes.CostCategoryEnum]glAccount{
	commontypes.CostCategorySeedsPlants:               {Code: "5100", Name: "Seeds and plants expense"},
	commontypes.CostCategoryFertilizersSoilAmendments: {Code: "5110", Name: "Fertilizers and soil amendments"},
	commontypes.CostCategoryPestDiseaseControl:        {Code: "5120", Name: "Pest and disease control"},
	commontypes.CostCategoryWaterIrrigation:           {Code: "5130", Name: "Water and irrigation"},
	commontypes.CostCategoryLaborWages:                {Code: "5200", Name: "Labor and wages"},
	commontypes.CostCategoryEquipmentPurchaseRental:   {Code: "5300", Name: "Equipment purchase and rental"},
	commontypes.CostCategoryEquipmentMaintenanceFuel:  {Code: "5310", Name: "Equipment maintenance and fuel"},
	commontypes.CostCategoryUtilitiesElectricityGas:   {Code: "5400", Name: "Utilities"},
	commontypes.CostCategoryLandRentMortgage:          {Code: "5500", Name: "Land rent and mortgage"},
	commontypes.CostCategoryInsurance:                 {Code: "5600", Name: "Insurance"},
	commontypes.CostCategoryLicensesPermits:           {Code: "5700", Name: "Licenses and permits"},
	commontypes.CostCategoryFeedLivestock:             {Code: "5800", Name: "Feed and livestock"},
	commontypes.CostCategoryVeterinaryServices:        {Code: "5810", Name: "Veterinary services"},
	commontypes.CostCategoryPackagingSupplies:         {Code: "5900", Name: "Packaging supplies"},
	commontypes.CostCategoryTransportationLogistics:   {Code: "5910", Name: "Transportation and logistics"},
	commontypes.CostCategoryMarketingSales:            {Code: "5920", Name: "Marketing and sales"},
	commontypes.CostCategoryTrainingConsultancy:       {Code: "5930", Name: "Training and consultancy"},
	commontypes.CostCategoryMiscellaneous:             {Code: "5999", Name: "Miscellaneous expense"},
}

var defaultIncomeAccount = glAccount{Code: "4100", Name: "Farm income"}
var defaultCashAccount = glAccount{Code: "1000", Name: "Cash and bank"}

type Handler struct {
	pool  *pgxpool.Pool
	q     *db.Queries
	store filestorage.Store
}

func NewHandler(pool *pgxpool.Pool, store filestorage.Store) *Handler {
	return &Handler{pool: pool, q: db.New(pool), store: store}
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

// Export — GET /farms/{id}/costs/export?format=csv|gl_csv
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
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "gl_csv" {
		httputil.WriteError(w, http.StatusBadRequest, "format must be csv or gl_csv")
		return
	}
	rows, err := h.q.ListCostTransactionsByFarmExport(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	glOverrides, err := h.q.ListFarmFinanceAccountMappings(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	glMap := buildGLAccountMap(glOverrides)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	filename := "farm-costs-" + strconv.FormatInt(farmID, 10) + ".csv"
	if format == "gl_csv" {
		filename = "farm-costs-gl-" + strconv.FormatInt(farmID, 10) + ".csv"
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	cw := csv.NewWriter(w)
	if format == "gl_csv" {
		_ = cw.Write([]string{
			"date",
			"entry_type",
			"account_code",
			"account_name",
			"debit_amount",
			"credit_amount",
			"currency",
			"category",
			"description",
		})
	} else {
		_ = cw.Write([]string{"date", "category", "amount", "currency", "is_income", "description"})
	}
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
		if format == "gl_csv" {
			account := mapCostToGL(row.Category, row.IsIncome, glMap)
			debit := ""
			credit := ""
			if row.IsIncome {
				credit = amt
			} else {
				debit = amt
			}
			_ = cw.Write([]string{
				dateStr,
				boolEntryType(row.IsIncome),
				account.Code,
				account.Name,
				debit,
				credit,
				row.Currency,
				cat,
				desc,
			})
			continue
		}
		_ = cw.Write([]string{dateStr, cat, amt, row.Currency, strconv.FormatBool(row.IsIncome), desc})
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// ListCoaMappings — GET /farms/{id}/finance/coa-mappings
func (h *Handler) ListCoaMappings(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostRead(w, r, h.q, farmID) {
		return
	}
	rows, err := h.q.ListFarmFinanceAccountMappings(r.Context(), farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	overrides := map[commontypes.CostCategoryEnum]db.Gr33ncoreFarmFinanceAccountMapping{}
	for _, row := range rows {
		overrides[row.CostCategory] = row
	}
	out := make([]coaMappingView, 0, len(costCategoryToGL))
	for cat, def := range costCategoryToGL {
		if o, ok := overrides[cat]; ok {
			out = append(out, coaMappingView{
				Category:    string(cat),
				AccountCode: o.AccountCode,
				AccountName: o.AccountName,
				Source:      "override",
			})
			continue
		}
		out = append(out, coaMappingView{
			Category:    string(cat),
			AccountCode: def.Code,
			AccountName: def.Name,
			Source:      "default",
		})
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

// UpsertCoaMappings — PUT /farms/{id}/finance/coa-mappings
func (h *Handler) UpsertCoaMappings(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, farmID) {
		return
	}
	var body struct {
		Mappings []struct {
			Category    string `json:"category"`
			AccountCode string `json:"account_code"`
			AccountName string `json:"account_name"`
		} `json:"mappings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if len(body.Mappings) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "mappings required")
		return
	}
	for _, m := range body.Mappings {
		cat := commontypes.CostCategoryEnum(strings.TrimSpace(m.Category))
		if _, ok := costCategoryToGL[cat]; !ok {
			httputil.WriteError(w, http.StatusBadRequest, "invalid category: "+m.Category)
			return
		}
		code := strings.TrimSpace(m.AccountCode)
		name := strings.TrimSpace(m.AccountName)
		if code == "" || name == "" {
			httputil.WriteError(w, http.StatusBadRequest, "account_code and account_name required")
			return
		}
		_, err := h.q.UpsertFarmFinanceAccountMapping(r.Context(), db.UpsertFarmFinanceAccountMappingParams{
			FarmID:       farmID,
			CostCategory: cat,
			AccountCode:  code,
			AccountName:  name,
		})
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	h.ListCoaMappings(w, r)
}

// ResetCoaMappingByCategory — DELETE /farms/{id}/finance/coa-mappings/{category}
func (h *Handler) ResetCoaMappingByCategory(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, farmID) {
		return
	}
	cat := commontypes.CostCategoryEnum(strings.TrimSpace(r.PathValue("category")))
	if _, ok := costCategoryToGL[cat]; !ok {
		httputil.WriteError(w, http.StatusBadRequest, "invalid category")
		return
	}
	_, err = h.q.ResetFarmFinanceAccountMappingByCategory(r.Context(), db.ResetFarmFinanceAccountMappingByCategoryParams{
		FarmID:       farmID,
		CostCategory: cat,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.ListCoaMappings(w, r)
}

// ResetCoaMappingsAll — DELETE /farms/{id}/finance/coa-mappings
func (h *Handler) ResetCoaMappingsAll(w http.ResponseWriter, r *http.Request) {
	farmID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireCostWrite(w, r, h.q, farmID) {
		return
	}
	if _, err := h.q.ResetFarmFinanceAccountMappingsAll(r.Context(), farmID); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.ListCoaMappings(w, r)
}

func mapCostToGL(category commontypes.CostCategoryEnum, isIncome bool, overrides map[commontypes.CostCategoryEnum]glAccount) glAccount {
	if isIncome {
		return defaultIncomeAccount
	}
	if o, ok := overrides[category]; ok {
		return o
	}
	if a, ok := costCategoryToGL[category]; ok {
		return a
	}
	return glAccount{Code: "5998", Name: "Unmapped expense"}
}

func boolEntryType(isIncome bool) string {
	if isIncome {
		return "credit"
	}
	return "debit"
}

func buildGLAccountMap(rows []db.Gr33ncoreFarmFinanceAccountMapping) map[commontypes.CostCategoryEnum]glAccount {
	m := make(map[commontypes.CostCategoryEnum]glAccount, len(rows))
	for _, row := range rows {
		m[row.CostCategory] = glAccount{
			Code: row.AccountCode,
			Name: row.AccountName,
		}
	}
	return m
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
	oldReceiptID := existing.ReceiptFileID
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
	h.cleanupReplacedReceipt(r.Context(), oldReceiptID, row.ReceiptFileID)
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
	h.cleanupDeletedReceipt(r.Context(), existing.ReceiptFileID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) cleanupReplacedReceipt(ctx context.Context, oldReceiptID, newReceiptID *int64) {
	if oldReceiptID == nil {
		return
	}
	if newReceiptID != nil && *oldReceiptID == *newReceiptID {
		return
	}
	h.cleanupDeletedReceipt(ctx, oldReceiptID)
}

func (h *Handler) cleanupDeletedReceipt(ctx context.Context, receiptID *int64) {
	if receiptID == nil {
		return
	}
	if err := fileattachutil.DeleteAttachmentIfUnreferenced(ctx, h.pool, h.store, *receiptID); err != nil {
		log.Printf("receipt cleanup attachment %d: %v", *receiptID, err)
	}
}
