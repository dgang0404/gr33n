package eval

import "testing"

func TestFilterFixturesByIDs_emptyReturnsAll(t *testing.T) {
	all := SmokeFixtures()
	got := FilterFixturesByIDs(all, "")
	if len(got) != len(all) {
		t.Fatalf("got %d want %d", len(got), len(all))
	}
}

func TestFilterFixturesByIDs_single(t *testing.T) {
	got := FilterFixturesByIDs(SmokeFixtures(), "smoke-ec-ph")
	if len(got) != 1 || got[0].ID != "smoke-ec-ph" {
		t.Fatalf("got %+v", got)
	}
}

func TestFilterFixturesByIDs_multiple(t *testing.T) {
	got := FilterFixturesByIDs(SmokeFixtures(), " smoke-morning-walk , smoke-ec-ph ")
	if len(got) != 2 {
		t.Fatalf("got %d", len(got))
	}
	if got[0].ID != "smoke-morning-walk" || got[1].ID != "smoke-ec-ph" {
		t.Fatalf("order/ids wrong: %+v", got)
	}
}

func TestFilterFixturesByIDs_unknownReturnsEmpty(t *testing.T) {
	if len(FilterFixturesByIDs(SmokeFixtures(), "nope")) != 0 {
		t.Fatal("expected empty")
	}
}
