package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/farmguardian/eval"
)

func main() {
	apiURL := flag.String("api", envOr("GUARDIAN_EVAL_API", "http://127.0.0.1:8080"), "gr33n API base URL")
	token := flag.String("token", os.Getenv("GUARDIAN_EVAL_TOKEN"), "JWT bearer token (or set GUARDIAN_EVAL_TOKEN)")
	farmID := flag.Int64("farm-id", 1, "demo farm id for grounded questions")
	modelsFlag := flag.String("models", "all", "comma-separated model names or 'all'")
	manualFlag := flag.Bool("manual", false, "print UI checklist for -suite and exit")
	suiteFlag := flag.String("suite", envOr("GUARDIAN_EVAL_SUITE", "regression"), "smoke | phase127 | regression | all")
	promptIDsFlag := flag.String("prompt-ids", envOr("GUARDIAN_EVAL_PROMPT_IDS", ""), "comma-separated fixture IDs to run (subset of suite)")
	reportPath := flag.String("report", farmguardian.DefaultEvalReportPath(), "output JSON report path")
	qaArchive := flag.String("qa-archive", "", "optional full QA run JSON path (default data/guardian_qa_runs/…)")
	llmBase := flag.String("llama-url", os.Getenv("LLM_BASE_URL"), "Ollama OpenAI base (for model discovery when models=all)")
	failOnRegression := flag.Bool("fail-on-regression", false, "exit non-zero if any fixture fails its heuristic, instead of always exiting 0")
	checkPendingProposals := flag.Bool("check-pending-proposals", false, "after each passed write-intent prompt, verify its proposal_id is in the pending queue (and skip the end-of-run batch check — proposals expire after 5m while prompts take 20+ min)")
	confirmProposals := flag.Bool("confirm-proposals", false, "with -check-pending-proposals: confirm each passed proposal immediately after its prompt and verify DB side effects (requires -check-pending-proposals)")
	leavePending := flag.Bool("leave-pending", false, "after each passed write-intent prompt, bump proposal TTL in DB so Pending tab stays populated for UI review (implies -check-pending-proposals; uses DATABASE_URL)")
	flag.Parse()

	if *leavePending {
		*checkPendingProposals = true
	}

	if *manualFlag {
		eval.PrintManualChecklist(*suiteFlag)
		return
	}

	if strings.TrimSpace(*token) == "" {
		log.Fatal("JWT required: pass -token or set GUARDIAN_EVAL_TOKEN (use make dev-auth-test login token)")
	}

	suite := strings.ToLower(strings.TrimSpace(*suiteFlag))
	scenarioSuite := eval.IsScenarioSuite(suite)

	var fixtures []eval.Question
	var scenarios []eval.Scenario
	if scenarioSuite {
		scenarios = eval.ScenariosForSuite(suite)
		scenarios = eval.FilterScenariosByIDs(scenarios, *promptIDsFlag)
		if len(scenarios) == 0 {
			if strings.TrimSpace(*promptIDsFlag) != "" {
				log.Fatalf("no scenarios for suite %q prompt-ids %q", suite, *promptIDsFlag)
			}
			log.Fatalf("no scenarios for suite %q", suite)
		}
	} else {
		fixtures = eval.FixturesForSuite(suite)
		fixtures = eval.FilterFixturesByIDs(fixtures, *promptIDsFlag)
		if len(fixtures) == 0 {
			if strings.TrimSpace(*promptIDsFlag) != "" {
				log.Fatalf("no fixtures for suite %q prompt-ids %q", suite, *promptIDsFlag)
			}
			log.Fatalf("no fixtures for suite %q", suite)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Hour)
	defer cancel()

	modelNames, err := resolveModels(ctx, *modelsFlag, *llmBase)
	if err != nil {
		log.Fatal(err)
	}
	if len(modelNames) == 0 {
		log.Fatal("no chat-capable models to evaluate")
	}

	client := eval.NewAPIClient(*apiURL, *token, *farmID)

	runOpts := eval.RunSuiteOptions{
		WarmupGrounded: suite == "smoke" || suite == "phase127" || suite == "phase128" || suite == "p128" || suite == "change-requests" || suite == "change_requests" || suite == "proposals" || suite == "pr" || scenarioSuite,
		WarmupTimeout:  eval.WarmupTimeoutFromEnv(),
		WarmupAsync:    suite == "smoke" || suite == "phase127" || suite == "change-requests" || suite == "change_requests" || suite == "proposals" || suite == "pr" || scenarioSuite,
		LogPath:        strings.TrimSpace(os.Getenv("GUARDIAN_EVAL_LOG")),
		CheckPendingPerPrompt: *checkPendingProposals,
		ConfirmPerPrompt:      *confirmProposals,
		LeavePending:          *leavePending,
		LeavePendingTTL:       eval.LeavePendingTTLFromEnv(),
	}

	rep := farmguardian.EvalReport{
		Models:  map[string]farmguardian.EvalSummary{},
		Details: map[string][]farmguardian.EvalQuestionScore{},
	}
	expectedProposals := 0
	var requiredProposalIDs []string
	failed := false
	for _, model := range modelNames {
		if scenarioSuite {
			log.Printf("evaluating model %q suite=%s (%d scenarios)…", model, suite, len(scenarios))
			scores, runErr := eval.RunScenarioSuite(ctx, client, model, scenarios, runOpts)
			if runErr != nil {
				log.Printf("eval scenario suite error: %v", runErr)
				failed = true
			}
			rep.Models[normalizeModelKey(model)] = eval.BuildReport(model, scores, *reportPath)
			details := eval.ToEvalQuestionScores(scores)
			rep.Details[normalizeModelKey(model)] = details
			printModelSummary(model, rep.Models[normalizeModelKey(model)])
			requiredProposalIDs = append(requiredProposalIDs, eval.PassedScenarioProposalIDs(scenarios, details)...)
			if archive := qaArchivePath(*qaArchive, suite, model); archive != "" {
				if err := farmguardian.SaveQARunArchive(archive, suite, model, details); err != nil {
					log.Printf("qa archive %q: %v", archive, err)
				} else {
					fmt.Printf("  QA archive: %s\n", archive)
				}
			}
			continue
		}
		log.Printf("evaluating model %q suite=%s (%d questions)…", model, suite, len(fixtures))
		scores, runErr := eval.RunSuite(ctx, client, model, fixtures, runOpts)
		if runErr != nil {
			log.Printf("eval suite error: %v", runErr)
			failed = true
		}
		rep.Models[normalizeModelKey(model)] = eval.BuildReport(model, scores, *reportPath)
		details := eval.ToEvalQuestionScores(scores)
		rep.Details[normalizeModelKey(model)] = details
		printModelSummary(model, rep.Models[normalizeModelKey(model)])
		expectedProposals += passedProposalFixtures(fixtures, details)
		requiredProposalIDs = append(requiredProposalIDs, passedProposalIDs(fixtures, details)...)
		if archive := qaArchivePath(*qaArchive, suite, model); archive != "" {
			if err := farmguardian.SaveQARunArchive(archive, suite, model, details); err != nil {
				log.Printf("qa archive %q: %v", archive, err)
			} else {
				fmt.Printf("  QA archive: %s\n", archive)
			}
		}
	}

	if err := farmguardian.SaveEvalReport(*reportPath, rep); err != nil {
		log.Fatal(err)
	}
	farmguardian.RefreshEvalCache()
	fmt.Printf("\nEval report written to %s\n", *reportPath)

	if *failOnRegression {
		if regressions := regressionFailures(rep.Details); len(regressions) > 0 {
			fmt.Printf("\nGuardian eval regression — %d fixture(s) failed their heuristic:\n", len(regressions))
			for _, f := range regressions {
				fmt.Println("  - " + f)
			}
			failed = true
		}
	}

	if scenarioSuite && len(requiredProposalIDs) > 0 {
		fmt.Printf("\nMulti-turn UI scenarios left %d proposal(s) pending (TTL bumped).\n", len(requiredProposalIDs))
		fmt.Println("  Open http://localhost:5173/chat?tab=pending to Confirm, Refine, or Dismiss.")
		for _, id := range requiredProposalIDs {
			fmt.Printf("  - %s\n", id)
		}
	}

	if *checkPendingProposals && !scenarioSuite {
		if runOpts.CheckPendingPerPrompt {
			if *leavePending {
				fmt.Printf("\nPending queue verified per prompt; TTL bumped for UI review (%d proposal_id(s)).\n", len(requiredProposalIDs))
				fmt.Println("  Open http://localhost:5173/chat?tab=pending to Confirm or Dismiss.")
			} else {
				fmt.Printf("\nPending queue verified per prompt (%d proposal_id(s) from passed write-intent fixtures).\n", len(requiredProposalIDs))
			}
		} else if err := reportPendingProposals(ctx, client, expectedProposals, requiredProposalIDs); err != nil {
			fmt.Printf("\nPending change-request queue check failed: %v\n", err)
			failed = true
		}
	}

	if *confirmProposals && !scenarioSuite {
		if !*checkPendingProposals {
			fmt.Println("\nConfirm skipped: -confirm-proposals requires -check-pending-proposals")
			failed = true
		} else if runOpts.ConfirmPerPrompt {
			fmt.Printf("\nConfirm→DB verified per prompt for passed write-intent fixtures.\n")
		} else if !failed {
			for model, details := range rep.Details {
				fmt.Printf("\nConfirming passed write-intent proposals (%s)…\n", model)
				if err := eval.ConfirmAndVerifyPassedProposals(ctx, client, fixtures, details); err != nil {
					fmt.Printf("Confirm→DB check failed: %v\n", err)
					failed = true
				} else {
					fmt.Printf("Confirm→DB check passed for %s\n", model)
				}
			}
		}
	}

	if failed {
		os.Exit(1)
	}
}

// passedProposalFixtures counts how many of this run's ExpectProposal
// fixtures actually passed their heuristic — the number of pending
// change-request rows we'd expect to find afterward.
func passedProposalFixtures(fixtures []eval.Question, scores []farmguardian.EvalQuestionScore) int {
	expectByID := make(map[string]bool, len(fixtures))
	for _, q := range fixtures {
		if q.ExpectProposal {
			expectByID[q.ID] = true
		}
	}
	n := 0
	for _, s := range scores {
		if expectByID[s.ID] && s.Passed {
			n++
		}
	}
	return n
}

// passedProposalIDs collects proposal_id values from write-intent fixtures that
// passed their heuristic — used to verify those exact rows reached the pending
// queue instead of counting stale proposals left from earlier runs.
func passedProposalIDs(fixtures []eval.Question, scores []farmguardian.EvalQuestionScore) []string {
	expectByID := make(map[string]bool, len(fixtures))
	for _, q := range fixtures {
		if q.ExpectProposal {
			expectByID[q.ID] = true
		}
	}
	var out []string
	for _, s := range scores {
		if !expectByID[s.ID] || !s.Passed {
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

// reportPendingProposals fetches Guardian's pending change-request queue
// (GET /v1/chat/proposals?status=pending — the same endpoint the UI's PR
// queue reads) and confirms at least `expected` rows are sitting there,
// printing each one found. This is the actual functional check: a chat
// response can echo a "proposal" object without ever persisting a
// confirmable row, so this is what proves the write-intent flow really
// works end to end, not just that the LLM formatted valid proposal JSON.
func reportPendingProposals(ctx context.Context, client *eval.APIClient, expected int, requiredIDs []string) error {
	pending, err := client.FetchPendingProposals(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("\nPending change-request queue: %d row(s)\n", len(pending))
	for _, p := range pending {
		fmt.Printf("  - [%s] %s — %s (risk: %s)\n", p.ProposalID, p.Tool, p.Summary, p.RiskTier)
	}
	if len(requiredIDs) > 0 {
		pendingSet := make(map[string]struct{}, len(pending))
		for _, p := range pending {
			if id := strings.TrimSpace(p.ProposalID); id != "" {
				pendingSet[id] = struct{}{}
			}
		}
		var missing []string
		for _, id := range requiredIDs {
			if _, ok := pendingSet[id]; !ok {
				missing = append(missing, id)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("expected proposal_id(s) from this run in pending queue, missing: %s", strings.Join(missing, ", "))
		}
		fmt.Printf("Verified %d proposal_id(s) from this run are pending.\n", len(requiredIDs))
		return nil
	}
	if expected > 0 && len(pending) < expected {
		return fmt.Errorf("expected at least %d pending proposal(s) from this run's write-intent fixtures, found %d — a proposal may be echoed in the chat response without actually being persisted", expected, len(pending))
	}
	return nil
}

// regressionFailures returns a sorted "<model>/<id>: <notes>" line for every
// fixture that failed its heuristic — the pure logic behind
// -fail-on-regression, split out so it's unit-testable without a live LLM.
func regressionFailures(details map[string][]farmguardian.EvalQuestionScore) []string {
	var out []string
	for model, scores := range details {
		for _, s := range scores {
			if !s.Passed {
				out = append(out, fmt.Sprintf("%s/%s: %s", model, s.ID, s.Notes))
			}
		}
	}
	sort.Strings(out)
	return out
}

func qaArchivePath(explicit, suite, model string) string {
	if strings.TrimSpace(explicit) == "none" {
		return ""
	}
	if strings.TrimSpace(explicit) != "" {
		return explicit
	}
	return farmguardian.DefaultQARunArchivePath(suite, model)
}

func resolveModels(ctx context.Context, modelsFlag, llmBase string) ([]string, error) {
	if strings.TrimSpace(modelsFlag) != "" && !strings.EqualFold(strings.TrimSpace(modelsFlag), "all") {
		parts := strings.Split(modelsFlag, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out, nil
	}
	if llmBase == "" {
		return nil, fmt.Errorf("LLM_BASE_URL required when models=all")
	}
	discovered, err := farmguardian.DiscoverOllamaModels(ctx, llmBase, http.DefaultClient)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(discovered))
	for _, m := range discovered {
		if farmguardian.IsChatCapable(m.Capabilities) {
			out = append(out, m.Name)
		}
	}
	return out, nil
}

func normalizeModelKey(name string) string {
	return strings.TrimSuffix(strings.TrimSpace(name), ":latest")
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func printModelSummary(model string, s farmguardian.EvalSummary) {
	fmt.Printf("  %s: grounded cite %.0f%% · decline %.0f%% · proposal %.0f%% · latency %.0fms\n",
		model,
		s.GroundedCitationRate*100,
		s.DeclineRate*100,
		s.ProposalValidRate*100,
		s.MeanLatencyMs,
	)
}
