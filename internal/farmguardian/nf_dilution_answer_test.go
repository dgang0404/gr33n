package farmguardian

import (
	"strings"
	"testing"
)

func TestEnsureNFDilutionRatiosInAnswer_jms(t *testing.T) {
	q := "What dilution should I use for JMS soil drench?"
	got := EnsureNFDilutionRatiosInAnswer("Use a mild microbial drench.", q)
	lower := strings.ToLower(got)
	if !strings.Contains(lower, "1:10") && !strings.Contains(lower, "1:20") {
		// Catalog may use either; at least one application dilution must appear.
		if !strings.Contains(lower, "catalog") || !strings.Contains(lower, ":") {
			t.Fatalf("expected catalog dilution ratios in answer: %q", got)
		}
	}
	// Already quoted — unchanged content path still ok.
	with := EnsureNFDilutionRatiosInAnswer("JMS soil is 1:20 and foliar 1:10.", q)
	if !strings.Contains(with, "1:20") {
		t.Fatalf("expected preserved ratios: %q", with)
	}
}

func TestEnsureNFDilutionRatiosInAnswer_skipsNonDilution(t *testing.T) {
	q := "How do I make JMS?"
	in := "Follow the JADAM microbial solution steps."
	if got := EnsureNFDilutionRatiosInAnswer(in, q); got != in {
		t.Fatalf("expected no rewrite for non-dilution Q, got %q", got)
	}
}
