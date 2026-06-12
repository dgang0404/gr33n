package bootstraptemplates

import "testing"

func TestEmbeddedCatalog_HasJadamTemplate(t *testing.T) {
	c := embeddedCatalog()
	if !c.IsValid("jadam_indoor_photoperiod_v1") {
		t.Fatal("expected jadam template")
	}
	if c.IsValid("not_a_template") {
		t.Fatal("unexpected valid key")
	}
	list := c.List()
	if len(list) < 5 {
		t.Fatalf("expected >=5 templates, got %d", len(list))
	}
}
