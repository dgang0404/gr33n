package eval

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"gr33n-api/internal/farmguardian"
)

// ScenarioTurn is one message in a multi-turn Guardian dialogue smoke test.
type ScenarioTurn struct {
	Prompt string
}

// Scenario is a multi-turn write-intent flow: propose, optionally refine via
// follow-up turns, then either confirm via API or leave pending for UI review.
type Scenario struct {
	ID              string
	Category        string
	Turns           []ScenarioTurn
	ExpectProposal  bool
	Grounded        bool
	Model           string
	VerifyFixtureID string  // maps to write-feed / write-task confirm verification
	ConfirmFinal    bool    // API confirm + DB verify after all turns
	LeavePending    bool    // bump TTL so Pending tab stays populated
	MinRevision      int     // optional: require pending proposal revision >= N
	WantVolumeLiters float64 // optional feed revise check (0.3 after correction)
	WantTitle        string  // optional create_task title after revise
	RequireTaskZone  bool    // optional: require zone_id on final create_task proposal
	WantDueDate            string  // optional create_task due_date (YYYY-MM-DD) after revise
	WantDueDateOffsetDays  int     // optional: due_date = UTC today + N days (for relative revise)
}

// IsScenarioSuite reports whether the eval suite runs multi-turn scenarios.
func IsScenarioSuite(suite string) bool {
	switch strings.ToLower(strings.TrimSpace(suite)) {
	case "change-requests-ui", "change_requests_ui", "pr-ui", "pr_ui",
		"change-requests-ui-quick", "change_requests_ui_quick", "pr-ui-quick":
		return true
	default:
		return false
	}
}

// RunScenarioSuite executes multi-turn scenarios sequentially (one scenario at a time).
func RunScenarioSuite(ctx context.Context, api *APIClient, model string, scenarios []Scenario, opts RunSuiteOptions) ([]ScoreResult, error) {
	var out []ScoreResult
	groundedWarmed := false
	for i, sc := range scenarios {
		log.Printf("eval: [%d/%d] starting scenario %q (%d turns, confirm=%v leave_pending=%v)",
			i+1, len(scenarios), sc.ID, len(sc.Turns), sc.ConfirmFinal, sc.LeavePending)
		res, err := runOneScenario(ctx, api, model, sc, opts, &groundedWarmed)
		if err != nil {
			return out, fmt.Errorf("%s: %w", sc.ID, err)
		}
		status := "pass"
		if !res.Passed {
			status = "fail"
		}
		log.Printf("eval: [%d/%d] scenario %q %s (proposals=%d)",
			i+1, len(scenarios), sc.ID, status, res.ProposalCount)
		out = append(out, res)
	}
	return out, nil
}

