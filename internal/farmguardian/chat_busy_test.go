package farmguardian

import "testing"

func TestGroundedChatBusyLock(t *testing.T) {
	if !TryAcquireGroundedChat() {
		t.Fatal("expected acquire")
	}
	if TryAcquireGroundedChat() {
		t.Fatal("expected second acquire to fail")
	}
	if !GroundedChatBusy() {
		t.Fatal("expected busy")
	}
	ReleaseGroundedChat()
	if GroundedChatBusy() {
		t.Fatal("expected not busy after release")
	}
	if !TryAcquireGroundedChat() {
		t.Fatal("expected re-acquire")
	}
	ReleaseGroundedChat()
}

func TestEarlySSEEnabled_defaultOn(t *testing.T) {
	t.Setenv("GUARDIAN_EARLY_SSE", "")
	if !EarlySSEEnabled() {
		t.Fatal("expected default on")
	}
	t.Setenv("GUARDIAN_EARLY_SSE", "0")
	if EarlySSEEnabled() {
		t.Fatal("expected off when 0")
	}
}
