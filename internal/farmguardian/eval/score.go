package eval

import (
	"strings"
	"time"

	"gr33n-api/internal/farmguardian"
)

// Question is one eval fixture prompt.
type Question struct {
	ID              string
	Category        string // field_guide | farm_state | out_of_scope | write_intent | ungrounded
	Prompt          string
	ExpectCitation  bool
	ExpectDecline   bool
	ExpectProposal  bool
	ExpectTool      string // optional log/tool evidence hint (Phase 131)
	Grounded        bool
	Model           string // optional per-fixture model override (Phase 131 smoke)
	ContextRef      *farmguardian.ContextRef // optional UI entry point (Phase 60 morning check)
}

// Fixtures returns the Phase 122 eval question set plus RegressionFixtures append.
func Fixtures() []Question {
	base := []Question{
		{ID: "fg-apple-nursery", Category: "field_guide", Prompt: "What should I watch for in an apple nursery according to the field guides?", ExpectCitation: true, Grounded: true},
		{ID: "fg-tomato-veg", Category: "field_guide", Prompt: "Summarize tomato vegetative stage care from our field guides.", ExpectCitation: true, Grounded: true},
		{ID: "fg-citation-format", Category: "field_guide", Prompt: "What EC range does the platform recommend for hydro lettuce?", ExpectCitation: true, Grounded: true},
		{ID: "farm-alerts", Category: "farm_state", Prompt: "What unread alerts do I have right now?", Grounded: true},
		{
			ID:         "farm-morning-walkthrough",
			Category:   "farm_state",
			Prompt:     MorningWalkPrompt(),
			Grounded:   true,
			ExpectTool: "walk_farm",
			ContextRef: MorningWalkContextRef(),
		},
		{ID: "farm-zones", Category: "farm_state", Prompt: "List my zone names and active crop cycles.", Grounded: true},
		{ID: "farm-plants", Category: "farm_state", Prompt: "Which plants are registered on this farm?", Grounded: true},
		{ID: "farm-low-stock", Category: "farm_state", Prompt: "Anything running low in supplies?", Grounded: true},
		{ID: "farm-programs", Category: "farm_state", Prompt: "Which fertigation programs target Veg Tent?", Grounded: true},
		{ID: "decline-geography", Category: "out_of_scope", Prompt: "What's the capital of France?", ExpectDecline: true, Grounded: true},
		{ID: "decline-coding", Category: "out_of_scope", Prompt: "Write me a Python script to scrape Reddit.", ExpectDecline: true, Grounded: true},
		{ID: "decline-invent-zone", Category: "out_of_scope", Prompt: "What is the VPD in the Secret Mars Dome zone?", ExpectDecline: true, Grounded: true},
		{ID: "write-feed", Category: "write_intent", Prompt: "Set the feed volume to 0.3 liters for the Veg Tent program.", ExpectProposal: true, Grounded: true},
		{ID: "write-ack", Category: "write_intent", Prompt: "Acknowledge the highest severity unread alert.", ExpectProposal: true, Grounded: true},
		{ID: "write-schedule", Category: "write_intent", Prompt: "Pause the lights schedule for Veg Tent until tomorrow.", ExpectProposal: true, Grounded: true},
		{ID: "readonly-ec", Category: "farm_state", Prompt: "Why is EC trending high in Veg Tent?", Grounded: true},
		{ID: "readonly-vpd", Category: "farm_state", Prompt: "What is VPD and where do I see it in the UI?", Grounded: true},
		{ID: "fg-general", Category: "field_guide", Prompt: "How should I wire a DHT22 on the Pi according to platform docs?", ExpectCitation: true, Grounded: true},
		{ID: "decline-medical", Category: "out_of_scope", Prompt: "What medicine should I take for a headache?", ExpectDecline: true, Grounded: true},
		{ID: "farm-sensors", Category: "farm_state", Prompt: "Which sensors are in Veg Tent?", Grounded: true},
		{ID: "write-task", Category: "write_intent", Prompt: "Create a task to refill calcium nitrate when stock is low.", ExpectProposal: true, Grounded: true},
		// Phase 127 — device + fertigation posture / field guides
		{ID: "farm-devices", Category: "farm_state", Prompt: "Are any edge devices offline on this farm?", Grounded: true},
		{ID: "farm-fert-schedule", Category: "farm_state", Prompt: "Which active fertigation programs are manual-only without a schedule?", Grounded: true},
		{ID: "fg-demo-pi", Category: "field_guide", Prompt: "Which relay channel is the veg grow light on the gr33n demo farm?", ExpectCitation: true, Grounded: true},
		{ID: "fg-fertigation-triage", Category: "field_guide", Prompt: "Program is active but no dose ran — what should I check first per field guides?", ExpectCitation: true, Grounded: true},
	}
	return append(base, RegressionFixtures()...)
}

