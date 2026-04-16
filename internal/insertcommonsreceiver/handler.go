package insertcommonsreceiver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/insertcommonsschema"
)

const (
	maxBodyBytes         = 1 << 20
	maxIdempotencyKeyLen = 128
	pgerrUniqueViolation = "23505"
)

// Handler is a minimal HTTP ingest service for Insert Commons pilot deployments.
type Handler struct {
	q             *db.Queries
	sharedSecret  string
	allowNoAuth   bool
	retentionDays int
}

// NewHandler builds the receiver. If sharedSecret is empty and allowNoAuth is true, ingest is
// unauthenticated (local pilots only). Otherwise sharedSecret must match the farm's INSERT_COMMONS_SHARED_SECRET.
func NewHandler(pool *pgxpool.Pool, sharedSecret string, allowNoAuth bool, retentionDays int) *Handler {
	return &Handler{
		q:             db.New(pool),
		sharedSecret:  strings.TrimSpace(sharedSecret),
		allowNoAuth:   allowNoAuth,
		retentionDays: retentionDays,
	}
}

// ServeHTTP handles GET /health, GET /v1/stats, and POST /v1/ingest.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/health":
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "gr33n-insert-commons-receiver"})
	case r.Method == http.MethodGet && r.URL.Path == "/v1/stats":
		h.stats(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/v1/ingest":
		h.ingest(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	if h.sharedSecret == "" && h.allowNoAuth {
		return true
	}
	if h.sharedSecret == "" {
		httputil.WriteError(w, http.StatusInternalServerError, "INSERT_COMMONS_SHARED_SECRET is not set on the receiver")
		return false
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		httputil.WriteError(w, http.StatusUnauthorized, "Authorization: Bearer <token> required")
		return false
	}
	if strings.TrimSpace(strings.TrimPrefix(auth, prefix)) != h.sharedSecret {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid bearer token")
		return false
	}
	return true
}

func ingestIdempotencyKey(r *http.Request) (string, error) {
	k := strings.TrimSpace(r.Header.Get("Gr33n-Idempotency-Key"))
	if k == "" {
		k = strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	}
	if len(k) > maxIdempotencyKeyLen {
		return "", errors.New("idempotency key too long")
	}
	return k, nil
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	if !h.checkAuth(w, r) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	st, err := h.q.InsertCommonsReceiverStats(ctx)
	if err != nil {
		log.Printf("insert commons stats: %v", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load stats")
		return
	}
	since := time.Now().UTC().AddDate(0, 0, -30)
	days, err := h.q.InsertCommonsReceiverDailyCounts(ctx, since)
	if err != nil {
		log.Printf("insert commons daily stats: %v", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load daily stats")
		return
	}

	daily := make([]map[string]any, 0, len(days))
	for _, row := range days {
		item := map[string]any{"ingest_count": row.IngestCount}
		if row.Day.Valid {
			item["day"] = row.Day.Time.Format("2006-01-02")
		}
		daily = append(daily, item)
	}

	out := map[string]any{
		"service":             "gr33n-insert-commons-receiver",
		"retention_days":      h.retentionDays,
		"total_payloads":      st.TotalPayloads,
		"distinct_pseudonyms": st.DistinctPseudonyms,
		"ingests_by_utc_day":  daily,
	}
	if ts, ok := st.OldestReceivedAt.(time.Time); ok {
		out["oldest_received_at"] = ts.UTC().Format(time.RFC3339Nano)
	}
	if ts, ok := st.NewestReceivedAt.(time.Time); ok {
		out["newest_received_at"] = ts.UTC().Format(time.RFC3339Nano)
	}

	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ingest(w http.ResponseWriter, r *http.Request) {
	if !h.checkAuth(w, r) {
		return
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))), "application/json") {
		httputil.WriteError(w, http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	idem, err := ingestIdempotencyKey(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	raw, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "failed to read body")
		return
	}
	if len(raw) > maxBodyBytes {
		httputil.WriteError(w, http.StatusRequestEntityTooLarge, "body too large")
		return
	}

	farmPseudo, genAt, err := insertcommonsschema.ValidatePayload(raw)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	sum := sha256.Sum256(raw)
	payloadHash := hex.EncodeToString(sum[:])

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if idem != "" {
		ik := idem
		prev, qerr := h.q.GetInsertCommonsReceivedPayloadByFarmIdempotency(ctx, db.GetInsertCommonsReceivedPayloadByFarmIdempotencyParams{
			FarmPseudonym:        farmPseudo,
			SourceIdempotencyKey: &ik,
		})
		if qerr == nil {
			if prev.PayloadHash != payloadHash {
				httputil.WriteError(w, http.StatusConflict, "idempotency key already stored with a different payload")
				return
			}
			h.writeIngestOK(w, true, prev.ID)
			return
		}
		if !errors.Is(qerr, pgx.ErrNoRows) {
			log.Printf("insert commons ingest: idempotency lookup: %v", qerr)
			httputil.WriteError(w, http.StatusInternalServerError, "failed to verify idempotency")
			return
		}
	}

	var idemPtr *string
	if idem != "" {
		ik := idem
		idemPtr = &ik
	}

	id, err := h.q.InsertInsertCommonsReceivedPayload(ctx, db.InsertInsertCommonsReceivedPayloadParams{
		PayloadHash:          payloadHash,
		FarmPseudonym:        farmPseudo,
		SchemaVersion:        insertcommonsschema.SchemaVersion,
		GeneratedAt:          genAt,
		Payload:              raw,
		SourceIdempotencyKey: idemPtr,
	})
	duplicate := false
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == pgerrUniqueViolation {
			if existing, e2 := h.q.GetInsertCommonsReceivedPayloadIDByHash(ctx, payloadHash); e2 == nil {
				duplicate = true
				id = existing
				err = nil
			} else if idem != "" {
				ik := idem
				row, e3 := h.q.GetInsertCommonsReceivedPayloadByFarmIdempotency(ctx, db.GetInsertCommonsReceivedPayloadByFarmIdempotencyParams{
					FarmPseudonym:        farmPseudo,
					SourceIdempotencyKey: &ik,
				})
				if e3 == nil {
					if row.PayloadHash != payloadHash {
						httputil.WriteError(w, http.StatusConflict, "idempotency key already stored with a different payload")
						return
					}
					duplicate = true
					id = row.ID
					err = nil
				}
			}
		}
		if err != nil {
			log.Printf("insert commons ingest: db: %v", err)
			httputil.WriteError(w, http.StatusInternalServerError, "failed to persist payload")
			return
		}
	}

	if h.retentionDays > 0 {
		cutoff := time.Now().UTC().AddDate(0, 0, -h.retentionDays)
		go func() {
			cctx, ccancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer ccancel()
			if err := h.q.DeleteInsertCommonsReceivedPayloadsBefore(cctx, cutoff); err != nil {
				log.Printf("insert commons receiver retention cleanup: %v", err)
			}
		}()
	}

	h.writeIngestOK(w, !duplicate, id)
}

func (h *Handler) writeIngestOK(w http.ResponseWriter, duplicate bool, id int64) {
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"accepted":   !duplicate,
		"duplicate":  duplicate,
		"storage_id": id,
		"schema":     insertcommonsschema.SchemaVersion,
	})
}
