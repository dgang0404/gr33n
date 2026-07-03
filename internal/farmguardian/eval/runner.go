package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// RunModel executes all fixtures for one model name.
func RunModel(ctx context.Context, api *APIClient, model string) ([]ScoreResult, error) {
	var out []ScoreResult
	for _, q := range Fixtures() {
		in, err := api.RunQuestion(ctx, model, q)
		if err != nil {
			out = append(out, ScoreResult{ID: q.ID, Category: q.Category, Passed: false, Notes: err.Error()})
			continue
		}
		out = append(out, Score(in))
	}
	return out, nil
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
