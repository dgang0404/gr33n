package authhandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

type IssueTokenFunc func(username string, exp time.Duration, extra map[string]any) (string, error)

type Handler struct {
	mu               sync.RWMutex
	adminUsername    string
	passwordHash     []byte
	hashFilePath     string
	issueToken       IssueTokenFunc
	pool             *pgxpool.Pool
	adminBindUserID  uuid.UUID // JWT user_id for env-admin login (farm RBAC requires it)
	adminBindEmail   string    // optional email claim for env-admin (matches seeded profile)
}

func NewHandler(adminUsername string, passwordHash []byte, hashFilePath string, issueToken IssueTokenFunc, pool *pgxpool.Pool, adminBindUserID uuid.UUID, adminBindEmail string) *Handler {
	return &Handler{
		adminUsername:   adminUsername,
		passwordHash:    passwordHash,
		hashFilePath:    hashFilePath,
		issueToken:      issueToken,
		pool:            pool,
		adminBindUserID: adminBindUserID,
		adminBindEmail:  adminBindEmail,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	const tokenExp = 24 * time.Hour

	// Try DB-backed login first
	if h.pool != nil {
		q := db.New(h.pool)
		email := body.Username
		authUser, err := q.GetAuthUserByEmail(r.Context(), &email)
		if err == nil && authUser.PasswordHash != nil {
			if err := bcrypt.CompareHashAndPassword(authUser.PasswordHash, []byte(body.Password)); err == nil {
				token, err := h.issueToken(body.Username, tokenExp, map[string]any{
					"user_id": authUser.ID.String(),
					"email":   email,
				})
				if err != nil {
					httputil.WriteError(w, http.StatusInternalServerError, "could not issue token")
					return
				}
				httputil.WriteJSON(w, http.StatusOK, map[string]any{
					"token":      token,
					"expires_in": int(tokenExp.Seconds()),
					"user_id":    authUser.ID.String(),
				})
				return
			}
			httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
	}

	// Fallback: env-admin login
	if body.Username != h.adminUsername {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	h.mu.RLock()
	hash := h.passwordHash
	h.mu.RUnlock()
	if hash == nil {
		httputil.WriteError(w, http.StatusUnauthorized, "no password configured")
		return
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte(body.Password)); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	extra := map[string]any{}
	if h.adminBindUserID != uuid.Nil {
		extra["user_id"] = h.adminBindUserID.String()
		if h.adminBindEmail != "" {
			extra["email"] = h.adminBindEmail
		}
	}
	token, err := h.issueToken(body.Username, tokenExp, extra)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	out := map[string]any{"token": token, "expires_in": int(tokenExp.Seconds())}
	if h.adminBindUserID != uuid.Nil {
		out["user_id"] = h.adminBindUserID.String()
	}
	httputil.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Email == "" || body.Password == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(body.Password) < 8 {
		httputil.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	q := db.New(h.pool)

	// Check if user already exists (invite flow: null password_hash)
	existing, err := q.GetAuthUserByEmail(r.Context(), &body.Email)
	if err == nil {
		if existing.PasswordHash != nil {
			httputil.WriteError(w, http.StatusConflict, "user already exists")
			return
		}
		// Invited user setting password for the first time
		if err := q.UpdateAuthUserPasswordHash(r.Context(), db.UpdateAuthUserPasswordHashParams{
			ID:           existing.ID,
			PasswordHash: hash,
		}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		const tokenExp = 24 * time.Hour
		token, err := h.issueToken(body.Email, tokenExp, map[string]any{
			"user_id": existing.ID.String(),
			"email":   body.Email,
		})
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "could not issue token")
			return
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"token":      token,
			"expires_in": int(tokenExp.Seconds()),
			"user_id":    existing.ID.String(),
		})
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// New user registration
	authUser, err := q.CreateAuthUser(r.Context(), db.CreateAuthUserParams{
		Email:        &body.Email,
		PasswordHash: hash,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create user: "+err.Error())
		return
	}

	fullName := body.FullName
	if fullName == "" {
		fullName = body.Email
	}
	_, err = q.CreateProfile(r.Context(), db.CreateProfileParams{
		UserID:      authUser.ID,
		FullName:    &fullName,
		Email:       body.Email,
		Role:        commontypes.UserRoleEnum("user"),
		Preferences: []byte("{}"),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create profile: "+err.Error())
		return
	}

	const tokenExp = 24 * time.Hour
	token, err := h.issueToken(body.Email, tokenExp, map[string]any{
		"user_id": authUser.ID.String(),
		"email":   body.Email,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"token":      token,
		"expires_in": int(tokenExp.Seconds()),
		"user_id":    authUser.ID.String(),
	})
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if len(body.NewPassword) < 8 {
		httputil.WriteError(w, http.StatusBadRequest, "new password must be at least 8 characters")
		return
	}
	h.mu.RLock()
	currentHash := h.passwordHash
	h.mu.RUnlock()
	if currentHash == nil {
		httputil.WriteError(w, http.StatusBadRequest, "no password configured for env-admin")
		return
	}
	if err := bcrypt.CompareHashAndPassword(currentHash, []byte(body.CurrentPassword)); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 12)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	if h.hashFilePath != "" {
		_ = os.MkdirAll(filepath.Dir(h.hashFilePath), 0700)
		if err := os.WriteFile(h.hashFilePath, newHash, 0600); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to persist hash")
			return
		}
	}
	h.mu.Lock()
	h.passwordHash = newHash
	h.mu.Unlock()
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
}
