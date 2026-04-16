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

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmauthz"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/insertcommonsschema"
)

const (
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

func insertCommonsBundleIdempotency(b db.Gr33ncoreInsertCommonsBundle) string {
	if b.IdempotencyKey == nil {
		return ""
	}
	return strings.TrimSpace(*b.IdempotencyKey)
}

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

func (h *Handler) deliverInsertCommons(ctx context.Context, payload []byte, idempotencyKey string) (httpStatus int, respSnippet string, err error) {
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
	if k := strings.TrimSpace(idempotencyKey); k != "" && len(k) <= insertCommonsMaxIdempotencyKeyLen {
		req.Header.Set("Gr33n-Idempotency-Key", k)
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

// runOutboundInsertCommonsPost applies receiver delivery (or skip) and updates farm Insert Commons columns.
func (h *Handler) runOutboundInsertCommonsPost(ctx context.Context, farmID int64, farmRow db.Gr33ncoreFarm, payloadBytes []byte, idempotencyKey string) (
	farmRowOut db.Gr33ncoreFarm,
	eventStatus string,
	httpStatus int,
	respSnippet string,
	deliveryLabel string,
	eventHTTP *int32,
	eventErr *string,
	err error,
) {
	cfg := insertCommonsIngestFromEnv()
	var deliveryErr error
	if cfg.URL == "" {
		deliveryLabel = "skipped_no_receiver"
		httpStatus = 0
	} else {
		httpStatus, respSnippet, deliveryErr = h.deliverInsertCommons(ctx, payloadBytes, idempotencyKey)
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

	switch deliveryLabel {
	case "delivered":
		eventStatus = "delivered"
		hs := int32(httpStatus)
		eventHTTP = &hs
		farmRowOut, err = h.q.MarkFarmInsertCommonsDelivered(ctx, db.MarkFarmInsertCommonsDeliveredParams{
			ID:                              farmID,
			InsertCommonsLastDeliveryStatus: strPtr(eventStatus),
		})
	case "skipped_no_receiver":
		eventStatus = "skipped_no_receiver"
		farmRowOut, err = h.q.MarkFarmInsertCommonsSkippedReceiver(ctx, db.MarkFarmInsertCommonsSkippedReceiverParams{
			ID:                              farmID,
			InsertCommonsLastDeliveryStatus: strPtr(eventStatus),
		})
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
	}
	return farmRowOut, eventStatus, httpStatus, respSnippet, deliveryLabel, eventHTTP, eventErr, err
}

// buildInsertCommonsIngestPayloadBytes builds and validates the JSON body that sync or preview uses (no I/O except DB reads).
func (h *Handler) buildInsertCommonsIngestPayloadBytes(ctx context.Context, farmID int64, farmRow db.Gr33ncoreFarm) ([]byte, error) {
	costSummary, err := h.q.GetCostSummaryByFarm(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cost aggregates: %w", err)
	}
	categoryTotals, err := h.q.GetCostCategoryTotalsByFarm(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to load cost category aggregates: %w", err)
	}
	taskCounts, err := h.q.CountTasksByStatusForFarm(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to load task aggregates: %w", err)
	}
	deviceCounts, err := h.q.CountDevicesByStatusForFarm(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to load device aggregates: %w", err)
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
			"net":      numericToFloat64(row.Net),
			"tx_count": row.TxCount,
		})
	}

	payload := map[string]any{
		"schema_version": insertcommonsschema.SchemaVersion,
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
		return nil, fmt.Errorf("failed to build sync payload: %w", err)
	}
	if _, _, err := insertcommonsschema.ValidatePayload(payloadBytes); err != nil {
		return nil, fmt.Errorf("insert commons payload validation failed: %w", err)
	}
	return payloadBytes, nil
}

// InsertCommonsPreview — GET /farms/{id}/insert-commons/preview
// Read-only: same ingest shape as sync would build, without persisting or POSTing to a receiver.
func (h *Handler) InsertCommonsPreview(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
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

	payloadBytes, err := h.buildInsertCommonsIngestPayloadBytes(ctx, farmID, farmRow)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var payloadObj map[string]any
	if err := json.Unmarshal(payloadBytes, &payloadObj); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "invalid payload json")
		return
	}
	farmPseudo, _ := payloadObj["farm_pseudonym"].(string)
	genAtStr, _ := payloadObj["generated_at"].(string)
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"valid":          true,
		"schema_version": insertcommonsschema.SchemaVersion,
		"farm_pseudonym": farmPseudo,
		"generated_at":   genAtStr,
		"payload":        payloadObj,
		"privacy_notice": "Preview only; no data sent or stored. Same coarse aggregates as sync.",
		"receiver_contract": map[string]any{
			"method": http.MethodPost,
			"url":    strings.TrimSpace(os.Getenv("INSERT_COMMONS_INGEST_URL")),
			"auth":   "Optional Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>",
			"schema": insertcommonsschema.SchemaVersion,
		},
	})
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
			mod := "gr33ncore"
			tbl := "insert_commons_sync"
			ik := idem
			auditlog.Submit(ctx, h.q, r, auditlog.Event{
				FarmID:         auditlog.FarmIDPtr(farmID),
				Action:         db.Gr33ncoreUserActionTypeEnumExecuteAction,
				TargetSchema:   &mod,
				TargetTable:    &tbl,
				TargetRecordID: &ik,
				Details: map[string]any{
					"kind":              "insert_commons_sync",
					"duplicate":         true,
					"delivery_status":   prev.Status,
					"prior_http_status": prev.HttpStatus,
				},
			})
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
					"schema": insertcommonsschema.SchemaVersion,
				},
			})
			return
		}
		if strings.EqualFold(prev.Status, "pending_approval") {
			var bid int64
			if prev.BundleID != nil {
				bid = *prev.BundleID
			}
			mod := "gr33ncore"
			tbl := "insert_commons_sync"
			ik := idem
			auditlog.Submit(ctx, h.q, r, auditlog.Event{
				FarmID:         auditlog.FarmIDPtr(farmID),
				Action:         db.Gr33ncoreUserActionTypeEnumExecuteAction,
				TargetSchema:   &mod,
				TargetTable:    &tbl,
				TargetRecordID: &ik,
				Details: map[string]any{
					"kind":            "insert_commons_sync",
					"duplicate":       true,
					"delivery_status": prev.Status,
					"bundle_id":       bid,
				},
			})
			var prevPayload map[string]any
			_ = json.Unmarshal(prev.Payload, &prevPayload)
			var agg any
			if prevPayload != nil {
				agg = prevPayload["aggregates"]
			}
			httputil.WriteJSON(w, http.StatusOK, map[string]any{
				"ok":               false,
				"duplicate":        true,
				"pending_approval": true,
				"bundle_id":        bid,
				"farm_id":          farmID,
				"idempotency_key":  idem,
				"delivery_status":  prev.Status,
				"last_sync_at":     farmRow.InsertCommonsLastSyncAt,
				"last_attempt_at":  farmRow.InsertCommonsLastAttemptAt,
				"aggregates":       agg,
				"privacy_notice":   "Only coarse aggregates are included; revoke anytime by turning sharing off.",
				"receiver_contract": map[string]any{
					"method": http.MethodPost,
					"url":    strings.TrimSpace(os.Getenv("INSERT_COMMONS_INGEST_URL")),
					"auth":   "Optional Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>",
					"schema": insertcommonsschema.SchemaVersion,
				},
			})
			return
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to evaluate idempotency")
		return
	}

	if farmRow.InsertCommonsRequireApproval {
		if b, err := h.q.GetInsertCommonsBundlePendingByFarmIdempotencyKey(ctx, db.GetInsertCommonsBundlePendingByFarmIdempotencyKeyParams{
			FarmID:         farmID,
			IdempotencyKey: strPtr(idem),
		}); err == nil {
			var payloadObj map[string]any
			_ = json.Unmarshal(b.Payload, &payloadObj)
			var agg any
			if payloadObj != nil {
				agg = payloadObj["aggregates"]
			}
			httputil.WriteJSON(w, http.StatusOK, map[string]any{
				"ok":               false,
				"duplicate":        true,
				"pending_approval": true,
				"bundle_id":        b.ID,
				"farm_id":          farmID,
				"idempotency_key":  idem,
				"delivery_status":  "pending_approval",
				"last_sync_at":     farmRow.InsertCommonsLastSyncAt,
				"last_attempt_at":  farmRow.InsertCommonsLastAttemptAt,
				"aggregates":       agg,
				"privacy_notice":   "Only coarse aggregates are included; revoke anytime by turning sharing off.",
				"receiver_contract": map[string]any{
					"method": http.MethodPost,
					"url":    strings.TrimSpace(os.Getenv("INSERT_COMMONS_INGEST_URL")),
					"auth":   "Optional Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>",
					"schema": insertcommonsschema.SchemaVersion,
				},
			})
			return
		} else if !errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to check pending Insert Commons bundle")
			return
		}
	}

	if _, err := h.q.MarkFarmInsertCommonsAttempt(ctx, farmID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record sync attempt")
		return
	}

	payloadBytes, err := h.buildInsertCommonsIngestPayloadBytes(ctx, farmID, farmRow)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var ingestAggregates any
	var payloadMap map[string]any
	if err := json.Unmarshal(payloadBytes, &payloadMap); err == nil {
		ingestAggregates = payloadMap["aggregates"]
	}

	if farmRow.InsertCommonsRequireApproval {
		sum := sha256.Sum256(payloadBytes)
		payloadHash := hex.EncodeToString(sum[:])
		bundle, err := h.q.InsertInsertCommonsBundle(ctx, db.InsertInsertCommonsBundleParams{
			FarmID:         farmID,
			IdempotencyKey: strPtr(idem),
			PayloadHash:    payloadHash,
			Payload:        payloadBytes,
		})
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to queue Insert Commons bundle")
			return
		}
		bid := bundle.ID
		if _, err := h.q.UpsertInsertCommonsSyncEvent(ctx, db.UpsertInsertCommonsSyncEventParams{
			FarmID:         farmID,
			IdempotencyKey: strPtr(idem),
			Status:         "pending_approval",
			HttpStatus:     nil,
			Error:          nil,
			Payload:        payloadBytes,
			BundleID:       &bid,
		}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to record sync history")
			return
		}
		farmRowOut, err := h.q.GetFarmByID(ctx, farmID)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to load farm")
			return
		}
		mod := "gr33ncore"
		tbl := "insert_commons_sync"
		ik := idem
		auditlog.Submit(ctx, h.q, r, auditlog.Event{
			FarmID:         auditlog.FarmIDPtr(farmID),
			Action:         db.Gr33ncoreUserActionTypeEnumExecuteAction,
			TargetSchema:   &mod,
			TargetTable:    &tbl,
			TargetRecordID: &ik,
			Details: map[string]any{
				"kind":            "insert_commons_sync",
				"delivery_status": "pending_approval",
				"bundle_id":       bid,
				"ok":              false,
			},
		})
		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"ok":               false,
			"pending_approval": true,
			"bundle_id":        bid,
			"farm_id":          farmID,
			"idempotency_key":  idem,
			"delivery_status":  "pending_approval",
			"last_sync_at":     farmRowOut.InsertCommonsLastSyncAt,
			"last_attempt_at":  farmRowOut.InsertCommonsLastAttemptAt,
			"aggregates":       ingestAggregates,
			"privacy_notice":   "Only coarse aggregates are included; revoke anytime by turning sharing off.",
			"receiver_contract": map[string]any{
				"method": http.MethodPost,
				"url":    strings.TrimSpace(os.Getenv("INSERT_COMMONS_INGEST_URL")),
				"auth":   "Optional Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>",
				"schema": insertcommonsschema.SchemaVersion,
			},
		})
		return
	}

	farmRowOut, eventStatus, httpStatus, respSnippet, deliveryLabel, eventHTTP, eventErr, err := h.runOutboundInsertCommonsPost(ctx, farmID, farmRow, payloadBytes, idem)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to finalize Insert Commons sync")
		return
	}
	if deliveryLabel != "delivered" && deliveryLabel != "skipped_no_receiver" {
		retryable := deliveryLabel == "failed_retryable" || deliveryLabel == "failed_transport"
		if retryable && farmRowOut.InsertCommonsBackoffUntil.Valid {
			retryAfter := int(math.Ceil(time.Until(farmRowOut.InsertCommonsBackoffUntil.Time).Seconds()))
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
		BundleID:       nil,
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
		"aggregates":      ingestAggregates,
		"privacy_notice":  "Only coarse aggregates are included; revoke anytime by turning sharing off.",
		"receiver_contract": map[string]any{
			"method": http.MethodPost,
			"url":    strings.TrimSpace(os.Getenv("INSERT_COMMONS_INGEST_URL")),
			"auth":   "Optional Authorization: Bearer <INSERT_COMMONS_SHARED_SECRET>",
			"schema": insertcommonsschema.SchemaVersion,
		},
	}
	if strings.TrimSpace(respSnippet) != "" {
		resp["receiver_error_excerpt"] = respSnippet
	}

	st := "success"
	if !ok {
		st = "failure"
	}
	mod := "gr33ncore"
	tbl := "insert_commons_sync"
	ik := idem
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumExecuteAction,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &ik,
		Status:         st,
		Details: map[string]any{
			"kind":            "insert_commons_sync",
			"delivery_status": eventStatus,
			"http_status":     httpStatus,
			"ok":              ok,
		},
	})

	httputil.WriteJSON(w, status, resp)
}