// ScoreInput is one model answer to score.
type ScoreInput struct {
	Question       Question
	Answer         string
	CitationCount  int
	ProposalCount  int
	ProposalIDs    []string
	Citations      []farmguardian.CitationSummary
	Relevance      farmguardian.AnswerRelevance
	Critique       farmguardian.AnswerCritique
	Latency        time.Duration
	RepairAttempt   bool
	RepairRecovered bool
	AccuracyNote    string
}

// ScoreResult is the automatic scoring outcome.
type ScoreResult struct {
	ID            string
	Category      string
	Passed        bool
	Notes         string
	LatencyMs     float64
	RepairUsed    bool
	Prompt        string
	Answer        string
	Error         string
	CitationCount int
	ProposalCount int
	ProposalIDs   []string
	Grounded      bool
	Model         string
	LogEvidence   []string
	Citations     []farmguardian.CitationSummary
	Relevance     farmguardian.AnswerRelevance
	CritiquePass   *bool
	CritiqueReason string
	AccuracyNote   string
}

// Score evaluates one answer heuristically.
func Score(in ScoreInput) ScoreResult {
	res := ScoreResult{
		ID:        in.Question.ID,
		Category:  in.Question.Category,
		LatencyMs: float64(in.Latency.Milliseconds()),
		RepairUsed: in.RepairAttempt,
	}
	a := strings.ToLower(strings.TrimSpace(in.Answer))
	switch {
	case in.Question.ID == "smoke-cherry-forest":
		res.Passed = len(a) > 80 && (strings.Contains(a, "goldenrod") || strings.Contains(a, "blackberry") || strings.Contains(a, "cherry"))
		if !res.Passed {
			res.Notes = "expected forest-garden answer mentioning cherry/goldenrod/blackberry"
		}
	case in.Question.ID == "regression-cherry-goldenrod-jlf", in.Question.ID == "smoke-cherry-jlf", in.Question.ID == "smoke-nf-goldenrod":
		res.Passed, res.Notes = scoreRegressionCherryGoldenrodJLF(in)
	case in.Question.ID == "smoke-morning-walk", in.Question.ID == "farm-morning-walkthrough":
		res.Passed = len(a) > 40 && !looksLikeInvention(a)
		if !res.Passed && res.Notes == "" {
			res.Notes = "expected morning walkthrough answer with farm specifics"
		}
	case in.Question.ID == "smoke-unread-alerts":
		hasFarmContent := strings.Contains(a, "alert") || strings.Contains(a, "humidity") ||
			strings.Contains(a, "ohn") || strings.Contains(a, "photoperiod") || strings.Contains(a, "seed")
		res.Passed = len(a) > 40 && hasFarmContent &&
			(in.CitationCount > 0 || citationRefPresent(in.Answer))
		if !res.Passed && res.Notes == "" {
			res.Notes = "expected alert summary with numbered citations [1]/[2] or citation_count > 0"
		}
	case in.Question.ID == "smoke-ec-ph":
		hasPH := strings.Contains(a, "ph")
		hasEC := strings.Contains(a, "ec") || in.CitationCount > 0 || citationRefPresent(in.Answer)
		res.Passed = hasPH && hasEC
		if !res.Passed && res.Notes == "" {
			res.Notes = "expected EC guidance and explicit pH targets from documentation"
		}
	case in.Question.ID == "farm-devices", in.Question.ID == "p128-devices":
		res.Passed = len(a) > 15 && !looksLikeInvention(a) &&
			(strings.Contains(a, "device") || strings.Contains(a, "offline") ||
				strings.Contains(a, "edge") || strings.Contains(a, "pi") || strings.Contains(a, "online"))
		if !res.Passed {
			res.Notes = "expected device health from snapshot (no invented GPIO)"
		}
	case in.Question.ID == "farm-fert-schedule", in.Question.ID == "p128-fert-manual":
		res.Passed = len(a) > 15 && !looksLikeInvention(a) &&
			(strings.Contains(a, "manual") || strings.Contains(a, "outdoor") || strings.Contains(a, "jlf") ||
				strings.Contains(a, "schedule") || strings.Contains(a, "program"))
		if !res.Passed {
			res.Notes = "expected manual-only program names or schedule posture"
		}
	case in.Question.ID == "fg-demo-pi", in.Question.ID == "p128-demo-pi":
		res.Passed = in.CitationCount > 0 || citationRefPresent(in.Answer) ||
			strings.Contains(a, "relay") || strings.Contains(a, "veg") || strings.Contains(a, "channel")
		if !res.Passed {
			res.Notes = "expected demo-farm-pi-layout citation or relay channel"
		}
	case in.Question.ID == "fg-fertigation-triage", in.Question.ID == "p128-fert-triage":
		res.Passed = in.CitationCount > 0 || citationRefPresent(in.Answer) ||
			strings.Contains(a, "schedule") || strings.Contains(a, "reservoir") ||
			strings.Contains(a, "pi") || strings.Contains(a, "pump") || strings.Contains(a, "dose")
		if !res.Passed {
			res.Notes = "expected fertigation-troubleshooting steps"
		}
	case strings.HasPrefix(in.Question.ID, "smoke-nf-"):
		res.Passed, res.Notes = scoreSmokeNaturalFarming(in)
	case in.Question.ExpectCitation:
		res.Passed = in.CitationCount > 0 || citationRefPresent(in.Answer)
		if !res.Passed {
			res.Notes = "expected citation"
		}
	case in.Question.ExpectDecline:
		res.Passed = looksLikeDecline(a) && !looksLikeInvention(a)
		if !res.Passed {
			res.Notes = "expected polite decline without invention"
		}
	case in.Question.ExpectProposal:
		res.Passed = in.ProposalCount > 0 || proposalJSONPresent(in.Answer)
		if !res.Passed {
			res.Notes = "expected valid proposal"
		}
	default:
		res.Passed = len(a) > 20 && !looksLikeInvention(a)
		if !res.Passed {
			res.Notes = "expected grounded farm answer"
		}
	}
	if in.RepairRecovered && in.Question.ExpectProposal {
		res.Passed = true
		res.Notes = "proposal repair recovered"
	}
	applySmokeTopicDrift(&res, in)
	applyAnswerCritique(&res, in)
	return res
}

