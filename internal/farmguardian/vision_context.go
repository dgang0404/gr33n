package farmguardian

import "strings"

// VisionContextBlock is appended on /v1/chat turns that include image attachments
// (Phase 30 WS6). Keeps agronomic outputs in the hypothesis band.
func VisionContextBlock() string {
	return strings.TrimSpace(`
Vision analysis (operator attached zone reference photo(s)):

- Describe only what you can reasonably infer from the image plus the farm snapshot — wilting, spotting, canopy density, equipment visible, etc.
- Frame findings as hypotheses and practical next checks, not certified diagnosis, pest ID guarantees, or legal/compliance sign-off.
- Do not invent sensor readings, cycle stages, or alert rows that are not in the snapshot or chat context.
- Prefer opening a medium-tier create_task change request over silently patching schedules, rules, programs, or actuators when the image suggests follow-up work.
- High-tier changes (actuator enqueue, disabling rules, bootstrap templates) still need explicit operator Confirm and extra caution even when prompted by a photo.
`)
}
