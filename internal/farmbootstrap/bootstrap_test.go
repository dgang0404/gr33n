package farmbootstrap

import "testing"

func TestRequestedTemplate(t *testing.T) {
	t.Parallel()
	v, ok := RequestedTemplate(nil)
	if ok || v != "" {
		t.Fatalf("nil pointer: ok=%v v=%q", ok, v)
	}
	blank := ""
	v, ok = RequestedTemplate(&blank)
	if ok {
		t.Fatal("empty string should not count as requested")
	}
	s := " jadam_indoor_photoperiod_v1 "
	v, ok = RequestedTemplate(&s)
	if !ok || v != "jadam_indoor_photoperiod_v1" {
		t.Fatalf("got %q ok=%v", v, ok)
	}
}

func TestIsBlankChoice(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"", "none", "NONE", " blank ", "Blank"} {
		if !IsBlankChoice(s) {
			t.Fatalf("%q should be blank choice", s)
		}
	}
	if IsBlankChoice("jadam_indoor_photoperiod_v1") {
		t.Fatal("template key should not be blank")
	}
}

func TestIsKnownTemplate(t *testing.T) {
	if !IsKnownTemplate(JadamIndoorPhotoperiodV1) {
		t.Fatal("expected seeded template to be known")
	}
	if IsKnownTemplate("not_a_real_template_v999") {
		t.Fatal("unknown template should return false")
	}
}