func runOneScenario(ctx context.Context, api *APIClient, model string, sc Scenario, opts RunSuiteOptions, groundedWarmed *bool) (ScoreResult, error) {
	if len(sc.Turns) == 0 {
		return scoreResultFromScenarioError(sc, model, fmt.Errorf("scenario has no turns")), nil
	}
	m := model
	if strings.TrimSpace(sc.Model) != "" {
		m = strings.TrimSpace(sc.Model)
	}
	if sc.Grounded && opts.WarmupGrounded && groundedWarmed != nil && !*groundedWarmed {
		*groundedWarmed = true
		warm := func() error {
			return api.WarmupFarmCounsel(ctx, m, opts.WarmupTimeout)
		}
		var err error
		if opts.WarmupAsync {
			go func() { _ = warm() }()
		} else {
			err = warm()
		}
		if err != nil {
			if opts.RequireWarmup {
				return scoreResultFromScenarioError(sc, model, fmt.Errorf("guardian warmup required: %w", err)), nil
			}
			log.Printf("eval: warmup before grounded block: %v (continuing)", err)
		} else if !opts.WarmupAsync {
			log.Printf("eval: counsel model ready before grounded block")
		}
	}

	sessionID := uuid.New().String()
	var lastIn ScoreInput
	var lastTurnIdx int
	var sessionProposalIDs []string
	for ti, turn := range sc.Turns {
		q := scenarioTurnQuestion(sc, ti, turn)
		in, sid, err := api.RunQuestionInSession(ctx, m, q, sessionID)
		if err != nil {
			return scoreResultFromScenarioError(sc, m, err), nil
		}
		if sid = strings.TrimSpace(sid); sid != "" {
			sessionID = sid
		}
		lastIn = in
		lastTurnIdx = ti + 1
		sessionProposalIDs = append(sessionProposalIDs, in.ProposalIDs...)
		log.Printf("eval: scenario %q turn %d/%d done in %.1fs (proposals=%d session=%s)",
			sc.ID, ti+1, len(sc.Turns), in.Latency.Seconds(), in.ProposalCount, truncate(sessionID, 8))
		if len(in.ProposalIDs) > 0 && (len(sc.Turns) > 1 || sc.LeavePending) {
			extendScenarioProposalTTL(ctx, sc.ID, ti+1, in.ProposalIDs, opts)
		}
	}

	scoreQ := scenarioScoreQuestion(sc, sc.Turns[len(sc.Turns)-1].Prompt)
	lastIn.Question = scoreQ
	if sc.ExpectProposal {
		enrichProposalFromPending(ctx, api, sessionID, &lastIn, sc, sessionProposalIDs)
	}
	res := scoreScenario(lastIn, sc)
	enrichScoreResult(&res, lastIn, m)
	res.ID = sc.ID
	res.Category = sc.Category
	res.Prompt = formatScenarioPrompt(sc)

	if !res.Passed {
		res.Notes = fmt.Sprintf("turn %d/%d: %s", lastTurnIdx, len(sc.Turns), res.Notes)
		return res, nil
	}
	if !sc.ExpectProposal || len(res.ProposalIDs) == 0 {
		return res, nil
	}

	propID, prop, err := resolveScenarioProposal(ctx, api, sessionID, res.ProposalIDs, sc)
	if err != nil {
		res.Passed = false
		res.Notes = err.Error()
		return res, nil
	}
	res.ProposalIDs = []string{propID}

	if err := VerifyPendingProposalIDs(ctx, api, []string{propID}); err != nil {
		res.Passed = false
		res.Notes = "pending queue: " + err.Error()
		return res, nil
	}

	if sc.ConfirmFinal {
		fixtureID := strings.TrimSpace(sc.VerifyFixtureID)
		if fixtureID == "" {
			fixtureID = sc.ID
		}
		if err := ConfirmAndVerifyProposal(ctx, api, fixtureID, propID); err != nil {
			res.Passed = false
			res.Notes = err.Error()
			return res, nil
		}
		res.Notes = fmt.Sprintf("confirmed proposal %s (tool=%s rev=%d)", propID, prop.Tool, prop.Revision)
		return res, nil
	}

	if sc.LeavePending {
		ttl := opts.LeavePendingTTL
		if ttl <= 0 {
			ttl = LeavePendingTTLFromEnv()
		}
		n, err := BumpProposalExpiry(ctx, []string{propID}, ttl)
		if err != nil {
			res.Passed = false
			res.Notes = "bump pending TTL: " + err.Error()
			return res, nil
		}
		res.Notes = fmt.Sprintf("left pending for UI: %s (tool=%s rev=%d, %d row(s), expires ~%s)",
			propID, prop.Tool, prop.Revision, n, time.Now().UTC().Add(ttl).Format(time.RFC3339))
	}
	return res, nil
}

func scenarioTurnQuestion(sc Scenario, turnIdx int, turn ScenarioTurn) Question {
	return Question{
		ID:             fmt.Sprintf("%s-turn-%d", sc.ID, turnIdx+1),
		Category:       sc.Category,
		Prompt:         turn.Prompt,
		ExpectProposal: sc.ExpectProposal && turnIdx == len(sc.Turns)-1,
		Grounded:       sc.Grounded,
		Model:          sc.Model,
	}
}

func scenarioScoreQuestion(sc Scenario, lastPrompt string) Question {
	return Question{
		ID:             sc.ID,
		Category:       sc.Category,
		Prompt:         lastPrompt,
		ExpectProposal: sc.ExpectProposal,
		Grounded:       sc.Grounded,
		Model:          sc.Model,
	}
}

func formatScenarioPrompt(sc Scenario) string {
	var parts []string
	for i, t := range sc.Turns {
		parts = append(parts, fmt.Sprintf("[%d] %s", i+1, t.Prompt))
	}
	return strings.Join(parts, " → ")
}

func scoreResultFromScenarioError(sc Scenario, model string, err error) ScoreResult {
	return ScoreResult{
		ID:       sc.ID,
		Category: sc.Category,
		Passed:   false,
		Notes:    err.Error(),
		Error:    err.Error(),
		Prompt:   formatScenarioPrompt(sc),
		Grounded: sc.Grounded,
		Model:    model,
	}
}

func enrichProposalFromPending(ctx context.Context, api *APIClient, sessionID string, in *ScoreInput, sc Scenario, turnProposalIDs []string) {
	if in == nil || in.ProposalCount > 0 {
		return
	}
	propID, _, err := resolveScenarioProposal(ctx, api, sessionID, append(turnProposalIDs, in.ProposalIDs...), Scenario{ID: sc.ID})
	if err != nil {
		return
	}
	in.ProposalCount = 1
	in.ProposalIDs = []string{propID}
}

// scoreScenario scores a multi-turn scenario. Last-turn answers may be dialogue-only
// (no inline proposal) while the write-intent row from an earlier turn stays pending.
func scoreScenario(in ScoreInput, sc Scenario) ScoreResult {
	res := Score(in)
	if len(sc.Turns) <= 1 || !sc.ExpectProposal || in.ProposalCount == 0 {
		return res
	}
	if !res.Passed && res.Notes == "expected valid proposal" {
		res.Passed = true
		res.Notes = "multi-turn: proposal from session pending queue (last turn may be dialogue only)"
	}
	return res
}

