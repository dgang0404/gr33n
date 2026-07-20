package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"gr33n-api/internal/farmguardian"
)

// APIClient runs grounded chat turns against a live gr33n API.
type APIClient struct {
	BaseURL string
	Token   string
	FarmID  int64
	HTTP    *http.Client
}

// NewAPIClient builds a client with eval-appropriate HTTP timeout.
func NewAPIClient(baseURL, token string, farmID int64) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Token:   token,
		FarmID:  farmID,
		HTTP:    &http.Client{Timeout: ClientTimeoutFromEnv()},
	}
}

type chatResponse struct {
	Answer       string                         `json:"answer"`
	SessionID    string                         `json:"session_id"`
	Citations    []farmguardian.CitationSummary `json:"citations"`
	Proposals    []farmguardian.ActionProposal  `json:"proposals"`
	Debug        *farmguardian.TurnDebug        `json:"debug,omitempty"`
	AccuracyNote string                         `json:"accuracy_note,omitempty"`
}

func accuracyNoteFromChatResponse(parsed chatResponse) string {
	if note := strings.TrimSpace(parsed.AccuracyNote); note != "" {
		return note
	}
	if parsed.Debug != nil {
		return strings.TrimSpace(parsed.Debug.AccuracyNote)
	}
	return ""
}

// RunQuestion posts one eval prompt with a model override.
func (c *APIClient) RunQuestion(ctx context.Context, model string, q Question) (ScoreInput, error) {
	in, _, err := c.RunQuestionInSession(ctx, model, q, "")
	return in, err
}

