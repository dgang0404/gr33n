package farmguardian

import "testing"

func TestPsEntry_found(t *testing.T) {
	loaded := map[string]ollamaPsModel{
		"phi3:mini": {Name: "phi3:mini", SizeVRAM: 0},
	}
	found, cpu := psEntry(loaded, "phi3:mini")
	if !found || !cpu {
		t.Fatalf("found=%v cpu=%v", found, cpu)
	}
}

func TestPsEntry_missing(t *testing.T) {
	loaded := map[string]ollamaPsModel{}
	if found, _ := psEntry(loaded, "tinyllama"); found {
		t.Fatal("expected missing")
	}
}

func TestEmbedModelFromEnv(t *testing.T) {
	t.Setenv("EMBEDDING_MODEL", "gte-test")
	if got := EmbedModelFromEnv(); got != "gte-test" {
		t.Fatalf("got %q", got)
	}
}

func TestVisionModelFromEnv(t *testing.T) {
	t.Setenv("LLM_VISION_MODEL", "llava")
	if got := VisionModelFromEnv(); got != "llava" {
		t.Fatalf("got %q", got)
	}
}
