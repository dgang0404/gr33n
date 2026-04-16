package farm

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
)

const (
	insertCommonsPayloadVersion        = "gr33n.insert_commons.v1"
	insertCommonsMaxAttemptsPer10Min   = 20
	insertCommonsMaxIdempotencyKeyLen  = 128
	insertCommonsIngestBodyMaxBytes    = 1 << 20
	insertCommonsIngestErrorSnippetLen = 4096
)

func numericToFloat64(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}

func strPtr(s string) *string { return &s }

func insertCommonsPseudonymKey() []byte {
	if k := strings.TrimSpace(os.Getenv("INSERT_COMMONS_PSEUDONYM_KEY")); k != "" {
		return []byte(k)
	}
	if k := strings.TrimSpace(os.Getenv("INSERT_COMMONS_SHARED_SECRET")); k != "" {
		return []byte(k)
	}
	// Dev fallback: stable per-process, but not stable across restarts.
	return []byte("dev-only-insert-commons-pseudonym-key")
}

func farmPseudonym(farmID int64) string {
	mac := hmac.New(sha256.New, insertCommonsPseudonymKey())
	fmt.Fprintf(mac, "gr33n:farm:%d", farmID)
	sum := mac.Sum(nil)
	// Short, non-reversible token for receivers; still unique enough at federation scale.
	return hex.EncodeToString(sum[:12])
}

func coarseTimezoneLabel(tz string) string {
	t := strings.TrimSpace(strings.ToUpper(tz))
	if t == "" || t == "UTC" || t == "ETC/UTC" {
		return "UTC"
	}
	if strings.Contains(t, "/") {
		return "IANA_REGIONAL"
	}
	return "OTHER"
}

func insertCommonsBackoffDuration(consecutiveFailures int32) time.Duration {
	if consecutiveFailures <= 0 {
		return 0
	}
	// Exponential backoff starting at 30s, capped at 1h.
	exp := int(consecutiveFailures) - 1
	if exp < 0 {
		exp = 0
	}
	if exp > 10 {
		exp = 10
	}
	d := time.Duration(30<<exp) * time.Second
	const max = time.Hour
	if d > max {
		d = max
	}
	return d
}

type insertCommonsIngestCfg struct {
	URL        string
	AuthBearer string
}

func insertCommonsIngestFromEnv() insertCommonsIngestCfg {
	return insertCommonsIngestCfg{
		URL:        strings.TrimSpace(os.Getenv("INSERT_COMMONS_INGEST_URL")),
		AuthBearer: strings.TrimSpace(os.Getenv("INSERT_COMMONS_SHARED_SECRET")),
	}
}

func (h *Handler) deliverInsertCommons(ctx context.Context, payload []byte) (httpStatus int, respSnippet string, err error) {
	cfg := insertCommonsIngestFromEnv()
	if cfg.URL == "" {
		return 0, "", nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(payload))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.AuthBearer != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AuthBearer)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, insertCommonsIngestBodyMaxBytes))
	snippet := strings.TrimSpace(string(body))
	if len(snippet) > insertCommonsIngestErrorSnippetLen {
		snippet = snippet[:insertCommonsIngestErrorSnippetLen] + "…"
	}
	return resp.StatusCode, snippet, nil
}