// RunQuestionInSession posts one eval prompt, optionally continuing an existing session_id.
func (c *APIClient) RunQuestionInSession(ctx context.Context, model string, q Question, sessionID string) (ScoreInput, string, error) {
	if c == nil || c.HTTP == nil {
		return ScoreInput{}, "", fmt.Errorf("eval API client not configured")
	}
	if strings.TrimSpace(q.Model) != "" {
		model = strings.TrimSpace(q.Model)
	}
	body := map[string]any{
		"message": q.Prompt,
		"stream":  false,
		"model":   model,
	}
	if q.Grounded && c.FarmID > 0 {
		body["farm_id"] = c.FarmID
	}
	if q.ContextRef != nil {
		body["context_ref"] = q.ContextRef
	}
	if sessionID = strings.TrimSpace(sessionID); sessionID != "" {
		body["session_id"] = sessionID
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/v1/chat", bytes.NewReader(raw))
	if err != nil {
		return ScoreInput{}, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Guardian-Eval-Id", q.ID)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	start := time.Now()
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return ScoreInput{}, "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	latency := time.Since(start)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ScoreInput{Question: q, Latency: latency}, "", fmt.Errorf("chat HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return ScoreInput{}, "", err
	}
	return ScoreInput{
		Question:      q,
		Answer:        parsed.Answer,
		CitationCount: len(parsed.Citations),
		ProposalCount: len(parsed.Proposals),
		ProposalIDs:   proposalIDsFromResponse(parsed.Proposals),
		Citations:     parsed.Citations,
		Relevance:     farmguardian.RelevanceFromTurnDebug(parsed.Debug),
		Critique:      farmguardian.CritiqueFromTurnDebug(parsed.Debug),
		AccuracyNote:  accuracyNoteFromChatResponse(parsed),
		Latency:       latency,
	}, strings.TrimSpace(parsed.SessionID), nil
}

func proposalIDsFromResponse(props []farmguardian.ActionProposal) []string {
	if len(props) == 0 {
		return nil
	}
	out := make([]string, 0, len(props))
	for _, p := range props {
		if id := strings.TrimSpace(p.ProposalID); id != "" {
			out = append(out, id)
		}
	}
	return out
}

// RunSuite executes fixtures sequentially (one prompt at a time).
// When warmupGrounded is true, runs farm_counsel warmup before the first grounded prompt.
// When CheckPendingPerPrompt is set, verifies each passed write-intent proposal is still
// pending immediately after its chat turn (before ProposalTTL expires).
// When ConfirmPerPrompt is set, confirms and verifies DB side effects the same way.
func RunSuite(ctx context.Context, api *APIClient, model string, fixtures []Question, opts RunSuiteOptions) ([]ScoreResult, error) {
	var out []ScoreResult
	groundedWarmed := false
	for i, q := range fixtures {
		log.Printf("eval: [%d/%d] starting %q (grounded=%v)", i+1, len(fixtures), q.ID, q.Grounded)
		if q.Grounded && opts.WarmupGrounded && !groundedWarmed {
			groundedWarmed = true
			warmModel := model
			if strings.TrimSpace(q.Model) != "" {
				warmModel = strings.TrimSpace(q.Model)
			}
			if err := api.WarmupFarmCounsel(ctx, warmModel, opts.WarmupTimeout); err != nil {
				log.Printf("eval: warmup before grounded block: %v (continuing)", err)
			} else {
				log.Printf("eval: counsel model ready before grounded block")
			}
		}
		m := model
		if strings.TrimSpace(q.Model) != "" {
			m = strings.TrimSpace(q.Model)
		}
		in, err := api.RunQuestion(ctx, m, q)
		if err != nil {
			log.Printf("eval: [%d/%d] %q failed: %v", i+1, len(fixtures), q.ID, err)
			out = append(out, scoreResultFromError(q, m, err))
			continue
		}
		res := Score(in)
		enrichScoreResult(&res, in, m)
		status := "pass"
		if !res.Passed {
			status = "fail"
		}
		log.Printf("eval: [%d/%d] %q %s in %.1fs (proposals=%d citations=%d)",
			i+1, len(fixtures), q.ID, status, in.Latency.Seconds(), res.ProposalCount, res.CitationCount)
		if q.ExpectProposal && res.Passed && len(res.ProposalIDs) > 0 {
			if opts.CheckPendingPerPrompt {
				if err := VerifyPendingProposalIDs(ctx, api, res.ProposalIDs); err != nil {
					return out, fmt.Errorf("%s pending queue: %w", q.ID, err)
				}
				log.Printf("eval: [%d/%d] %q verified pending queue (%d proposal_id(s))",
					i+1, len(fixtures), q.ID, len(res.ProposalIDs))
			}
			if opts.LeavePending {
				ttl := opts.LeavePendingTTL
				if ttl <= 0 {
					ttl = LeavePendingTTLFromEnv()
				}
				n, err := BumpProposalExpiry(ctx, res.ProposalIDs, ttl)
				if err != nil {
					return out, fmt.Errorf("%s bump pending TTL: %w", q.ID, err)
				}
				log.Printf("eval: [%d/%d] %q left pending for UI (%d proposal(s), expires ~%s)",
					i+1, len(fixtures), q.ID, n, time.Now().UTC().Add(ttl).Format(time.RFC3339))
			}
			if opts.ConfirmPerPrompt {
				for _, pid := range res.ProposalIDs {
					if err := ConfirmAndVerifyProposal(ctx, api, q.ID, pid); err != nil {
						return out, err
					}
				}
			}
		}
		if opts.LogPath != "" && q.ExpectTool != "" {
			ev := ScrapeLogEvidence(opts.LogPath, q.ID, q.ExpectTool)
			if len(ev) > 0 {
				res.LogEvidence = ev
				if (q.ID == "smoke-morning-walk" || q.ID == "p128-devices") && !res.Passed && smokeAnswerAllowsLogOverride(q, in.Answer) {
					res.Passed = true
					res.Notes = "log evidence: " + strings.Join(ev, "; ")
				}
			}
		}
		out = append(out, res)
	}
	return out, nil
}

// RunSuiteOptions configures sequential eval runs.
type RunSuiteOptions struct {
	WarmupGrounded       bool
	WarmupTimeout        time.Duration
	WarmupAsync          bool
	LogPath              string
	CheckPendingPerPrompt bool
	ConfirmPerPrompt     bool
	LeavePending         bool
	LeavePendingTTL      time.Duration
}

// RunModel executes regression fixtures for one model name.
func RunModel(ctx context.Context, api *APIClient, model string) ([]ScoreResult, error) {
	return RunSuite(ctx, api, model, Fixtures(), RunSuiteOptions{})
}

func scoreResultFromError(q Question, model string, err error) ScoreResult {
	return ScoreResult{
		ID:       q.ID,
		Category: q.Category,
		Passed:   false,
		Notes:    err.Error(),
		Error:    err.Error(),
		Prompt:   q.Prompt,
		Grounded: q.Grounded,
		Model:    model,
	}
}

func enrichScoreResult(res *ScoreResult, in ScoreInput, model string) {
	if res == nil {
		return
	}
	res.Prompt = in.Question.Prompt
	res.Answer = in.Answer
	res.CitationCount = in.CitationCount
	res.ProposalCount = in.ProposalCount
	res.ProposalIDs = append([]string(nil), in.ProposalIDs...)
	res.Citations = in.Citations
	res.Relevance = in.Relevance
	if in.Critique.Enabled && !in.Critique.Skipped {
		pass := in.Critique.Pass
		res.CritiquePass = &pass
		res.CritiqueReason = in.Critique.Reason
	}
	res.Grounded = in.Question.Grounded
	res.Model = model
	res.AccuracyNote = in.AccuracyNote
}

// BuildReport aggregates scores into a farmguardian.EvalSummary for one model.
func BuildReport(model string, scores []ScoreResult, reportPath string) farmguardian.EvalSummary {
	cite, dec, prop, lat, repair := Aggregate(scores)
	return farmguardian.EvalSummary{
		Status:               "evaluated",
		EvaluatedAt:          time.Now().UTC().Format(time.RFC3339),
		TotalQuestions:       len(scores),
		GroundedCitationRate: cite,
		DeclineRate:          dec,
		ProposalValidRate:    prop,
		MeanLatencyMs:        lat,
		RepairAttemptsAvg:    repair,
		ReportPath:           reportPath,
	}
}

// ToEvalQuestionScores maps runner results for JSON persistence.
func ToEvalQuestionScores(scores []ScoreResult) []farmguardian.EvalQuestionScore {
	out := make([]farmguardian.EvalQuestionScore, len(scores))
	for i, s := range scores {
		out[i] = farmguardian.EvalQuestionScore{
			ID: s.ID, Category: s.Category, Passed: s.Passed,
			LatencyMs: s.LatencyMs, RepairUsed: s.RepairUsed, Notes: s.Notes,
			Prompt: s.Prompt, Answer: s.Answer, Error: s.Error,
			CitationCount: s.CitationCount, ProposalCount: s.ProposalCount,
			ProposalIDs: append([]string(nil), s.ProposalIDs...),
			Grounded: s.Grounded, Model: s.Model, LogEvidence: s.LogEvidence,
			Citations: s.Citations,
			QuestionAnswerRelevance: s.Relevance.QuestionAnswerCosine,
			OpeningTailRelevance:    s.Relevance.OpeningTailCosine,
			LowRelevance:            s.Relevance.LowRelevance,
			CritiquePass:            s.CritiquePass,
			CritiqueReason:          s.CritiqueReason,
			AccuracyNote:            s.AccuracyNote,
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
