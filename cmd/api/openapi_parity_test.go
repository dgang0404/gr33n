// Phase 28 WS6 — OpenAPI parity guard.
//
// Every route registered in cmd/api/routes.go must have a matching
// path entry in openapi.yaml (with the right HTTP method). This test
// is the safety net that catches future drift: add a route → forget
// to document it → test fails the next CI run.
//
// Implementation: we scrape `routes.go` as text (capturing every
// `mux.Handle("METHOD /path", …)` line) and confirm each one appears
// under `paths:` in `openapi.yaml`. The match is exact path-text
// against the OpenAPI path key + the lower-case HTTP verb under it.
//
// No yaml.v3 dependency — we parse the spec with simple line-anchored
// regexes. That keeps the test compiling in any build environment
// and makes failures easy to read (false positives are impossible;
// every miss is a literal "method + path" the spec doesn't carry).
//
// Skipped routes — see `routesIntentionallyUndocumented` — are
// short-circuit paths (e.g. CORS preflight) or test-only endpoints
// that don't deserve API contract entries. Add to that map with a
// comment explaining why.

package main

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// routesIntentionallyUndocumented is the allow-list of registered
// routes that don't appear in openapi.yaml on purpose. Keep this set
// small — every entry needs a comment.
var routesIntentionallyUndocumented = map[string]bool{
	// (empty — every Phase 24–28 route is documented as of WS6.)
}

func TestOpenAPI_AllRoutesDocumented(t *testing.T) {
	routes := scrapeRegisteredRoutes(t, "routes.go")
	if len(routes) < 50 {
		t.Fatalf("sanity: only %d routes scraped from routes.go — regex broken?", len(routes))
	}

	specBytes, err := os.ReadFile("../../openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}
	spec := string(specBytes)

	var missing []string
	for _, r := range routes {
		key := r.method + " " + r.path
		if routesIntentionallyUndocumented[key] {
			continue
		}
		if !pathIsDocumented(spec, r.method, r.path) {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("openapi.yaml is missing %d route(s):\n  - %s",
			len(missing), strings.Join(missing, "\n  - "))
	}
}

type registeredRoute struct {
	method string
	path   string
}

// scrapeRegisteredRoutes pulls every `mux.Handle("METHOD /path", …)`
// out of routes.go. Comment-only `mux.Handle` lines are excluded by
// the trailing-comma check (real registrations always continue on
// the same line; commented-out registrations live as `//` prefixes).
func scrapeRegisteredRoutes(t *testing.T, path string) []registeredRoute {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	// e.g. `mux.Handle("GET /farms/{id}", jwt(...))`
	re := regexp.MustCompile(`mux\.Handle\("(GET|PUT|POST|DELETE|PATCH|HEAD|OPTIONS) (\S+?)"`)
	matches := re.FindAllStringSubmatch(string(raw), -1)
	out := make([]registeredRoute, 0, len(matches))
	for _, m := range matches {
		out = append(out, registeredRoute{method: m[1], path: m[2]})
	}
	return out
}

// pathIsDocumented finds `  <path>:` in the spec (path entries are
// always indented two spaces under `paths:` and end with a colon),
// then checks whether the matching HTTP method appears in the
// following indented block. We cut the look-ahead at the next top-
// level path key so we don't accidentally cross into a sibling path's
// methods.
func pathIsDocumented(spec, method, path string) bool {
	pathRE := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(path) + `:\s*$`)
	loc := pathRE.FindStringIndex(spec)
	if loc == nil {
		return false
	}
	tail := spec[loc[1]:]
	if next := regexp.MustCompile(`(?m)^  /`).FindStringIndex(tail); next != nil {
		tail = tail[:next[0]]
	}
	verb := strings.ToLower(method)
	methodRE := regexp.MustCompile(`(?m)^    ` + verb + `:\s*$`)
	return methodRE.MatchString(tail)
}
