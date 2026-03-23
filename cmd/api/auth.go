package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gr33n-api/internal/httputil"
)

// package-level vars set by main() before registerRoutes()
var (
	piAPIKey  string
	jwtSecret []byte
)

type contextKey string

const claimsKey contextKey = "claims"

// ── API Key middleware (Pi → API) ────────────────────────────────────────────
// Protects POST /sensors/{id}/readings and PATCH /devices/{id}/status.
// Pi sends:  X-API-Key: <PI_API_KEY>
func requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if piAPIKey == "" {
			// PI_API_KEY not configured — pass through (dev mode)
			next.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get("X-API-Key")
		if key == "" {
			httputil.WriteError(w, http.StatusUnauthorized, "X-API-Key required")
			return
		}
		if key != piAPIKey {
			httputil.WriteError(w, http.StatusForbidden, "invalid API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── JWT middleware (Dashboard → API) ─────────────────────────────────────────
// Protects all dashboard routes.
// Vue sends:  Authorization: Bearer <token>
func requireJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(jwtSecret) == 0 {
			// JWT_SECRET not configured — pass through (dev mode)
			next.ServeHTTP(w, r)
			return
		}

		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httputil.WriteError(w, http.StatusUnauthorized, "Authorization: Bearer <token> required")
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		}, jwt.WithExpirationRequired())

		if err != nil || !token.Valid {
			httputil.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, token.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// IssueToken is called by the login handler to mint a signed JWT.
// exp: duration until expiry (e.g. 24 * time.Hour)
func IssueToken(username string, exp time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(exp).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}
