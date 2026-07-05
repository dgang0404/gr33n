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

func TestIsEmbeddingModel(t *testing.T) {
	t.Setenv("EMBEDDING_MODEL", "rjmalagon/gte-qwen2-1.5b-instruct-embed-f16")
	if !IsEmbeddingModel("rjmalagon/gte-qwen2-1.5b-instruct-embed-f16:latest", nil) {
		t.Fatal("embed in name should match")
	}
	if IsEmbeddingModel("phi3:mini", nil) {
		t.Fatal("phi3 should not be embed")
	}
	if IsSelectableChatModel(ModelInfo{Name: "rjmalagon/gte-qwen2-1.5b-instruct-embed-f16:latest"}) {
		t.Fatal("embed model should not be selectable for chat")
	}
	if !IsSelectableChatModel(ModelInfo{Name: "phi3:mini", Capabilities: []string{"completion"}}) {
		t.Fatal("phi3 should be selectable")
	}
}
