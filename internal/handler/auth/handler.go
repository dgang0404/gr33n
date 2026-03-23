package authhandler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gr33n-api/internal/httputil"
)

type Handler struct {
	mu            sync.RWMutex
	adminUsername string
	passwordHash  []byte
	hashFilePath  string
	issueToken    func(username string, exp time.Duration) (string, error)
}

func NewHandler(adminUsername string, passwordHash []byte, hashFilePath string, issueToken func(string, time.Duration) (string, error)) *Handler {
	return &Handler{adminUsername: adminUsername, passwordHash: passwordHash, hashFilePath: hashFilePath, issueToken: issueToken}
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
	if body.Username != h.adminUsername {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	h.mu.RLock()
	hash := h.passwordHash
	h.mu.RUnlock()
	if err := bcrypt.CompareHashAndPassword(hash, []byte(body.Password)); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	const tokenExp = 24 * time.Hour
	token, err := h.issueToken(body.Username, tokenExp)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"token": token, "expires_in": int(tokenExp.Seconds())})
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
