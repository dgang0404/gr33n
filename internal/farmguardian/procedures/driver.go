package procedures

import (
	"fmt"
	"regexp"
	"strings"

	"gr33n-api/internal/farmguardian"
)

var (
	startRe = regexp.MustCompile(`(?i)^(?:start(?:\s+procedure)?|begin)\s+([a-z0-9][-a-z0-9]*)`)
	stopRe  = regexp.MustCompile(`(?i)^(?:stop|cancel|end)\s+(?:procedure|walkthrough)`)
)

// SuggestID maps informal operator phrases to a procedure id (Phase 37 WS7).
func SuggestID(message string) string {
	lower := strings.ToLower(strings.TrimSpace(message))
	switch {
	case strings.Contains(lower, "wire") && (strings.Contains(lower, "relay") || strings.Contains(lower, "light") || strings.Contains(lower, "pi")):
		return "wire-pi-relay-light"
	case strings.Contains(lower, "sensor") && (strings.Contains(lower, "no read") || strings.Contains(lower, "not read") || strings.Contains(lower, "nothing")):
		return "diagnose-sensor-no-reading"
	case strings.Contains(lower, "actuator") || strings.Contains(lower, "won't turn") || strings.Contains(lower, "wont turn") || strings.Contains(lower, "won't fire"):
		return "diagnose-actuator-wont-fire"
	case strings.Contains(lower, "pi offline") || strings.Contains(lower, "pi is offline") || strings.Contains(lower, "device offline"):
		return "diagnose-pi-offline"
	default:
		return ""
	}
}

// HandleTurn processes procedure control messages. When handled is true, answer is the full assistant text.
func HandleTurn(repoRoot, message string, meta SessionMeta) (handled bool, answer string, newMeta SessionMeta, payload *TurnPayload) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return false, "", meta, nil
	}

	if strings.Contains(strings.ToLower(msg), "list procedure") {
		summary, err := ListSummary(repoRoot)
		if err != nil {
			return true, "Field procedures are not loaded on this server (check docs/field-guides/procedures).", meta, nil
		}
		return true, summary, meta, nil
	}

	if stopRe.MatchString(msg) {
		if meta.Active != nil && meta.Active.Status == StatusActive {
			meta.Active.Status = StatusStopped
			return true, "Stopped the guided procedure. You can start again anytime with `start procedure <id>`.", meta, nil
		}
		return false, "", meta, nil
	}

	if m := startRe.FindStringSubmatch(msg); len(m) == 2 {
		return startProcedure(repoRoot, m[1], meta)
	}

	if id := SuggestID(msg); id != "" && meta.Active == nil {
		if strings.Contains(strings.ToLower(msg), "help") || strings.Contains(strings.ToLower(msg), "how") || strings.Contains(strings.ToLower(msg), "walk") {
			return startProcedure(repoRoot, id, meta)
		}
	}

	if meta.Active == nil || meta.Active.Status != StatusActive {
		return false, "", meta, nil
	}

	return advanceProcedure(repoRoot, msg, meta)
}

func startProcedure(repoRoot, id string, meta SessionMeta) (bool, string, SessionMeta, *TurnPayload) {
	p, err := Get(repoRoot, id)
	if err != nil {
		return true, fmt.Sprintf("I don't have a procedure called %q. Say `list procedures` or try: wire-pi-relay-light, diagnose-sensor-no-reading, diagnose-actuator-wont-fire, diagnose-pi-offline.", id), meta, nil
	}
	meta.Active = &ActiveState{ID: p.ID, StepN: 1, Status: StatusActive}
	if blocked, ans, pl := safetyBlockIfNeeded(p, meta.Active); blocked {
		return true, ans, meta, pl
	}
	ans, payload := presentStep(p, meta.Active)
	return true, ans, meta, payload
}

func safetyBlockIfNeeded(p Procedure, active *ActiveState) (blocked bool, answer string, payload *TurnPayload) {
	step := stepByN(p, active.StepN)
	tier := step.NormalizeStepTier()
	if tier != farmguardian.SafetyTierQualifiedPersonRequired && !step.StopUnlessQualified {
		return false, "", nil
	}
	active.Status = StatusSafetyStopped
	stop := farmguardian.SafetyStopForTier(tier)
	ans := fmt.Sprintf("**Step %d of %d — safety stop**\n\n%s\n\n%s", step.N, len(p.Steps), strings.TrimSpace(step.Say), stop)
	pl := payloadFor(p, active, step, true)
	pl.SafetyStopped = true
	return true, ans, pl
}

