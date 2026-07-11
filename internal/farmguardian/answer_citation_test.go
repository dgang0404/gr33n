package farmguardian

import (
	"strings"
	"testing"
)

const smokeMorningWalkURLs = `Check [source#2](https://gr33n.com/sources/field_guide) and [task#5](https://gr33n.com/tasks) plus ([live farm snapshot](#)).`

const smokeMorningWalkGr33nDocs = `See [task #5](https://gr33n-docs/phase_40_unified_farmer_ux_zone_cockpit.plan.md#tasks) and [platform_doc #2](https://gr33n-docs/local-operator-bootstrap.md#14-confirmed-actions).`

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

const smokeRelativePlanLink = `See [zone cockpit](phase_40_unified_farmer_ux_zone_cockpit.plan.md#tasks) and [bootstrap](local-operator-bootstrap.md#14-confirmed-actions).`

func TestSanitizeCitationURLs_gr33ncoreSensorAlerts(t *testing.T) {
	t.Parallel()
	in := `1. [Humidity High](https://gr33ncore.sensor_alerts/unread): reading 72.4% RH.`
	got, meta := SanitizeCitationURLs(in)
	if !meta.Sanitized || meta.LinksRewritten != 1 {
		t.Fatalf("meta=%+v", meta)
	}
	if strings.Contains(got, "gr33ncore") || strings.Contains(got, "sensor_alerts") {
		t.Fatalf("url still present: %q", got)
	}
	if !strings.Contains(got, "Humidity High") {
		t.Fatalf("label lost: %q", got)
	}
}

func TestAnswerContainsFakeCitationURL_gr33ncore(t *testing.T) {
	t.Parallel()
	in := `[OHN low](https://gr33ncore.sensor_alerts/unread)`
	if !AnswerContainsFakeCitationURL(in) {
		t.Fatal("expected gr33ncore link flagged as fake")
	}
}

func TestSanitizeCitationURLs_relativePlanLinks(t *testing.T) {
	t.Parallel()
	got, meta := SanitizeCitationURLs(smokeRelativePlanLink)
	if !meta.Sanitized || meta.LinksRewritten != 2 {
		t.Fatalf("meta=%+v", meta)
	}
	if strings.Contains(got, ".plan.md") || strings.Contains(got, ".md#") {
		t.Fatalf("relative md links still present: %q", got)
	}
	if !strings.Contains(got, "zone cockpit") || !strings.Contains(got, "bootstrap") {
		t.Fatalf("labels lost: %q", got)
	}
}

func TestSanitizeCitationURLs_gr33nDocs(t *testing.T) {
	t.Parallel()
	got, meta := SanitizeCitationURLs(smokeMorningWalkGr33nDocs)
	if !meta.Sanitized || meta.LinksRewritten != 2 {
		t.Fatalf("meta=%+v", meta)
	}
	if strings.Contains(got, "gr33n-docs") {
		t.Fatalf("gr33n-docs still present: %q", got)
	}
	if !strings.Contains(got, "task #5") || !strings.Contains(got, "platform_doc #2") {
		t.Fatalf("labels lost: %q", got)
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
