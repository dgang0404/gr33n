package eval

import "strings"

// FilterFixturesByIDs returns fixtures whose ID is in ids (comma-separated). Empty ids returns all.
func FilterFixturesByIDs(fixtures []Question, ids string) []Question {
	want := parseIDSet(ids)
	if len(want) == 0 {
		return fixtures
	}
	out := make([]Question, 0, len(want))
	for _, q := range fixtures {
		if want[strings.TrimSpace(q.ID)] {
			out = append(out, q)
		}
	}
	return out
}

func parseIDSet(ids string) map[string]bool {
	ids = strings.TrimSpace(ids)
	if ids == "" {
		return nil
	}
	out := make(map[string]bool)
	for _, part := range strings.Split(ids, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out[part] = true
		}
	}
	return out
}
