package insertcommonsreceiver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
)

// SchemaVersion must match the farm-side sender (internal/handler/farm/insert_commons.go).
const SchemaVersion = "gr33n.insert_commons.v1"

const maxBodyBytes = 1 << 20

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

// ServeHTTP handles GET /health and POST /v1/ingest only.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/health":
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "gr33n-insert-commons-receiver"})
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

func (h *Handler) ingest(w http.ResponseWriter, r *http.Request) {
	if !h.checkAuth(w, r) {
		return
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))), "application/json") {
		httputil.WriteError(w, http.StatusBadRequest, "Content-Type must be application/json")
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

	farmPseudo, genAt, err := validatePayload(raw)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	sum := sha256.Sum256(raw)
	payloadHash := hex.EncodeToString(sum[:])

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	id, err := h.q.InsertInsertCommonsReceivedPayload(ctx, db.InsertInsertCommonsReceivedPayloadParams{
		PayloadHash:   payloadHash,
		FarmPseudonym: farmPseudo,
		SchemaVersion: SchemaVersion,
		GeneratedAt:   genAt,
		Payload:       raw,
	})
	duplicate := false
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			duplicate = true
			id, err = h.q.GetInsertCommonsReceivedPayloadIDByHash(ctx, payloadHash)
			if errors.Is(err, pgx.ErrNoRows) {
				err = errors.New("duplicate without stored row")
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

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"accepted":   !duplicate,
		"duplicate":  duplicate,
		"storage_id": id,
		"schema":     SchemaVersion,
	})
}

func validatePayload(raw []byte) (farmPseudo string, genAt time.Time, err error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", time.Time{}, fmt.Errorf("invalid json: %w", err)
	}
	required := []string{"schema_version", "generated_at", "farm_pseudonym", "farm_profile", "aggregates", "privacy"}
	for _, k := range required {
		if _, ok := root[k]; !ok {
			return "", time.Time{}, fmt.Errorf("missing required field: %s", k)
		}
	}
	var ver string
	if err := json.Unmarshal(root["schema_version"], &ver); err != nil {
		return "", time.Time{}, errors.New("invalid schema_version")
	}
	if strings.TrimSpace(ver) != SchemaVersion {
		return "", time.Time{}, fmt.Errorf("unsupported schema_version %q (expected %s)", ver, SchemaVersion)
	}
	var genStr string
	if err := json.Unmarshal(root["generated_at"], &genStr); err != nil {
		return "", time.Time{}, errors.New("invalid generated_at")
	}
	genStr = strings.TrimSpace(genStr)
	var parseErr error
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		genAt, parseErr = time.Parse(layout, genStr)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		return "", time.Time{}, errors.New("generated_at must be RFC3339 or RFC3339Nano")
	}
	if genStr == "" {
		return "", time.Time{}, errors.New("generated_at is empty")
	}
	now := time.Now().UTC()
	if genAt.After(now.Add(10 * time.Minute)) {
		return "", time.Time{}, errors.New("generated_at is too far in the future")
	}
	if genAt.Before(now.Add(-365 * 24 * time.Hour)) {
		return "", time.Time{}, errors.New("generated_at is too old")
	}

	if err := json.Unmarshal(root["farm_pseudonym"], &farmPseudo); err != nil {
		return "", time.Time{}, errors.New("invalid farm_pseudonym")
	}
	farmPseudo = strings.TrimSpace(farmPseudo)
	if farmPseudo == "" {
		return "", time.Time{}, errors.New("farm_pseudonym is required")
	}

	var fp map[string]any
	if err := json.Unmarshal(root["farm_profile"], &fp); err != nil {
		return "", time.Time{}, errors.New("farm_profile must be an object")
	}
	for _, k := range []string{"scale_tier", "timezone_bucket", "currency", "operational_status"} {
		if _, ok := fp[k]; !ok {
			return "", time.Time{}, fmt.Errorf("farm_profile missing %q", k)
		}
	}

	var agg map[string]any
	if err := json.Unmarshal(root["aggregates"], &agg); err != nil {
		return "", time.Time{}, errors.New("aggregates must be an object")
	}
	for _, k := range []string{"costs", "tasks", "devices"} {
		v, ok := agg[k]
		if !ok {
			return "", time.Time{}, fmt.Errorf("aggregates missing %q", k)
		}
		if _, isObj := v.(map[string]any); !isObj {
			return "", time.Time{}, fmt.Errorf("aggregates.%s must be an object", k)
		}
	}

	var priv map[string]any
	if err := json.Unmarshal(root["privacy"], &priv); err != nil {
		return "", time.Time{}, errors.New("privacy must be an object")
	}
	if _, ok := priv["includes_pii"]; !ok {
		return "", time.Time{}, errors.New("privacy.includes_pii is required")
	}

	return farmPseudo, genAt.UTC(), nil
}