func insertCommonsBundlePublicRow(b db.Gr33ncoreInsertCommonsBundle) map[string]any {
	out := map[string]any{
		"id":           b.ID,
		"farm_id":      b.FarmID,
		"payload_hash": b.PayloadHash,
		"status":       b.Status,
		"created_at":   b.CreatedAt,
		"updated_at":   b.UpdatedAt,
	}
	if b.IdempotencyKey != nil {
		out["idempotency_key"] = *b.IdempotencyKey
	}
	if b.ReviewerUserID.Valid {
		out["reviewer_user_id"] = uuid.UUID(b.ReviewerUserID.Bytes).String()
	}
	if b.ReviewedAt.Valid {
		out["reviewed_at"] = b.ReviewedAt.Time
	}
	if b.ReviewNote != nil {
		out["review_note"] = *b.ReviewNote
	}
	if b.DeliveryHttpStatus != nil {
		out["delivery_http_status"] = *b.DeliveryHttpStatus
	}
	if b.DeliveryError != nil {
		out["delivery_error"] = *b.DeliveryError
	}
	return out
}

func (h *Handler) syncBundleAndEventAfterDelivery(ctx context.Context, farmID int64, bundle db.Gr33ncoreInsertCommonsBundle, eventStatus string, httpStatus int, respSnippet string, deliveryLabel string, eventHTTP *int32, eventErr *string) error {
	bid := bundle.ID
	if _, err := h.q.UpsertInsertCommonsSyncEvent(ctx, db.UpsertInsertCommonsSyncEventParams{
		FarmID:         farmID,
		IdempotencyKey: bundle.IdempotencyKey,
		Status:         eventStatus,
		HttpStatus:     eventHTTP,
		Error:          eventErr,
		Payload:        bundle.Payload,
		BundleID:       &bid,
	}); err != nil {
		return err
	}
	switch deliveryLabel {
	case "delivered":
		hs := int32(httpStatus)
		_, err := h.q.MarkInsertCommonsBundleDelivered(ctx, db.MarkInsertCommonsBundleDeliveredParams{
			ID:                 bid,
			DeliveryHttpStatus: &hs,
			FarmID:             farmID,
		})
		return err
	case "skipped_no_receiver":
		_, err := h.q.MarkInsertCommonsBundleDelivered(ctx, db.MarkInsertCommonsBundleDeliveredParams{
			ID:                 bid,
			DeliveryHttpStatus: nil,
			FarmID:             farmID,
		})
		return err
	default:
		hs := int32(httpStatus)
		msg := respSnippet
		if eventErr != nil {
			msg = *eventErr
		}
		_, err := h.q.MarkInsertCommonsBundleDeliveryFailed(ctx, db.MarkInsertCommonsBundleDeliveryFailedParams{
			ID:                 bid,
			DeliveryHttpStatus: &hs,
			DeliveryError:      &msg,
			FarmID:             farmID,
		})
		return err
	}
}

