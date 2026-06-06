package farmguardian

import "strings"

// SetupModeActive returns true when the setup-mode persona should attach to a
// grounded chat turn (Phase 44 WS4).
func SetupModeActive(snap Snapshot, explicit bool) bool {
	return explicit || snap.ZoneCount == 0
}

// SetupModePromptBlock is appended to the grounded system prompt when setup
// mode is active. It steers operators toward wizards and existing matchers.
func SetupModePromptBlock(snap Snapshot) string {
	var b strings.Builder
	b.WriteString("Farm setup mode (operator is onboarding this farm):\n")
	b.WriteString("- Prefer the in-app wizards (farm setup, add grow room, connect edge device, comfort targets) over inventing configuration in chat.\n")
	if snap.ZoneCount == 0 {
		b.WriteString("- This farm has no grow rooms yet. Guide the operator through the checklist order: add a grow room → connect edge device → set comfort targets → turn on one schedule.\n")
	}
	b.WriteString("- For a first grow in an empty room, you may propose apply_grow_setup_pack only when Phase 32 matcher rules pass (zone resolved, no active cycle, no duplicate plant) — never auto-Confirm.\n")
	b.WriteString("- apply_bootstrap_template is not chat-first: direct farm admins to Farm setup wizard or Settings with preview; do not promise chat can apply unless they are farm admin and explicitly request it.\n")
	b.WriteString("- Pi wiring and offline devices: cite field procedures (start procedure wire-pi-relay-light, diagnose-pi-offline) and print URLs from field_guide/platform_doc when available.\n")
	b.WriteString("- Do not insert proposals without the operator sending a message.")
	return strings.TrimSpace(b.String())
}
