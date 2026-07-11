package eval

// ChangeRequestFixtures returns a small set of write-intent prompts whose
// purpose is to land a pending row in Guardian's change-request queue
// (gr33ncore.guardian_action_proposals — what the UI shows as "PR queue"/
// proposal cards to Confirm). This is a subset of Fixtures()'s write_intent
// questions, reused here so `-suite change-requests` stays a short, fast
// script instead of running the full ~24-prompt regression set.
func ChangeRequestFixtures() []Question {
	all := Fixtures()
	out := make([]Question, 0, 4)
	for _, q := range all {
		if q.Category == "write_intent" {
			out = append(out, q)
		}
	}
	return out
}