func advanceProcedure(repoRoot, msg string, meta SessionMeta) (bool, string, SessionMeta, *TurnPayload) {
	p, err := Get(repoRoot, meta.Active.ID)
	if err != nil {
		meta.Active = nil
		return true, "The active procedure is no longer available. Start again with `start procedure <id>`.", meta, nil
	}

	lower := strings.ToLower(msg)
	if strings.Contains(lower, "repeat") || strings.Contains(lower, "again") || strings.Contains(lower, "say step") {
		if blocked, ans, pl := safetyBlockIfNeeded(p, meta.Active); blocked {
			return true, ans, meta, pl
		}
		ans, payload := presentStep(p, meta.Active)
		return true, ans, meta, payload
	}
	if strings.Contains(lower, "help") && !isConfirm(msg) {
		step := stepByN(p, meta.Active.StepN)
		ans := fmt.Sprintf("**Help on step %d:** %s\n\nWhen ready, reply **done** or **yes** after you complete the step. Say **repeat** to hear it again, or **stop procedure** to exit.", meta.Active.StepN, step.Say)
		if r := strings.TrimSpace(step.Ref); r != "" {
			ans += fmt.Sprintf("\n\nMore detail: field-guides/%s", r)
		}
		return true, ans, meta, payloadFor(p, meta.Active, step, false)
	}
	if !isConfirm(msg) && !isFailed(msg) {
		return false, "", meta, nil
	}
	if isFailed(msg) {
		ans := fmt.Sprintf("Step %d didn't work — note what you see and ask for help, or say **stop procedure**. Common checks: power off, wire polarity, correct GPIO pin in gr33n.", meta.Active.StepN)
		return true, ans, meta, payloadFor(p, meta.Active, stepByN(p, meta.Active.StepN), false)
	}

	step := stepByN(p, meta.Active.StepN)
	tier := step.NormalizeStepTier()
	if tier == farmguardian.SafetyTierQualifiedPersonRequired || step.StopUnlessQualified {
		meta.Active.Status = StatusSafetyStopped
		stop := farmguardian.SafetyStopForTier(tier)
		ans := fmt.Sprintf("**Step %d — safety stop**\n\n%s\n\n%s", step.N, strings.TrimSpace(step.Say), stop)
		pl := payloadFor(p, meta.Active, step, true)
		pl.SafetyStopped = true
		return true, ans, meta, pl
	}

	if meta.Active.StepN >= len(p.Steps) {
		meta.Active.Status = StatusCompleted
		return true, fmt.Sprintf("**Procedure complete — %s**\n\nNice work. If you still need to register the actuator or device in gr33n, ask me and I'll open a change request you can Confirm.", p.Title), meta, &TurnPayload{
			ProcedureID: p.ID,
			Title:       p.Title,
			StepN:       meta.Active.StepN,
			StepTotal:   len(p.Steps),
			Status:      StatusCompleted,
			PrintPath:   printPath(p.ID),
		}
	}

	meta.Active.StepN++
	next := stepByN(p, meta.Active.StepN)
	tier = next.NormalizeStepTier()
	if tier == farmguardian.SafetyTierQualifiedPersonRequired || next.StopUnlessQualified {
		meta.Active.Status = StatusSafetyStopped
		stop := farmguardian.SafetyStopForTier(tier)
		ans := fmt.Sprintf("**Step %d of %d — safety stop**\n\n%s\n\n%s", next.N, len(p.Steps), strings.TrimSpace(next.Say), stop)
		pl := payloadFor(p, meta.Active, next, true)
		pl.SafetyStopped = true
		return true, ans, meta, pl
	}

	ans, payload := presentStep(p, meta.Active)
	return true, ans, meta, payload
}

func presentStep(p Procedure, active *ActiveState) (string, *TurnPayload) {
	step := stepByN(p, active.StepN)
	payload := payloadFor(p, active, step, false)
	tier := step.NormalizeStepTier()
	var b strings.Builder
	fmt.Fprintf(&b, "**%s** — step %d of %d", p.Title, step.N, len(p.Steps))
	if tier == farmguardian.SafetyTierCaution {
		b.WriteString(" · **caution**")
	}
	b.WriteString("\n\n")
	b.WriteString(strings.TrimSpace(step.Say))
	if c := strings.TrimSpace(step.Confirm); c != "" {
		fmt.Fprintf(&b, "\n\n**When done, confirm:** %s\n\nReply **done** or **yes** to continue, **help** for more detail, **repeat** to hear this again.", c)
	}
	if r := strings.TrimSpace(step.Ref); r != "" {
		fmt.Fprintf(&b, "\n\n_Citation: procedure#%d · field-guides/%s_", step.N, r)
	}
	fmt.Fprintf(&b, "\n\n_Print checklist:_ %s", printPath(p.ID))
	return b.String(), payload
}

func payloadFor(p Procedure, active *ActiveState, step Step, safetyStopped bool) *TurnPayload {
	return &TurnPayload{
		ProcedureID:   p.ID,
		Title:         p.Title,
		StepN:         step.N,
		StepTotal:     len(p.Steps),
		SafetyTier:    step.NormalizeStepTier(),
		Say:           step.Say,
		Confirm:       step.Confirm,
		Ref:           step.Ref,
		Status:        active.Status,
		SafetyStopped: safetyStopped,
		PrintPath:     printPath(p.ID),
	}
}

func printPath(id string) string {
	return "/v1/field-guides/procedures/" + id + "/print"
}

func stepByN(p Procedure, n int) Step {
	for _, s := range p.Steps {
		if s.N == n {
			return s
		}
	}
	if n >= 1 && n <= len(p.Steps) {
		return p.Steps[n-1]
	}
	return Step{N: n, Say: "Unknown step.", Confirm: "Continue?"}
}

func isConfirm(msg string) bool {
	lower := strings.ToLower(strings.TrimSpace(msg))
	for _, w := range []string{"done", "yes", "yep", "ok", "okay", "confirmed", "finished", "complete"} {
		if lower == w || strings.HasPrefix(lower, w+" ") {
			return true
		}
	}
	return false
}

func isFailed(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "didn't work") || strings.Contains(lower, "didnt work") || strings.Contains(lower, "failed") || strings.Contains(lower, "not working")
}

// ListSummary returns a short catalog for chat.
func ListSummary(repoRoot string) (string, error) {
	all, err := List(repoRoot)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("**Guided field procedures** (one step at a time):\n\n")
	for _, p := range all {
		fmt.Fprintf(&b, "- `%s` — %s\n", p.ID, p.Title)
	}
	b.WriteString("\nStart with: `start procedure wire-pi-relay-light`")
	return b.String(), nil
}
