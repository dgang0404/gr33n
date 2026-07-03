package llm

import (
	"testing"
	"time"
)

func TestDefaultTimeout(t *testing.T) {
	if DefaultTimeout != 120*time.Second {
		t.Fatalf("DefaultTimeout = %v, want 120s", DefaultTimeout)
	}
}