// ListInsertCommonsBundles — GET /farms/{id}/insert-commons/bundles
func (h *Handler) ListInsertCommonsBundles(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	if !farmauthz.RequireFarmCaps(w, r, h.q, farmID, func(c farmauthz.FarmCaps) bool {
		return c.Admin || c.EditCosts
	}, "insufficient role to view Insert Commons bundles") {
		return
	}
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
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
	rows, err := h.q.ListInsertCommonsBundlesByFarm(ctx, db.ListInsertCommonsBundlesByFarmParams{
		FarmID:  farmID,
		Column2: statusFilter,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list bundles")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, insertCommonsBundlePublicRow(row))
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

// ApproveInsertCommonsBundleHTTP — POST /farms/{id}/insert-commons/bundles/{bundle_id}/approve
func (h *Handler) ApproveInsertCommonsBundleHTTP(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	bundleID, err := strconv.ParseInt(r.PathValue("bundle_id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid bundle id")
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
		Note *string `json:"note"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	bundle, err := h.q.GetInsertCommonsBundleByID(ctx, bundleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "bundle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load bundle")
		return
	}
	if bundle.FarmID != farmID {
		httputil.WriteError(w, http.StatusForbidden, "bundle belongs to another farm")
		return
	}
	if bundle.Status != "pending_approval" {
		httputil.WriteError(w, http.StatusBadRequest, "bundle is not awaiting approval")
		return
	}

	bundle, err = h.q.ApproveInsertCommonsBundle(ctx, db.ApproveInsertCommonsBundleParams{
		ID:             bundleID,
		ReviewerUserID: pgtype.UUID{Bytes: uid, Valid: true},
		ReviewNote:     body.Note,
		FarmID:         farmID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusConflict, "could not approve bundle (wrong state or farm)")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to approve bundle")
		return
	}

	farmRow, err := h.q.GetFarmByID(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load farm")
		return
	}
	if farmRow.InsertCommonsBackoffUntil.Valid && time.Now().Before(farmRow.InsertCommonsBackoffUntil.Time) {
		retryAfter := int(math.Ceil(time.Until(farmRow.InsertCommonsBackoffUntil.Time).Seconds()))
		if retryAfter < 1 {
			retryAfter = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		httputil.WriteError(w, http.StatusTooManyRequests, "Insert Commons sync is temporarily backing off; retry delivery later")
		return
	}

	farmRowOut, eventStatus, httpStatus, respSnippet, deliveryLabel, eventHTTP, eventErr, err := h.runOutboundInsertCommonsPost(ctx, farmID, farmRow, bundle.Payload, insertCommonsBundleIdempotency(bundle))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to deliver bundle")
		return
	}
	if err := h.syncBundleAndEventAfterDelivery(ctx, farmID, bundle, eventStatus, httpStatus, respSnippet, deliveryLabel, eventHTTP, eventErr); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record bundle delivery outcome")
		return
	}
	if deliveryLabel != "delivered" && deliveryLabel != "skipped_no_receiver" {
		retryable := deliveryLabel == "failed_retryable" || deliveryLabel == "failed_transport"
		if retryable && farmRowOut.InsertCommonsBackoffUntil.Valid {
			retryAfter := int(math.Ceil(time.Until(farmRowOut.InsertCommonsBackoffUntil.Time).Seconds()))
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		}
	}

	deliveryOK := deliveryLabel == "delivered" || deliveryLabel == "skipped_no_receiver"
	respStatus := http.StatusOK
	if !deliveryOK {
		respStatus = http.StatusBadGateway
	}
	mod := "gr33ncore"
	tbl := "insert_commons_bundles"
	rid := strconv.FormatInt(bundleID, 10)
	stAudit := "failure"
	if deliveryOK {
		stAudit = "success"
	}
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumExecuteAction,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Status:         stAudit,
		Details: map[string]any{
			"kind":            "insert_commons_bundle_approved",
			"bundle_id":       bundleID,
			"delivery_status": eventStatus,
			"http_status":     httpStatus,
		},
	})
	resp := map[string]any{
		"ok":              deliveryOK,
		"farm_id":         farmID,
		"bundle_id":       bundleID,
		"delivery_status": eventStatus,
		"http_status":     httpStatus,
		"last_sync_at":    farmRowOut.InsertCommonsLastSyncAt,
		"last_attempt_at": farmRowOut.InsertCommonsLastAttemptAt,
	}
	if strings.TrimSpace(respSnippet) != "" {
		resp["receiver_error_excerpt"] = respSnippet
	}
	httputil.WriteJSON(w, respStatus, resp)
}

// RejectInsertCommonsBundleHTTP — POST /farms/{id}/insert-commons/bundles/{bundle_id}/reject
func (h *Handler) RejectInsertCommonsBundleHTTP(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	bundleID, err := strconv.ParseInt(r.PathValue("bundle_id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid bundle id")
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
		Note string `json:"note"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 8<<10)).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	note := strings.TrimSpace(body.Note)
	if note == "" {
		httputil.WriteError(w, http.StatusBadRequest, "note is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	bundle, err := h.q.GetInsertCommonsBundleByID(ctx, bundleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "bundle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load bundle")
		return
	}
	if bundle.FarmID != farmID {
		httputil.WriteError(w, http.StatusForbidden, "bundle belongs to another farm")
		return
	}
	if bundle.Status != "pending_approval" {
		httputil.WriteError(w, http.StatusBadRequest, "bundle is not awaiting approval")
		return
	}

	bundle, err = h.q.RejectInsertCommonsBundle(ctx, db.RejectInsertCommonsBundleParams{
		ID:             bundleID,
		ReviewerUserID: pgtype.UUID{Bytes: uid, Valid: true},
		ReviewNote:     &note,
		FarmID:         farmID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusConflict, "could not reject bundle")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reject bundle")
		return
	}
	bid := bundle.ID
	if _, err := h.q.UpsertInsertCommonsSyncEvent(ctx, db.UpsertInsertCommonsSyncEventParams{
		FarmID:         farmID,
		IdempotencyKey: bundle.IdempotencyKey,
		Status:         "rejected",
		HttpStatus:     nil,
		Error:          &note,
		Payload:        bundle.Payload,
		BundleID:       &bid,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update sync history")
		return
	}
	mod := "gr33ncore"
	tbl := "insert_commons_bundles"
	rid := strconv.FormatInt(bundleID, 10)
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumExecuteAction,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details: map[string]any{
			"kind":      "insert_commons_bundle_rejected",
			"bundle_id": bundleID,
		},
	})
	httputil.WriteJSON(w, http.StatusOK, insertCommonsBundlePublicRow(bundle))
}

// RetryInsertCommonsBundleDeliver — POST /farms/{id}/insert-commons/bundles/{bundle_id}/deliver
func (h *Handler) RetryInsertCommonsBundleDeliver(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	bundleID, err := strconv.ParseInt(r.PathValue("bundle_id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid bundle id")
		return
	}
	if !farmauthz.RequireFarmAdmin(w, r, h.q, farmID) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	bundle, err := h.q.GetInsertCommonsBundleByID(ctx, bundleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "bundle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load bundle")
		return
	}
	if bundle.FarmID != farmID {
		httputil.WriteError(w, http.StatusForbidden, "bundle belongs to another farm")
		return
	}
	if bundle.Status != "delivery_failed" {
		httputil.WriteError(w, http.StatusBadRequest, "bundle is not in delivery_failed state")
		return
	}

	farmRow, err := h.q.GetFarmByID(ctx, farmID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load farm")
		return
	}
	if farmRow.InsertCommonsBackoffUntil.Valid && time.Now().Before(farmRow.InsertCommonsBackoffUntil.Time) {
		retryAfter := int(math.Ceil(time.Until(farmRow.InsertCommonsBackoffUntil.Time).Seconds()))
		if retryAfter < 1 {
			retryAfter = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		httputil.WriteError(w, http.StatusTooManyRequests, "Insert Commons sync is temporarily backing off")
		return
	}

	farmRowOut, eventStatus, httpStatus, respSnippet, deliveryLabel, eventHTTP, eventErr, err := h.runOutboundInsertCommonsPost(ctx, farmID, farmRow, bundle.Payload, insertCommonsBundleIdempotency(bundle))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusForbidden, "Insert Commons sharing is disabled for this farm")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to deliver bundle")
		return
	}
	if err := h.syncBundleAndEventAfterDelivery(ctx, farmID, bundle, eventStatus, httpStatus, respSnippet, deliveryLabel, eventHTTP, eventErr); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record bundle delivery outcome")
		return
	}
	if deliveryLabel != "delivered" && deliveryLabel != "skipped_no_receiver" {
		retryable := deliveryLabel == "failed_retryable" || deliveryLabel == "failed_transport"
		if retryable && farmRowOut.InsertCommonsBackoffUntil.Valid {
			retryAfter := int(math.Ceil(time.Until(farmRowOut.InsertCommonsBackoffUntil.Time).Seconds()))
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		}
	}
	ok := deliveryLabel == "delivered" || deliveryLabel == "skipped_no_receiver"
	status := http.StatusOK
	if !ok {
		status = http.StatusBadGateway
	}
	resp := map[string]any{
		"ok":              ok,
		"farm_id":         farmID,
		"bundle_id":       bundleID,
		"delivery_status": eventStatus,
		"http_status":     httpStatus,
		"last_sync_at":    farmRowOut.InsertCommonsLastSyncAt,
		"last_attempt_at": farmRowOut.InsertCommonsLastAttemptAt,
	}
	if strings.TrimSpace(respSnippet) != "" {
		resp["receiver_error_excerpt"] = respSnippet
	}
	httputil.WriteJSON(w, status, resp)
}

