package authhandler

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"gr33n-api/internal/authctx"
	"gr33n-api/internal/authsecurity"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
)

const defaultInviteTTL = 7 * 24 * time.Hour

func (h *Handler) registrationMode() authsecurity.RegistrationMode {
	if h.regMode == "" {
		return authsecurity.RegistrationInvite
	}
	return h.regMode
}

func generateInviteCode() (string, error) {
	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "gr33n-" + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b[:])), nil
}

// CreateInvite handles POST /auth/invites (JWT required).
func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	if h.registrationMode() == authsecurity.RegistrationClosed {
		httputil.WriteError(w, http.StatusForbidden, "registration is closed on this server")
		return
	}
	userID, ok := authctx.UserID(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "login required")
		return
	}
	var body struct {
		TTLHours *int `json:"ttl_hours"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	ttl := defaultInviteTTL
	if body.TTLHours != nil && *body.TTLHours > 0 && *body.TTLHours <= 24*30 {
		ttl = time.Duration(*body.TTLHours) * time.Hour
	}

	code, err := generateInviteCode()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "could not generate invite code")
		return
	}
	q := db.New(h.pool)
	createdBy := pgtype.UUID{Bytes: userID, Valid: true}
	row, err := q.CreateRegistrationInvite(r.Context(), db.CreateRegistrationInviteParams{
		Code:      code,
		CreatedBy: createdBy,
		ExpiresAt: time.Now().Add(ttl),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"code":       row.Code,
		"expires_at": row.ExpiresAt,
		"id":         row.ID.String(),
	})
}

// ListInvites handles GET /auth/invites (JWT required).
func (h *Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	q := db.New(h.pool)
	rows, err := q.ListActiveRegistrationInvites(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		item := map[string]any{
			"id":         row.ID.String(),
			"code":       row.Code,
			"expires_at": row.ExpiresAt,
			"created_at": row.CreatedAt,
		}
		if row.CreatedBy.Valid {
			item["created_by"] = uuid.UUID(row.CreatedBy.Bytes).String()
		}
		out = append(out, item)
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"invites": out})
}

func (h *Handler) validateInviteForRegister(ctx context.Context, code string) (*db.AuthRegistrationInvite, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("invite_code required")
	}
	q := db.New(h.pool)
	row, err := q.GetRegistrationInviteByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid or expired invite code")
		}
		return nil, err
	}
	if row.UsedAt.Valid {
		return nil, errors.New("invite code already used")
	}
	if time.Now().After(row.ExpiresAt) {
		return nil, errors.New("invalid or expired invite code")
	}
	return &row, nil
}

func (h *Handler) consumeInvite(ctx context.Context, invite *db.AuthRegistrationInvite, userID uuid.UUID) {
	if invite == nil {
		return
	}
	q := db.New(h.pool)
	_ = q.MarkRegistrationInviteUsed(ctx, db.MarkRegistrationInviteUsedParams{
		ID:     invite.ID,
		UsedBy: pgtype.UUID{Bytes: userID, Valid: true},
	})
}
