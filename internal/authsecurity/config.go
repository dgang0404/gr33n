package authsecurity

import (
	"os"
	"strconv"
	"strings"
)

// RegistrationMode controls who may create new accounts via POST /auth/register.
type RegistrationMode string

const (
	RegistrationOpen   RegistrationMode = "open"
	RegistrationInvite RegistrationMode = "invite"
	RegistrationClosed RegistrationMode = "closed"
)

// RegistrationModeFromEnv reads REGISTRATION_MODE (open|invite|closed).
// Default invite for production-style installs; open when AUTH_MODE is dev or auth_test
// so local smoke tests keep working without extra env.
func RegistrationModeFromEnv(authMode string) RegistrationMode {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("REGISTRATION_MODE")))
	switch raw {
	case "open", "invite", "closed":
		return RegistrationMode(raw)
	case "":
		switch strings.ToLower(strings.TrimSpace(authMode)) {
		case "dev", "auth_test":
			return RegistrationOpen
		default:
			return RegistrationInvite
		}
	default:
		return RegistrationInvite
	}
}

// LoginMaxPerMinuteFromEnv returns AUTH_LOGIN_MAX_PER_MINUTE (default 10).
func LoginMaxPerMinuteFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("AUTH_LOGIN_MAX_PER_MINUTE"))
	if raw == "" {
		return 10
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 10
	}
	if n > 1000 {
		return 1000
	}
	return n
}

// LegacyPiKeyDisabled reports PI_LEGACY_KEY_DISABLED=true.
func LegacyPiKeyDisabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("PI_LEGACY_KEY_DISABLED")), "true")
}
