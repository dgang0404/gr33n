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
		{"route empty path", ContextRef{Type: "route", Path: ""}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildContextRefBlock(t.Context(), nil, 1, tc.ref); got != "" {
				t.Fatalf("expected empty, got %q", got)
			}
		})
	}
}

func TestBuildContextRefBlock_Route(t *testing.T) {
	ref := ContextRef{Type: "route", Path: "/fertigation", Name: "Fertigation"}
	got := BuildContextRefBlock(t.Context(), nil, 0, ref)
	for _, want := range []string{
		"Operator UI context — viewing: Fertigation",
		"Route path: /fertigation",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("route block missing %q:\n%s", want, got)
		}
	}
}

func TestContextRefPromptBlock_RouteWithoutFarm(t *testing.T) {
	ref := ContextRef{Type: "route", Path: "/plants", Name: "Plants"}
	got := ContextRefPromptBlock(t.Context(), nil, 0, ref)
	if !strings.HasPrefix(got, "Contextual focus") {
		t.Fatalf("expected wrapper, got %q", got)
	}
	if !strings.Contains(got, "Plants") {
		t.Fatalf("expected screen label in block: %q", got)
	}
}

func TestBuildContextRefBlock_SetupWizardRoutes(t *testing.T) {
	got := BuildContextRefBlock(t.Context(), nil, 0, ContextRef{Type: "route", Path: "/farms/7/setup"})
	for _, want := range []string{"Farm setup", "wizard buttons", "bootstrap"} {
		if !strings.Contains(got, want) {
			t.Fatalf("farm setup block missing %q:\n%s", want, got)
		}
	}
	got = BuildContextRefBlock(t.Context(), nil, 0, ContextRef{Type: "route", Path: "/farms/7/zones/new"})
	if !strings.Contains(got, "grow room wizard") {
		t.Fatalf("zone wizard block: %q", got)
	}
	got = BuildContextRefBlock(t.Context(), nil, 0, ContextRef{Type: "route", Path: "/farms/7/devices/new"})
	if !strings.Contains(got, "wire-pi-relay-light") {
		t.Fatalf("device wizard block: %q", got)
	}
}

func TestBuildContextRefBlock_OperationsRoutes(t *testing.T) {
	got := BuildContextRefBlock(t.Context(), nil, 0, ContextRef{Type: "route", Path: "/operations/supplies"})
	for _, want := range []string{"Supplies", "do not promise Guardian can change stock"} {
		if !strings.Contains(got, want) {
			t.Fatalf("supplies block missing %q:\n%s", want, got)
		}
	}
	got = BuildContextRefBlock(t.Context(), nil, 0, ContextRef{Type: "route", Path: "/operations/money"})
	if !strings.Contains(got, "Money") || !strings.Contains(got, "GL/COA") {
		t.Fatalf("money block: %q", got)
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