func shouldApplySmokeTopicDrift(q Question) bool {
	switch q.ID {
	case "smoke-morning-walk", "smoke-ec-ph", "smoke-cherry-forest", "smoke-unread-alerts", "smoke-cherry-jlf", "farm-morning-walkthrough",
		"smoke-nf-jlf-doc", "smoke-nf-jms-dilution", "smoke-nf-ready-batches", "smoke-nf-jms-make", "smoke-nf-jlf-start",
		"smoke-nf-combined-drench", "smoke-nf-ffj-flower", "smoke-nf-wca-foliar", "smoke-nf-goldenrod", "smoke-nf-lab",
		"p128-fert-triage", "p128-demo-pi", "fg-fertigation-triage", "fg-demo-pi":
		return true
	default:
		return q.Category == "field_guide" && q.ExpectCitation
	}
}

func applySmokeTopicDrift(res *ScoreResult, in ScoreInput) {
	if !res.Passed || !shouldApplySmokeTopicDrift(in.Question) {
		return
	}
	if note := farmguardian.SmokeTopicDriftNote(farmguardian.SmokeTopicDriftInput{
		QuestionID: in.Question.ID,
		Category:   in.Question.Category,
		Prompt:     in.Question.Prompt,
		Answer:     in.Answer,
		Citations:  in.Citations,
		Relevance:  in.Relevance,
	}); note != "" {
		res.Passed = false
		res.Notes = note
	}
}

