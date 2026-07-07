package farmguardian

import (
	"encoding/json"
	"fmt"
	"log"
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

// QAFeedbackReviewPrompt is stored on smoke/regression archives and logged after save.
const QAFeedbackReviewPrompt = "After smoke: run docs/guardian-feedback-review-runbook.md § Smoke quality checklist, then Settings → Guardian feedback (or GET /v1/chat/feedback/export)"

// QARunArchive is a full recorded eval run (Phase 131).
type QARunArchive struct {
	UpdatedAt              string              `json:"updated_at"`
	Suite                  string              `json:"suite"`
	Model                  string              `json:"model"`
	FeedbackReviewPrompt   string              `json:"feedback_review_prompt,omitempty"`
	Scores                 []EvalQuestionScore `json:"scores"`
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

// QARunSummary is the operator-facing snapshot of one archived QA run (Phase 140).
type QARunSummary struct {
	UpdatedAt  string `json:"updated_at"`
	Suite      string `json:"suite"`
	Model      string `json:"model"`
	ReportPath string `json:"report_path"`
	Passed     int    `json:"passed"`
	Total      int    `json:"total"`
	AllPassed  bool   `json:"all_passed"`
}

// SummarizeQARun counts heuristic pass/fail rows in an archive.
func SummarizeQARun(arch QARunArchive) QARunSummary {
	out := QARunSummary{
		UpdatedAt: arch.UpdatedAt,
		Suite:     arch.Suite,
		Model:     arch.Model,
		Total:     len(arch.Scores),
	}
	for _, s := range arch.Scores {
		if s.Passed {
			out.Passed++
		}
	}
	out.AllPassed = out.Total > 0 && out.Passed == out.Total
	return out
}

// LoadLatestQARun reads the newest JSON archive from dir (by UpdatedAt, then filename).
func LoadLatestQARun(dir string) (QARunArchive, string, error) {
	if dir == "" {
		dir = DefaultQARunsDir()
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return QARunArchive{}, "", err
		}
		return QARunArchive{}, "", err
	}
	var (
		bestArch  QARunArchive
		bestPath  string
		bestStamp time.Time
		found     bool
	)
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, ent.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var arch QARunArchive
		if err := json.Unmarshal(raw, &arch); err != nil {
			continue
		}
		stamp := parseQARunTimestamp(arch.UpdatedAt, ent.Name())
		if !found || stamp.After(bestStamp) {
			found = true
			bestStamp = stamp
			bestArch = arch
			bestPath = path
		}
	}
	if !found {
		return QARunArchive{}, "", os.ErrNotExist
	}
	return bestArch, bestPath, nil
}

func parseQARunTimestamp(updatedAt, filename string) time.Time {
	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(updatedAt)); err == nil {
		return t
	}
	// Filenames: 20060102T150405_smoke_phi3-mini.json
	if len(filename) >= 15 {
		if t, err := time.Parse("20060102T150405", filename[:15]); err == nil {
			return t
		}
	}
	return time.Time{}
}

// LatestQARunSummary loads the newest archive and returns summary + full scores.
func LatestQARunSummary() (QARunSummary, []EvalQuestionScore, error) {
	arch, path, err := LoadLatestQARun("")
	if err != nil {
		return QARunSummary{}, nil, err
	}
	sum := SummarizeQARun(arch)
	sum.ReportPath = path
	return sum, arch.Scores, nil
}

// SaveQARunArchive writes a full QA run with answers to disk.
func SaveQARunArchive(path, suite, model string, scores []EvalQuestionScore) error {
	if path == "" {
		return nil
	}
	arch := QARunArchive{
		UpdatedAt:            time.Now().UTC().Format(time.RFC3339),
		Suite:                suite,
		Model:                model,
		FeedbackReviewPrompt: QAFeedbackReviewPrompt,
		Scores:               scores,
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
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	log.Printf("guardian qa: archive saved %s — %s", path, QAFeedbackReviewPrompt)
	return nil
}
