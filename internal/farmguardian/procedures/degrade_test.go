package procedures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsFieldRelatedQuestion(t *testing.T) {
	if !IsFieldRelatedQuestion("help me wire the pi to a relay") {
		t.Fatal("expected field related")
	}
	if IsFieldRelatedQuestion("what is the capital of france") {
		t.Fatal("expected unrelated")
	}
}

func TestTryFieldDegrade_StartsSuggestedProcedure(t *testing.T) {
	root := findRepoRootDegrade(t)
	ans, _, pl, ok := TryFieldDegrade(root, "help me wire the pi to a light", SessionMeta{})
	if !ok || pl == nil || pl.ProcedureID != "wire-pi-relay-light" {
		t.Fatalf("ok=%v pl=%+v ans=%q", ok, pl, ans)
	}
	if !strings.Contains(strings.ToLower(ans), "offline") {
		t.Fatalf("expected degrade banner: %q", ans)
	}
}

func findRepoRootDegrade(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "docs", "field-guides", "procedures")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("repo root not found")
	return ""
}