// InsertCommonsSync — POST /farms/{id}/insert-commons/sync
func (h *Handler) InsertCommonsSync(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmCaps(w, r, h.q, farmID, func(c farmauthz.FarmCaps) bool {
		return c.Admin || c.EditCosts
	}, "insufficient role to sync Insert Commons aggregates") {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	farmRow, err := h.q.GetFarmByID(ctx, farmID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "farm not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load farm")
		return
	}
	if !farmRow.InsertCommonsOptIn {
		httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
		return
	}

	idem := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idem == "" {
		var body struct {
			IdempotencyKey string `json:"idempotency_key"`
		}
		_ = json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&body) // optional body
		idem = strings.TrimSpace(body.IdempotencyKey)
	}
	if len(idem) > insertCommonsMaxIdempotencyKeyLen {
		httputil.WriteError(w, http.StatusBadRequest, "idempotency key too long")
		return
	}
	if idem == "" {
		idem = uuid.NewString()
	}

	// Rate limit (abuse protection): attempts per farm per rolling window.
	since := time.Now().Add(-10 * time.Minute)
	attempts, err := h.q.CountInsertCommonsSyncAttemptsSince(ctx, db.CountInsertCommonsSyncAttemptsSinceParams{
		FarmID:    farmID,
		CreatedAt: since,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to evaluate sync rate limit")
		return
	}
	if attempts >= insertCommonsMaxAttemptsPer10Min {
		httputil.WriteError(w, http.StatusTooManyRequests, "Insert Commons sync rate limit exceeded; try again later")
		return
	}

	// Server-enforced backoff after repeated delivery failures.
	if farmRow.InsertCommonsBackoffUntil.Valid && time.Now().Before(farmRow.InsertCommonsBackoffUntil.Time) {
		retryAfter := int(math.Ceil(time.Until(farmRow.InsertCommonsBackoffUntil.Time).Seconds()))
		if retryAfter < 1 {
			retryAfter = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		httputil.WriteError(w, http.StatusTooManyRequests, "Insert Commons sync is temporarily backing off after repeated delivery failures")
		return
	}

	// Idempotent success: replay the prior successful outcome for this key.
	if prev, err := h.q.GetInsertCommonsSyncEventByFarmIdempotencyKey(ctx, db.GetInsertCommonsSyncEventByFarmIdempotencyKeyParams{
		FarmID:         farmID,
		IdempotencyKey: strPtr(idem),
	}); err == nil {
		if strings.EqualFold(prev.Status, "delivered") || strings.EqualFold(prev.Status, "skipped_no_receiver") {
			ok := strings.EqualFold(prev.Status, "delivered") || strings.EqualFold(prev.Status, "skipped_no_receiver")
			httputil.WriteJSON(w, http.StatusOK, map[string]any{
				"ok":              ok,
				"duplicate":       true,
				"farm_id":         farmID,
				"idempotency_key": idem,
				"delivery_status": prev.Status,
				"http_status":     prev.HttpStatus,
				"last_sync_at":    farmRow.InsertCommonsLastSyncAt,
				"last_attempt_at": farmRow.InsertCommonsLastAttemptAt,
				"privacy_notice":  "Only coarse aggregates are included; revoke anytime by turning sharing off.",
				"receiver_contract": map[string]any{
					"method": http.MethodPost,
					"url":    strings.TrimSpace(os.Getenv("INSERT_COMMONS_INGEST_URL")),
					"auth":   "Optional Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>",
					"schema": insertCommonsPayloadVersion,
				},
			})
			return
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to evaluate idempotency")
		return
	}

	if _, err := h.q.MarkFarmInsertCommonsAttempt(ctx, farmID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record sync attempt")
		return
	}

	costSummary, err := h.q.GetCostSummaryByFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load cost aggregates")
		return
	}
	categoryTotals, err := h.q.GetCostCategoryTotalsByFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load cost category aggregates")
		return
	}
	taskCounts, err := h.q.CountTasksByStatusForFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load task aggregates")
		return
	}
	deviceCounts, err := h.q.CountDevicesByStatusForFarm(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load device aggregates")
		return
	}

	tasksByStatus := map[string]int64{}
	for _, row := range taskCounts {
		tasksByStatus[string(row.Status)] = row.Cnt
	}
	devicesByStatus := map[string]int64{}
	for _, row := range deviceCounts {
		devicesByStatus[string(row.Status)] = row.Cnt
	}

	categories := []map[string]any{}
	for _, row := range categoryTotals {
		categories = append(categories, map[string]any{
			"category": string(row.Category),
			"currency": strings.TrimSpace(row.Currency),
			"income":   numericToFloat64(row.Income),
			"expense":  numericToFloat64(row.Expense),
			"tx_count": row.TxCount,
		})
	}

	payload := map[string]any{
		"schema_version": insertCommonsPayloadVersion,
		"generated_at":   time.Now().UTC().Format(time.RFC3339Nano),
		"farm_pseudonym": farmPseudonym(farmID),
		"farm_profile": map[string]any{
			"scale_tier":         string(farmRow.ScaleTier),
			"timezone_bucket":    coarseTimezoneLabel(farmRow.Timezone),
			"currency":           strings.TrimSpace(farmRow.Currency),
			"operational_status": string(farmRow.OperationalStatus),
		},
		"aggregates": map[string]any{
			"costs": map[string]any{
				"totals": map[string]any{
					"income":   numericToFloat64(costSummary.TotalIncome),
					"expenses": numericToFloat64(costSummary.TotalExpenses),
					"net":      numericToFloat64(costSummary.Net),
				},
				"by_category": categories,
			},
			"tasks": map[string]any{
				"by_status": tasksByStatus,
			},
			"devices": map[string]any{
				"by_status": devicesByStatus,
			},
		},
		"privacy": map[string]any{
			"includes_pii":               false,
			"includes_raw_location_text": false,
			"revocation":                 "Turn off Insert Commons sharing for the farm to stop future sends.",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to build sync payload")
		return
	}

	cfg := insertCommonsIngestFromEnv()
	var (
		httpStatus    int
		respSnippet   string
		deliveryErr   error
		deliveryLabel string
	)

	if cfg.URL == "" {
		deliveryLabel = "skipped_no_receiver"
		httpStatus = 0
	} else {
		httpStatus, respSnippet, deliveryErr = h.deliverInsertCommons(ctx, payloadBytes)
		if deliveryErr != nil {
			deliveryLabel = "failed_transport"
		} else if httpStatus >= 200 && httpStatus <= 299 {
			deliveryLabel = "delivered"
		} else if httpStatus == http.StatusTooManyRequests || httpStatus >= 500 {
			deliveryLabel = "failed_retryable"
		} else {
			deliveryLabel = "failed_client"
		}
	}

	var (
		eventStatus string
		eventHTTP   *int32
		eventErr    *string
		farmRowOut  db.Gr33ncoreFarm
	)

	switch deliveryLabel {
	case "delivered":
		eventStatus = "delivered"
		hs := int32(httpStatus)
		eventHTTP = &hs
		farmRowOut, err = h.q.MarkFarmInsertCommonsDelivered(ctx, db.MarkFarmInsertCommonsDeliveredParams{
			ID:                              farmID,
			InsertCommonsLastDeliveryStatus: strPtr(eventStatus),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, "failed to finalize successful sync")
			return
		}
	case "skipped_no_receiver":
		eventStatus = "skipped_no_receiver"
		farmRowOut, err = h.q.MarkFarmInsertCommonsSkippedReceiver(ctx, db.MarkFarmInsertCommonsSkippedReceiverParams{
			ID:                              farmID,
			InsertCommonsLastDeliveryStatus: strPtr(eventStatus),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, "failed to finalize sync")
			return
		}
	default:
		eventStatus = deliveryLabel
		if httpStatus != 0 {
			hs := int32(httpStatus)
			eventHTTP = &hs
		}
		msg := respSnippet
		if deliveryErr != nil {
			msg = deliveryErr.Error()
		}
		if strings.TrimSpace(msg) == "" {
			msg = "delivery failed"
		}
		eventErr = &msg

		backoffUntil := time.Time{}
		retryable := deliveryLabel == "failed_retryable" || deliveryLabel == "failed_transport"
		if retryable {
			backoffUntil = time.Now().Add(insertCommonsBackoffDuration(farmRow.InsertCommonsConsecutiveFailures))
		}

		farmRowOut, err = h.q.MarkFarmInsertCommonsSyncFailure(ctx, db.MarkFarmInsertCommonsSyncFailureParams{
			ID:                              farmID,
			InsertCommonsLastDeliveryStatus: strPtr(eventStatus),
			InsertCommonsLastError:          eventErr,
			InsertCommonsBackoffUntil:       pgtype.Timestamptz{Time: backoffUntil, Valid: retryable && !backoffUntil.IsZero()},
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
				return
			}
			httputil.WriteError(w, http.StatusInternalServerError, "failed to record sync failure")
			return
		}
		if retryable && !backoffUntil.IsZero() {
			retryAfter := int(math.Ceil(time.Until(backoffUntil).Seconds()))
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		}
	}

	if _, err := h.q.UpsertInsertCommonsSyncEvent(ctx, db.UpsertInsertCommonsSyncEventParams{
		FarmID:         farmID,
		IdempotencyKey: strPtr(idem),
		Status:         eventStatus,
		HttpStatus:     eventHTTP,
		Error:          eventErr,
		Payload:        payloadBytes,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record sync history")
		return
	}

	ok := deliveryLabel == "delivered" || deliveryLabel == "skipped_no_receiver"
	status := http.StatusOK
	if !ok {
		status = http.StatusBadGateway
	}

	resp := map[string]any{
		"ok":              ok,
		"farm_id":         farmID,
		"idempotency_key": idem,
		"delivery_status": eventStatus,
		"http_status":     httpStatus,
		"last_sync_at":    farmRowOut.InsertCommonsLastSyncAt,
		"last_attempt_at": farmRowOut.InsertCommonsLastAttemptAt,
		"backoff_until":   farmRowOut.InsertCommonsBackoffUntil,
		"aggregates":      payload["aggregates"],
		"privacy_notice":  "Only coarse aggregates are included; revoke anytime by turning sharing off.",
		"receiver_contract": map[string]any{
			"method": http.MethodPost,
			"url":    strings.TrimSpace(os.Getenv("INSERT_COMMONS_INGEST_URL")),
			"auth":   "Optional Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>",
			"schema": insertCommonsPayloadVersion,
		},
	}
	if strings.TrimSpace(respSnippet) != "" {
		resp["receiver_error_excerpt"] = respSnippet
	}

	httputil.WriteJSON(w, status, resp)
}

// InsertCommonsHistory — GET /farms/{id}/insert-commons/sync-events
func (h *Handler) InsertCommonsHistory(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmCaps(w, r, h.q, farmID, func(c farmauthz.FarmCaps) bool {
		return c.Admin || c.ViewCosts
	}, "insufficient role to view Insert Commons sync history") {
		return
	}

	limit := int32(25)
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 || v > 100 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = int32(v)
	}
	offset := int32(0)
	if s := strings.TrimSpace(r.URL.Query().Get("offset")); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			httputil.WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = int32(v)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.q.ListInsertCommonsSyncEventsByFarm(ctx, db.ListInsertCommonsSyncEventsByFarmParams{
		FarmID: farmID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list sync events")
		return
	}

	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		item := map[string]any{
			"id":          row.ID,
			"farm_id":     row.FarmID,
			"status":      row.Status,
			"http_status": row.HttpStatus,
			"created_at":  row.CreatedAt,
		}
		if row.IdempotencyKey != nil {
			item["idempotency_key"] = *row.IdempotencyKey
		}
		if row.Error != nil {
			item["error"] = *row.Error
		}
		out = append(out, item)
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}
