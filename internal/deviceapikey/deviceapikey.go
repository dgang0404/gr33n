// Package deviceapikey formats and verifies per-device Pi credentials (Phase 57).
package deviceapikey

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const prefix = "gdev_"

// Format builds the plaintext key shown once to operators: gdev_{deviceID}_{secret}.
func Format(deviceID int64, secret string) string {
	return fmt.Sprintf("%s%d_%s", prefix, deviceID, secret)
}

// Parse extracts device ID and secret from a plaintext device key.
func Parse(raw string) (deviceID int64, secret string, ok bool) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, prefix) {
		return 0, "", false
	}
	rest := strings.TrimPrefix(raw, prefix)
	parts := strings.SplitN(rest, "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, "", false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, "", false
	}
	return id, parts[1], true
}

// NewSecret returns a URL-safe random secret for embedding in gdev_{deviceID}_{secret}.
func NewSecret() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Hash returns a bcrypt hash of the full plaintext device key.
func Hash(plaintext string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// Verify compares a plaintext key against a stored bcrypt hash.
func Verify(plaintext, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext)) == nil
}

// ExtractFromRequest reads X-Device-Key or Authorization: Device <key>.
func ExtractFromRequest(headerDeviceKey, authHeader string) string {
	if s := strings.TrimSpace(headerDeviceKey); s != "" {
		return s
	}
	authHeader = strings.TrimSpace(authHeader)
	if strings.HasPrefix(strings.ToLower(authHeader), "device ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}
