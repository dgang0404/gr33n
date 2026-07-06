package farmguardian

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// EvalSummary is the operator-facing quality snapshot for one model (Phase 122).
type EvalSummary struct {
	Status               string  `json:"eval_status"`
	EvaluatedAt          string  `json:"evaluated_at,omitempty"`
	TotalQuestions       int     `json:"total_questions,omitempty"`
	GroundedCitationRate float64 `json:"grounded_citation_rate,omitempty"`
	DeclineRate          float64 `json:"decline_rate,omitempty"`
	ProposalValidRate    float64 `json:"proposal_valid_rate,omitempty"`
	MeanLatencyMs        float64 `json:"mean_latency_ms,omitempty"`
	RepairAttemptsAvg    float64 `json:"repair_attempts_avg,omitempty"`
	ReportPath           string  `json:"report_path,omitempty"`
}

// EvalReport is persisted by cmd/guardian-eval.
type EvalReport struct {
	UpdatedAt string                       `json:"updated_at"`
	Models    map[string]EvalSummary         `json:"models"`
	Details   map[string][]EvalQuestionScore `json:"details,omitempty"`
}

// EvalQuestionScore is one fixture result row.
type EvalQuestionScore struct {
	ID            string  `json:"id"`
	Category      string  `json:"category"`
	Passed        bool    `json:"passed"`
	LatencyMs     float64 `json:"latency_ms"`
	RepairUsed    bool    `json:"repair_used,omitempty"`
	Notes         string  `json:"notes,omitempty"`
	Prompt        string  `json:"prompt,omitempty"`
	Answer        string  `json:"answer,omitempty"`
	Error         string  `json:"error,omitempty"`
	CitationCount int     `json:"citation_count,omitempty"`
	ProposalCount int     `json:"proposal_count,omitempty"`
	Grounded      bool    `json:"grounded,omitempty"`
	Model         string  `json:"model,omitempty"`
	LogEvidence   []string `json:"log_evidence,omitempty"`
}

var (
	evalCacheMu sync.RWMutex
	evalCache   EvalReport
	evalLoaded  bool
)

// DefaultEvalReportPath is where guardian-eval writes scores.
func DefaultEvalReportPath() string {
	if p := strings.TrimSpace(os.Getenv("GUARDIAN_EVAL_REPORT")); p != "" {
		return p
	}
	return filepath.Join("data", "guardian_model_eval.json")
}

// LoadEvalReport reads the cached eval report from disk.
func LoadEvalReport(path string) (EvalReport, error) {
	if path == "" {
		path = DefaultEvalReportPath()
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return EvalReport{}, err
	}
	var rep EvalReport
	if err := json.Unmarshal(raw, &rep); err != nil {
		return EvalReport{}, err
	}
	return rep, nil
}

// SaveEvalReport writes the eval report atomically.
func SaveEvalReport(path string, rep EvalReport) error {
	if path == "" {
		path = DefaultEvalReportPath()
	}
	if rep.UpdatedAt == "" {
		rep.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if rep.Models == nil {
		rep.Models = map[string]EvalSummary{}
	}
	raw, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// RefreshEvalCache loads eval summaries for API/UI merge.
func RefreshEvalCache() {
	path := DefaultEvalReportPath()
	rep, err := LoadEvalReport(path)
	if err != nil {
		evalCacheMu.Lock()
		evalCache = EvalReport{Models: map[string]EvalSummary{}}
		evalLoaded = true
		evalCacheMu.Unlock()
		return
	}
	evalCacheMu.Lock()
	evalCache = rep
	evalLoaded = true
	evalCacheMu.Unlock()
}

// EvalSummaryForModel returns cached eval data for a model name.
func EvalSummaryForModel(name string) EvalSummary {
	evalCacheMu.RLock()
	defer evalCacheMu.RUnlock()
	if !evalLoaded {
		return EvalSummary{Status: "not_evaluated"}
	}
	for _, key := range modelLookupKeys(name) {
		if s, ok := evalCache.Models[key]; ok {
			if s.Status == "" {
				s.Status = "evaluated"
			}
			return s
		}
	}
	return EvalSummary{Status: "not_evaluated"}
}

// MergeEvalIntoModels attaches eval summaries to model list entries.
func MergeEvalIntoModels(models []ModelInfo) []ModelInfo {
	if len(models) == 0 {
		return models
	}
	out := make([]ModelInfo, len(models))
	for i, m := range models {
		eval := EvalSummaryForModel(m.Name)
		m.Eval = &eval
		out[i] = m
	}
	return out
}

// QARunArchive is a full recorded eval run (Phase 131).
type QARunArchive struct {
	UpdatedAt string              `json:"updated_at"`
	Suite     string              `json:"suite"`
	Model     string              `json:"model"`
	Scores    []EvalQuestionScore `json:"scores"`
}

// DefaultQARunsDir returns the guardian QA run archive directory.
func DefaultQARunsDir() string {
	if p := strings.TrimSpace(os.Getenv("GUARDIAN_QA_RUNS_DIR")); p != "" {
		return p
	}
	return filepath.Join("data", "guardian_qa_runs")
}

// DefaultQARunArchivePath builds a timestamped archive path for one suite+model run.
func DefaultQARunArchivePath(suite, model string) string {
	stamp := time.Now().UTC().Format("20060102T150405")
	safe := strings.NewReplacer(":", "-", "/", "-").Replace(strings.TrimSpace(model))
	if safe == "" {
		safe = "model"
	}
	name := fmt.Sprintf("%s_%s_%s.json", stamp, strings.TrimSpace(suite), safe)
	return filepath.Join(DefaultQARunsDir(), name)
}

// SaveQARunArchive writes a full QA run with answers to disk.
func SaveQARunArchive(path, suite, model string, scores []EvalQuestionScore) error {
	if path == "" {
		return nil
	}
	arch := QARunArchive{
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Suite:     suite,
		Model:     model,
		Scores:    scores,
	}
	raw, err := json.MarshalIndent(arch, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