func applyAnswerCritique(res *ScoreResult, in ScoreInput) {
	if !res.Passed || !in.Critique.Enabled || in.Critique.Skipped {
		return
	}
	if in.Critique.Pass {
		return
	}
	res.Passed = false
	if in.Critique.Reason != "" {
		res.Notes = "critique_fail: " + in.Critique.Reason
	} else {
		res.Notes = "critique_fail"
	}
}

func citationRefPresent(answer string) bool {
	return strings.Contains(answer, "[1]") || strings.Contains(answer, "[2]")
}

func proposalJSONPresent(answer string) bool {
	lower := strings.ToLower(answer)
	return strings.Contains(lower, `"tool"`) && strings.Contains(lower, "patch_")
}

func looksLikeDecline(lowerAnswer string) bool {
	for _, p := range []string{
		"farm operation", "can't help", "cannot help", "outside", "not related",
		"redirect", "gr33n", "dashboard", "guardian",
	} {
		if strings.Contains(lowerAnswer, p) {
			return true
		}
	}
	return false
}

func looksLikeInvention(lowerAnswer string) bool {
	return strings.Contains(lowerAnswer, "secret mars dome") || strings.Contains(lowerAnswer, "mars dome")
}

func scoreRegressionCherryGoldenrodJLF(in ScoreInput) (bool, string) {
	a := strings.ToLower(strings.TrimSpace(in.Answer))
	if len(a) < 60 {
		return false, "expected grounded JLF answer with catalog detail"
	}
	hasJLF := strings.Contains(a, "jlf") ||
		strings.Contains(a, "fermented plant juice") ||
		strings.Contains(a, "jadam liquid fertilizer")
	if !hasJLF {
		return false, "expected JLF / fermented plant juice framing"
	}
	hasDilutionOrCatalog := strings.Contains(a, "1:100") || strings.Contains(a, "1:30") ||
		strings.Contains(a, "1:20") ||
		strings.Contains(a, "process catalog") || strings.Contains(a, "lookup_process") ||
		strings.Contains(a, "field guide") || strings.Contains(a, "field-guides/") ||
		strings.Contains(a, "suggest_process_from_material") ||
		citationRefPresent(in.Answer) || in.CitationCount > 0
	if !hasDilutionOrCatalog {
		return false, "expected dilution band or process catalog / field guide citation"
	}
	if choGoldenrodRecipeClaim(a) {
		return false, "expected extension-method framing, not Cho-named goldenrod recipe"
	}
	if woodlandECInvention(a) {
		return false, "expected no invented EC mS/cm targets for woodland/forage context"
	}
	return true, ""
}

func choGoldenrodRecipeClaim(a string) bool {
	for _, phrase := range []string{
		"not a cho", "not cho-named", "not the cho", "isn't a cho", "is not a cho",
		"no cho-named", "avoid cho-named",
	} {
		if strings.Contains(a, phrase) {
			return false
		}
	}
	return strings.Contains(a, "cho's goldenrod") || strings.Contains(a, "cho goldenrod recipe") ||
		strings.Contains(a, "cho-named goldenrod")
}

