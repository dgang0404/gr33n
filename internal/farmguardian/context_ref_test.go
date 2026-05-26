package farmguardian

import (
	"strings"
	"testing"
)

func TestBuildContextRefBlock_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		ref  ContextRef
	}{
		{"zero id", ContextRef{Type: "alert", ID: 0}},
		{"unknown type", ContextRef{Type: "task", ID: 1}},
		{"empty type", ContextRef{Type: "", ID: 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildContextRefBlock(t.Context(), nil, 1, tc.ref); got != "" {
				t.Fatalf("expected empty, got %q", got)
			}
		})
	}
}

func TestContextRefPromptBlock_WrapsBody(t *testing.T) {
	ref := ContextRef{Type: "zone", ID: 1, Name: "Flower Room"}
	body := BuildContextRefBlock(t.Context(), nil, 1, ref)
	if body != "" {
		t.Fatal("expected empty body without db")
	}

	const sample = "Operator focus — zone #3: Flower Room"
	got := "Contextual focus (background — do not cite as [n]):\n" + sample
	want := ContextRefPromptBlock(t.Context(), nil, 0, ref)
	if want != "" {
		t.Fatalf("expected empty without farm id, got %q", want)
	}
	if !strings.HasPrefix(got, "Contextual focus") {
		t.Fatal("wrapper prefix missing")
	}
}
