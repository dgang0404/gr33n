package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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
	reportPath := flag.String("report", farmguardian.DefaultEvalReportPath(), "output JSON report path")
	llmBase := flag.String("llama-url", os.Getenv("LLM_BASE_URL"), "Ollama OpenAI base (for model discovery when models=all)")
	flag.Parse()

	if strings.TrimSpace(*token) == "" {
		log.Fatal("JWT required: pass -token or set GUARDIAN_EVAL_TOKEN (use make dev-auth-test login token)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	modelNames, err := resolveModels(ctx, *modelsFlag, *llmBase)
	if err != nil {
		log.Fatal(err)
	}
	if len(modelNames) == 0 {
		log.Fatal("no chat-capable models to evaluate")
	}

	client := &eval.APIClient{
		BaseURL: *apiURL,
		Token:   *token,
		FarmID:  *farmID,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}

	rep := farmguardian.EvalReport{
		Models:  map[string]farmguardian.EvalSummary{},
		Details: map[string][]farmguardian.EvalQuestionScore{},
	}
	for _, model := range modelNames {
		log.Printf("evaluating model %q (%d questions)…", model, len(eval.Fixtures()))
		scores, err := eval.RunModel(ctx, client, model)
		if err != nil {
			log.Printf("model %q failed: %v", model, err)
			continue
		}
		rep.Models[normalizeModelKey(model)] = eval.BuildReport(model, scores, *reportPath)
		details := make([]farmguardian.EvalQuestionScore, len(scores))
		for i, s := range scores {
			details[i] = farmguardian.EvalQuestionScore{
				ID: s.ID, Category: s.Category, Passed: s.Passed,
				LatencyMs: s.LatencyMs, RepairUsed: s.RepairUsed, Notes: s.Notes,
			}
		}
		rep.Details[normalizeModelKey(model)] = details
		printModelSummary(model, rep.Models[normalizeModelKey(model)])
	}

	if err := farmguardian.SaveEvalReport(*reportPath, rep); err != nil {
		log.Fatal(err)
	}
	farmguardian.RefreshEvalCache()
	fmt.Printf("\nEval report written to %s\n", *reportPath)
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
