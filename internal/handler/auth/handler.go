package authhandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"gr33n-api/internal/auditlog"
	"gr33n-api/internal/authctx"
	"gr33n-api/internal/authsecurity"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/httputil"
	"gr33n-api/internal/platform/commontypes"
)

type IssueTokenFunc func(username string, exp time.Duration, extra map[string]any) (string, error)

type Handler struct {
	mu              sync.RWMutex
	adminUsername   string
	passwordHash    []byte
	hashFilePath    string
	issueToken      IssueTokenFunc
	pool            *pgxpool.Pool
	adminBindUserID uuid.UUID
	adminBindEmail  string
	regMode         authsecurity.RegistrationMode
	loginLimiter    *authsecurity.LoginLimiter
}

func NewHandler(adminUsername string, passwordHash []byte, hashFilePath string, issueToken IssueTokenFunc, pool *pgxpool.Pool, adminBindUserID uuid.UUID, adminBindEmail string, regMode authsecurity.RegistrationMode, loginLimiter *authsecurity.LoginLimiter) *Handler {
	if loginLimiter == nil {
		loginLimiter = authsecurity.NewLoginLimiter(authsecurity.LoginMaxPerMinuteFromEnv())
	}
	return &Handler{
		adminUsername:   adminUsername,
		passwordHash:    passwordHash,
		hashFilePath:    hashFilePath,
		issueToken:      issueToken,
		pool:            pool,
		adminBindUserID: adminBindUserID,
		adminBindEmail:  adminBindEmail,
		regMode:         regMode,
		loginLimiter:    loginLimiter,
	}
}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return xff
	}
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i >= 0 {
		return host[:i]
	}
	return host
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
	username := strings.TrimSpace(body.Username)
	ip := clientIP(r)
	if !h.loginLimiter.Allow(ip, strings.ToLower(username)) {
		retry := h.loginLimiter.RetryAfter(ip, strings.ToLower(username))
		w.Header().Set("Retry-After", retryAfterSeconds(retry))
		auditlog.Submit(r.Context(), db.New(h.pool), r, auditlog.Event{
			Action: db.Gr33ncoreUserActionTypeEnumLoginFailure,
			Status: "failure",
			Details: map[string]any{
				"reason":   "rate_limited",
				"username": username,
				"ip":       ip,
			},
		})
		httputil.WriteError(w, http.StatusTooManyRequests, "too many login attempts; try again later")
		return
	}

	const tokenExp = 24 * time.Hour

	if h.pool != nil {
		q := db.New(h.pool)
		email := username
		authUser, err := q.GetAuthUserByEmail(r.Context(), &email)
		if err == nil && authUser.PasswordHash != nil {
			if err := bcrypt.CompareHashAndPassword(authUser.PasswordHash, []byte(body.Password)); err == nil {
				token, err := h.issueToken(username, tokenExp, map[string]any{
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

	if username != h.adminUsername {
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
	token, err := h.issueToken(username, tokenExp, extra)
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

func retryAfterSeconds(d time.Duration) string {
	sec := int(d.Seconds())
	if sec < 1 {
		sec = 1
	}
	return strconv.Itoa(sec)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	mode := h.registrationMode()
	if mode == authsecurity.RegistrationClosed {
		httputil.WriteError(w, http.StatusForbidden, "registration is closed; contact your farm operator for access")
		return
	}

	var body struct {
		Email      string `json:"email"`
		Password   string `json:"password"`
		FullName   string `json:"full_name"`
		InviteCode string `json:"invite_code"`
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

	existing, err := q.GetAuthUserByEmail(r.Context(), &body.Email)
	if err == nil {
		if existing.PasswordHash != nil {
			httputil.WriteError(w, http.StatusConflict, "user already exists")
			return
		}
		const tokenExp = 24 * time.Hour
		if err := q.UpdateAuthUserPasswordHash(r.Context(), db.UpdateAuthUserPasswordHashParams{
			ID:           existing.ID,
			PasswordHash: hash,
		}); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
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

	var invite *db.AuthRegistrationInvite
	if mode == authsecurity.RegistrationInvite {
		invite, err = h.validateInviteForRegister(r.Context(), body.InviteCode)
		if err != nil {
			httputil.WriteError(w, http.StatusForbidden, err.Error())
			return
		}
	}

	authUser, err := q.CreateAuthUser(r.Context(), db.CreateAuthUserParams{
		Email:        &body.Email,
		PasswordHash: hash,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create user: "+err.Error())
		return
	}
	h.consumeInvite(r.Context(), invite, authUser.ID)

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

	if userID, ok := authctx.UserID(r.Context()); ok && h.pool != nil {
		q := db.New(h.pool)
		email := authctx.Email(r.Context())
		if email == "" {
			httputil.WriteError(w, http.StatusBadRequest, "email claim required")
			return
		}
		authUser, err := q.GetAuthUserByEmail(r.Context(), &email)
		if err == nil && authUser.PasswordHash != nil {
			if err := bcrypt.CompareHashAndPassword(authUser.PasswordHash, []byte(body.CurrentPassword)); err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "current password is incorrect")
				return
			}
			newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 12)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "failed to hash password")
				return
			}
			if err := q.UpdateAuthUserPasswordHash(r.Context(), db.UpdateAuthUserPasswordHashParams{
				ID:           authUser.ID,
				PasswordHash: newHash,
			}); err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, err.Error())
				return
			}
			auditlog.Submit(r.Context(), q, r, auditlog.Event{
				Action: db.Gr33ncoreUserActionTypeEnumChangeSetting,
				Status: "success",
				Details: map[string]any{
					"setting": "password",
					"user_id": userID.String(),
				},
			})
			httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
			return
		}
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