// ExportInsertCommonsBundle — GET /farms/{id}/insert-commons/bundles/{bundle_id}/export
func (h *Handler) ExportInsertCommonsBundle(w http.ResponseWriter, r *http.Request) {
	farmID, err := httputil.PathID(r.URL.Path, 2)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid farm id")
		return
	}
	bundleID, err := strconv.ParseInt(r.PathValue("bundle_id"), 10, 64)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid bundle id")
		return
	}
	if !farmauthz.RequireFarmCaps(w, r, h.q, farmID, func(c farmauthz.FarmCaps) bool {
		return c.Admin || c.EditCosts
	}, "insufficient role to export Insert Commons bundles") {
		return
	}
	format := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("format")))
	if format == "" {
		format = "ingest"
	}
	if format != "ingest" && format != "package_v1" {
		httputil.WriteError(w, http.StatusBadRequest, "format must be ingest or package_v1")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	bundle, err := h.q.GetInsertCommonsBundleByID(ctx, bundleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "bundle not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load bundle")
		return
	}
	if bundle.FarmID != farmID {
		httputil.WriteError(w, http.StatusForbidden, "bundle belongs to another farm")
		return
	}

	filename := fmt.Sprintf("insert-commons-farm-%d-bundle-%d.json", farmID, bundleID)
	var out []byte
	switch format {
	case "ingest":
		out = bundle.Payload
	default:
		var inner any
		if err := json.Unmarshal(bundle.Payload, &inner); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "invalid stored payload")
			return
		}
		pkg := map[string]any{
			"package_version": insertcommonsschema.PackageVersionV1,
			"exported_at":     time.Now().UTC().Format(time.RFC3339Nano),
			"farm_id":         farmID,
			"bundle_id":       bundleID,
			"bundle_status":   bundle.Status,
			"payload_hash":    bundle.PayloadHash,
			"scrub_summary": map[string]any{
				"ingest_schema": insertcommonsschema.SchemaVersion,
				"notes":         "Payload built by gr33n farm API; only coarse aggregates.",
			},
			"payload": inner,
		}
		if bundle.IdempotencyKey != nil {
			pkg["idempotency_key"] = *bundle.IdempotencyKey
		}
		out, err = json.Marshal(pkg)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to build export")
			return
		}
		filename = fmt.Sprintf("insert-commons-package-farm-%d-bundle-%d.json", farmID, bundleID)
	}

	mod := "gr33ncore"
	tbl := "insert_commons_bundles"
	rid := strconv.FormatInt(bundleID, 10)
	auditlog.Submit(ctx, h.q, r, auditlog.Event{
		FarmID:         auditlog.FarmIDPtr(farmID),
		Action:         db.Gr33ncoreUserActionTypeEnumExportData,
		TargetSchema:   &mod,
		TargetTable:    &tbl,
		TargetRecordID: &rid,
		Details: map[string]any{
			"kind":      "insert_commons_bundle_export",
			"format":    format,
			"bytes":     len(out),
			"bundle_id": bundleID,
		},
	})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
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
		if row.BundleID != nil {
			item["bundle_id"] = *row.BundleID
		}
		out = append(out, item)
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}