func extendScenarioProposalTTL(ctx context.Context, scenarioID string, turn int, proposalIDs []string, opts RunSuiteOptions) {
	ttl := opts.LeavePendingTTL
	if ttl <= 0 {
		ttl = LeavePendingTTLFromEnv()
	}
	n, err := BumpProposalExpiry(ctx, proposalIDs, ttl)
	if err != nil {
		log.Printf("eval: scenario %q turn %d TTL bump: %v (continuing)", scenarioID, turn, err)
		return
	}
	log.Printf("eval: scenario %q turn %d extended proposal TTL (%d row(s), expires ~%s)",
		scenarioID, turn, n, time.Now().UTC().Add(ttl).Format(time.RFC3339))
}

func resolveScenarioProposal(ctx context.Context, api *APIClient, sessionID string, responseIDs []string, sc Scenario) (string, PendingProposal, error) {
	pending, err := api.FetchPendingProposals(ctx)
	if err != nil {
		return "", PendingProposal{}, err
	}
	sessionID = strings.TrimSpace(sessionID)
	var candidates []PendingProposal
	for _, p := range pending {
		if sessionID != "" && strings.TrimSpace(p.SessionID) == sessionID {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		for _, id := range responseIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			for _, p := range pending {
				if strings.TrimSpace(p.ProposalID) == id {
					candidates = append(candidates, p)
				}
			}
		}
	}
	if len(candidates) == 0 {
		return "", PendingProposal{}, fmt.Errorf("no pending proposal for session %s", truncate(sessionID, 8))
	}
	best := candidates[0]
	for _, p := range candidates[1:] {
		if p.Revision > best.Revision {
			best = p
		}
	}
	if sc.MinRevision > 0 && best.Revision < sc.MinRevision {
		return "", PendingProposal{}, fmt.Errorf("proposal rev %d want >= %d", best.Revision, sc.MinRevision)
	}
	if sc.WantVolumeLiters > 0 {
		got, err := float64FromAny(best.Args["total_volume_liters"])
		if err != nil {
			return "", PendingProposal{}, fmt.Errorf("proposal args missing total_volume_liters")
		}
		if math.Abs(got-sc.WantVolumeLiters) > 0.05 {
			return "", PendingProposal{}, fmt.Errorf("proposal volume %.3f want ~%.3f L", got, sc.WantVolumeLiters)
		}
	}
	if wantTitle := strings.TrimSpace(sc.WantTitle); wantTitle != "" {
		got, err := stringFromAny(best.Args["title"])
		if err != nil {
			return "", PendingProposal{}, fmt.Errorf("proposal args missing title")
		}
		if !strings.EqualFold(got, wantTitle) {
			return "", PendingProposal{}, fmt.Errorf("proposal title %q want %q", got, wantTitle)
		}
	}
	if sc.RequireTaskZone {
		zid, err := int64FromAny(best.Args["zone_id"])
		if err != nil || zid <= 0 {
			return "", PendingProposal{}, fmt.Errorf("proposal args missing zone_id")
		}
	}
	if wantDue := strings.TrimSpace(sc.WantDueDate); wantDue != "" || sc.WantDueDateOffsetDays > 0 {
		got, err := stringFromAny(best.Args["due_date"])
		if err != nil {
			return "", PendingProposal{}, fmt.Errorf("proposal args missing due_date")
		}
		want := wantDue
		if sc.WantDueDateOffsetDays > 0 {
			want = time.Now().UTC().AddDate(0, 0, sc.WantDueDateOffsetDays).Format("2006-01-02")
		}
		if got != want {
			return "", PendingProposal{}, fmt.Errorf("proposal due_date %q want %q", got, want)
		}
	}
	return strings.TrimSpace(best.ProposalID), best, nil
}

// ToEvalQuestionScoresFromScenarios maps scenario runner results for JSON persistence.
func ToEvalQuestionScoresFromScenarios(scores []ScoreResult) []farmguardian.EvalQuestionScore {
	return ToEvalQuestionScores(scores)
}

// PassedScenarioProposalIDs collects proposal_ids from passed leave-pending scenarios.
func PassedScenarioProposalIDs(scenarios []Scenario, scores []farmguardian.EvalQuestionScore) []string {
	leaveByID := make(map[string]bool, len(scenarios))
	for _, sc := range scenarios {
		if sc.LeavePending && sc.ExpectProposal {
			leaveByID[sc.ID] = true
		}
	}
	var out []string
	for _, s := range scores {
		if !leaveByID[s.ID] || !s.Passed {
			continue
		}
		for _, id := range s.ProposalIDs {
			if id = strings.TrimSpace(id); id != "" {
				out = append(out, id)
			}
		}
	}
	return out
}
