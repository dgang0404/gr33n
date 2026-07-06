package chat

import "testing"

func TestPhaseStatus_fields(t *testing.T) {
	m := phaseStatus("snapshot", "Reading live farm…")
	if m["phase"] != "snapshot" || m["message"] != "Reading live farm…" {
		t.Fatalf("%v", m)
	}
}

func TestPhaseStatus_orderDocumented(t *testing.T) {
	phases := []string{"preparing", "snapshot", "read_tools", "embedding", "generating", "awakening"}
	for _, p := range phases {
		m := phaseStatus(p, "x")
		if m["phase"] != p {
			t.Fatalf("phase %q", p)
		}
	}
}
