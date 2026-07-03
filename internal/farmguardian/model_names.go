package farmguardian

import "strings"

// NormalizeModelName strips a trailing :latest tag for lookup-only comparisons.
// Stored audit values and API responses keep the canonical Ollama name.
func NormalizeModelName(name string) string {
	name = strings.TrimSpace(name)
	return strings.TrimSuffix(name, ":latest")
}

// modelLookupKeys returns name variants tried when resolving against the cache.
func modelLookupKeys(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	seen := make(map[string]struct{}, 4)
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
	}
	add(name)
	bare := NormalizeModelName(name)
	add(bare)
	if bare != name {
		add(bare + ":latest")
	} else {
		add(name + ":latest")
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}
