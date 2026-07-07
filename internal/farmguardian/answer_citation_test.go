package farmguardian

import (
	"strings"
	"testing"
)

const smokeMorningWalkURLs = `Check [source#2](https://gr33n.com/sources/field_guide) and [task#5](https://gr33n.com/tasks) plus ([live farm snapshot](#)).`

func TestSanitizeCitationURLs_smokeMorningWalk(t *testing.T) {
	t.Parallel()
	got, meta := SanitizeCitationURLs(smokeMorningWalkURLs)
	if !meta.Sanitized || meta.LinksRewritten != 3 {
		t.Fatalf("meta=%+v", meta)
	}
	if strings.Contains(got, "gr33n.com") {
		t.Fatalf("gr33n.com still present: %q", got)
	}
	if !strings.Contains(got, "source#2") || !strings.Contains(got, "task#5") || !strings.Contains(got, "live farm snapshot") {
		t.Fatalf("labels lost: %q", got)
	}
	if AnswerContainsFakeCitationURL(got) {
		t.Fatal("sanitized answer should not contain fake URLs")
	}
}

func TestSanitizeCitationURLs_keepsRealURLs(t *testing.T) {
	t.Parallel()
	in := "See [Ollama docs](https://ollama.com/docs) for setup."
	got, meta := SanitizeCitationURLs(in)
	if meta.Sanitized || got != in {
		t.Fatalf("meta=%+v got=%q", meta, got)
	}
}

func TestAnswerContainsFakeCitationURL(t *testing.T) {
	t.Parallel()
	if !AnswerContainsFakeCitationURL(smokeMorningWalkURLs) {
		t.Fatal("expected fake URLs detected")
	}
	clean, _ := SanitizeCitationURLs(smokeMorningWalkURLs)
	if AnswerContainsFakeCitationURL(clean) {
		t.Fatal("expected clean after sanitize")
	}
}
