package authsecurity

import "testing"

func TestRegistrationModeFromEnv(t *testing.T) {
	t.Setenv("REGISTRATION_MODE", "")
	if got := RegistrationModeFromEnv("production"); got != RegistrationInvite {
		t.Fatalf("prod default want invite, got %q", got)
	}
	if got := RegistrationModeFromEnv("auth_test"); got != RegistrationOpen {
		t.Fatalf("auth_test default want open, got %q", got)
	}
	t.Setenv("REGISTRATION_MODE", "closed")
	if got := RegistrationModeFromEnv("auth_test"); got != RegistrationClosed {
		t.Fatalf("explicit closed want closed, got %q", got)
	}
}

func TestLoginLimiter(t *testing.T) {
	l := NewLoginLimiter(3)
	ip, user := "127.0.0.1", "dev@example.com"
	for i := 0; i < 3; i++ {
		if !l.Allow(ip, user) {
			t.Fatalf("attempt %d should allow", i+1)
		}
	}
	if l.Allow(ip, user) {
		t.Fatal("4th attempt should block")
	}
}
