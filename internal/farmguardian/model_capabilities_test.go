package farmguardian

import "testing"

func TestIsChatCapable(t *testing.T) {
	if !IsChatCapable(nil) {
		t.Fatal("empty capabilities should default to chat-capable")
	}
	if !IsChatCapable([]string{"completion"}) {
		t.Fatal("completion should be chat-capable")
	}
	if !IsChatCapable([]string{"completion", "vision"}) {
		t.Fatal("vision+completion should be chat-capable")
	}
	if IsChatCapable([]string{"embedding"}) {
		t.Fatal("embedding-only should be excluded")
	}
}
