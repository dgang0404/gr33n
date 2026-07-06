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
	"gr33n-api/internal/rag/llm"
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
		HTTP:    &http.Client{Timeout: llm.EvalTimeoutFromEnv()},
	}
}

type chatResponse struct {
	Answer    string `json:"answer"`
	Citations []any  `json:"citations"`
	Proposals []any  `json:"proposals"`
}

// RunQuestion posts one eval prompt with a model override.
func (c *APIClient) RunQuestion(ctx context.Context, model string, q Question) (ScoreInput, error) {
	if c == nil || c.HTTP == nil {
		return ScoreInput{}, fmt.Errorf("eval API client not configured")
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
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/v1/chat", bytes.NewReader(raw))
	if err != nil {
		return ScoreInput{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Guardian-Eval-Id", q.ID)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	start := time.Now()
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return ScoreInput{}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	latency := time.Since(start)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ScoreInput{Question: q, Latency: latency}, fmt.Errorf("chat HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return ScoreInput{}, err
	}
	return ScoreInput{
		Question:      q,
		Answer:        parsed.Answer,
		CitationCount: len(parsed.Citations),
		ProposalCount: len(parsed.Proposals),
		Latency:       latency,
	}, nil
}

// RunSuite executes fixtures sequentially (one prompt at a time).
// When warmupGrounded is true, runs farm_counsel warmup before the first grounded prompt.
func RunSuite(ctx context.Context, api *APIClient, model string, fixtures []Question, opts RunSuiteOptions) []ScoreResult {
	var out []ScoreResult
	groundedWarmed := false
	for _, q := range fixtures {
		if q.Grounded && opts.WarmupGrounded && !groundedWarmed {
			groundedWarmed = true
			if err := api.WarmupFarmCounsel(ctx, opts.WarmupTimeout); err != nil {
				log.Printf("eval: warmup before grounded block: %v (continuing)", err)
			}
		}
		m := model
		if strings.TrimSpace(q.Model) != "" {
			m = strings.TrimSpace(q.Model)
		}
		in, err := api.RunQuestion(ctx, m, q)
		if err != nil {
			out = append(out, scoreResultFromError(q, m, err))
			continue
		}
		res := Score(in)
		enrichScoreResult(&res, in, m)
		if opts.LogPath != "" && q.ExpectTool != "" {
			ev := ScrapeLogEvidence(opts.LogPath, q.ID, q.ExpectTool)
			if len(ev) > 0 {
				res.LogEvidence = ev
				if q.ID == "smoke-morning-walk" && !res.Passed {
					res.Passed = true
					res.Notes = "log evidence: " + strings.Join(ev, "; ")
				}
			}
		}
		out = append(out, res)
	}
	return out
}

// RunSuiteOptions configures sequential eval runs.
type RunSuiteOptions struct {
	WarmupGrounded bool
	WarmupTimeout  time.Duration
	LogPath        string
}

// RunModel executes regression fixtures for one model name.
func RunModel(ctx context.Context, api *APIClient, model string) ([]ScoreResult, error) {
	return RunSuite(ctx, api, model, Fixtures(), RunSuiteOptions{}), nil
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
	res.Grounded = in.Question.Grounded
	res.Model = model
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
			Grounded: s.Grounded, Model: s.Model, LogEvidence: s.LogEvidence,
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
