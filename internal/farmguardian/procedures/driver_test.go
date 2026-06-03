package procedures

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHandleTurn_StartAndAdvance(t *testing.T) {
	root := repoRootFromTest(t)
	ResetForTest()
	meta := SessionMeta{}
	handled, ans, newMeta, pl := HandleTurn(root, "start procedure wire-pi-relay-light", meta)
	if !handled || pl == nil || pl.StepN != 1 {
		t.Fatalf("start: handled=%v pl=%+v ans=%q", handled, pl, ans)
	}
	if newMeta.Active == nil || newMeta.Active.ID != "wire-pi-relay-light" {
		t.Fatalf("meta: %+v", newMeta.Active)
	}

	handled, ans, newMeta, pl = HandleTurn(root, "done", newMeta)
	if !handled || pl == nil || pl.StepN != 2 {
		t.Fatalf("advance: handled=%v step=%d ans=%q", handled, pl.StepN, ans)
	}

	// step 2 -> 3 is qualified_person_required
	handled, _, newMeta, pl = HandleTurn(root, "yes", newMeta)
	if !handled || pl == nil || !pl.SafetyStopped {
		t.Fatalf("safety stop: %+v status=%s", pl, newMeta.Active.Status)
	}
	if newMeta.Active.Status != StatusSafetyStopped {
		t.Fatalf("status: %s", newMeta.Active.Status)
	}
}

func repoRootFromTest(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
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