func scoreSmokeNaturalFarming(in ScoreInput) (bool, string) {
	a := strings.ToLower(strings.TrimSpace(in.Answer))
	if len(a) < 40 {
		return false, "expected grounded natural farming answer"
	}
	if looksLikeInvention(a) {
		return false, "expected no invented farm data"
	}
	cited := in.CitationCount > 0 || citationRefPresent(in.Answer)
	switch in.Question.ID {
	case "smoke-nf-jlf-doc":
		if !(strings.Contains(a, "jlf") || strings.Contains(a, "liquid fertilizer")) {
			return false, "expected JLF framing"
		}
		if !(strings.Contains(a, "1:100") || strings.Contains(a, "1:20") || strings.Contains(a, "dilut")) {
			return false, "expected dilution guidance from the guide"
		}
		if !cited {
			return false, "expected field guide citation"
		}
	case "smoke-nf-jms-dilution":
		if !strings.Contains(a, "jms") && !strings.Contains(a, "microbial") {
			return false, "expected JMS"
		}
		if !(strings.Contains(a, "1:10") && strings.Contains(a, "1:20")) {
			return false, "expected soil 1:10 and foliar 1:20"
		}
	case "smoke-nf-ready-batches":
		if !(strings.Contains(a, "batch") || strings.Contains(a, "ferment") || strings.Contains(a, "input") || strings.Contains(a, "ready")) {
			return false, "expected inventory / batch summary"
		}
	case "smoke-nf-jms-make":
		if !(strings.Contains(a, "jms") || strings.Contains(a, "microbial")) {
			return false, "expected JMS"
		}
		if !(strings.Contains(a, "potato") || strings.Contains(a, "leaf mold") || strings.Contains(a, "humus")) {
			return false, "expected core JMS ingredients from guide"
		}
	case "smoke-nf-jlf-start":
		if !(strings.Contains(a, "1:100") || (strings.Contains(a, "start") && strings.Contains(a, "1:20"))) {
			return false, "expected conservative 1:100 start dilution"
		}
	case "smoke-nf-combined-drench":
		if !(strings.Contains(a, "jlf") && strings.Contains(a, "jms")) {
			return false, "expected combined JLF + JMS drench"
		}
		if !(strings.Contains(a, "1:10") || strings.Contains(a, "1:20") || strings.Contains(a, "same tank") || strings.Contains(a, "same water")) {
			return false, "expected combined dilution or same-tank guidance"
		}
	case "smoke-nf-ffj-flower":
		if !(strings.Contains(a, "ffj") || strings.Contains(a, "fruit")) {
			return false, "expected FFJ / fruit framing"
		}
		if !(strings.Contains(a, "flower") || strings.Contains(a, "fruit") || strings.Contains(a, "bloom")) {
			return false, "expected flowering use case"
		}
	case "smoke-nf-wca-foliar":
		if !(strings.Contains(a, "wca") || strings.Contains(a, "calcium")) {
			return false, "expected WCA"
		}
		if !(strings.Contains(a, "1:1000") || strings.Contains(a, "foliar")) {
			return false, "expected WCA foliar dilution band"
		}
	case "smoke-nf-lab":
		if !(strings.Contains(a, "lab") || strings.Contains(a, "lactic")) {
			return false, "expected LAB"
		}
		if !(strings.Contains(a, "soil") || strings.Contains(a, "odor") || strings.Contains(a, "condition")) {
			return false, "expected soil / conditioning use"
		}
	default:
		return false, "unknown smoke-nf fixture"
	}
	return true, ""
}

func woodlandECInvention(a string) bool {
	if !strings.Contains(a, "ms/cm") {
		return false
	}
	for _, phrase := range []string{
		"no ec", "without ec", "not ec", "skip ec", "instead of ec",
		"don't use ec", "do not use ec", "avoid ec",
	} {
		if strings.Contains(a, phrase) {
			return false
		}
	}
	return strings.Contains(a, "target") || strings.Contains(a, "should be") ||
		strings.Contains(a, "maintain") || strings.Contains(a, "set ec")
}

func smokeAnswerAllowsLogOverride(q Question, answer string) bool {
	switch q.ID {
	case "smoke-morning-walk", "farm-morning-walkthrough":
		return farmguardian.SmokeTopicDriftNote(farmguardian.SmokeTopicDriftInput{
			QuestionID: q.ID,
			Category:   q.Category,
			Prompt:     q.Prompt,
			Answer:     answer,
		}) == ""
	default:
		return true
	}
}

// Aggregate builds per-model summary rates from score rows.
func Aggregate(scores []ScoreResult) (citationRate, declineRate, proposalRate, meanLatency, repairAvg float64) {
	if len(scores) == 0 {
		return
	}
	var citeN, citeD, decN, decD, propN, propD, repairN int
	var latSum float64
	for _, s := range scores {
		latSum += s.LatencyMs
		if s.RepairUsed {
			repairN++
		}
		switch s.Category {
		case "field_guide":
			citeD++
			if s.Passed {
				citeN++
			}
		case "out_of_scope":
			decD++
			if s.Passed {
				decN++
			}
		case "write_intent":
			propD++
			if s.Passed {
				propN++
			}
		}
	}
	if citeD > 0 {
		citationRate = float64(citeN) / float64(citeD)
	}
	if decD > 0 {
		declineRate = float64(decN) / float64(decD)
	}
	if propD > 0 {
		proposalRate = float64(propN) / float64(propD)
	}
	meanLatency = latSum / float64(len(scores))
	repairAvg = float64(repairN) / float64(len(scores))
	return
}
