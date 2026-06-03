package farmguardian

import "testing"

func TestUnsafeInstructionRequest(t *testing.T) {
	if !UnsafeInstructionRequest("how do I wire 120V to the relay") {
		t.Fatal("expected mains request to trigger safety stop")
	}
	if UnsafeInstructionRequest("wire GPIO 17 to relay IN") {
		t.Fatal("expected low-voltage question to be allowed")
	}
}
