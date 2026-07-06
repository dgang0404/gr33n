// Package reingest runs async farm-scoped RAG ingest jobs (Phase 135).
package reingest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"gr33n-api/internal/ai"
	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
	"gr33n-api/internal/rag/embed"
	"gr33n-api/internal/rag/ingest"
)

const (
	ScopeFieldGuides  = "field_guides"
	ScopePlatformDocs = "platform_docs"
	ScopeOperational  = "operational"
	ScopeAll          = "all"

	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"

	jobTimeout = 45 * time.Minute
)

var (
	mu   sync.Mutex
	jobs = map[int64]*Job{}
)

// Job tracks one farm-scoped ingest run.
type Job struct {
	FarmID     int64      `json:"farm_id"`
	Scope      string     `json:"scope"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	Error      string     `json:"error,omitempty"`
}

// NormalizeScope validates API scope values.
func NormalizeScope(scope string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(scope)) {
	case ScopeFieldGuides, ScopePlatformDocs, ScopeOperational, ScopeAll:
		return strings.TrimSpace(strings.ToLower(scope)), nil
	default:
		return "", fmt.Errorf("scope must be field_guides, platform_docs, operational, or all")
	}
}

// ForFarm returns the latest job snapshot for a farm (may be nil).
func ForFarm(farmID int64) *Job {
	mu.Lock()
	defer mu.Unlock()
	j, ok := jobs[farmID]
	if !ok || j == nil {
		return nil
	}
	cp := *j
	return &cp
}

// Start launches an async ingest when embed is reachable on LAN.
func Start(ctx context.Context, q *db.Queries, farmID int64, scope string) (*Job, error) {
	scope, err := NormalizeScope(scope)
	if err != nil {
		return nil, err
	}
	if farmID <= 0 {
		return nil, errors.New("farm_id is required")
	}

	field := farmguardian.BuildFieldAssistantHealth(ctx, nil, 0, 0)
	if !field.EmbeddingConfigured {
		return nil, errors.New("EMBEDDING_API_KEY is not configured")
	}
	embBase := strings.TrimSpace(os.Getenv("EMBEDDING_BASE_URL"))
	if embBase == "" || !farmguardian.IsLocalInferenceURL(embBase) {
		return nil, errors.New("re-ingest requires a LAN/local EMBEDDING_BASE_URL")
	}
	embKey := strings.TrimSpace(os.Getenv("EMBEDDING_API_KEY"))
	if err := ai.VerifyChatBackend(ctx, embBase, embKey); err != nil {
		return nil, fmt.Errorf("embedding backend unreachable: %w", err)
	}

	mu.Lock()
	if existing, ok := jobs[farmID]; ok && existing != nil && existing.Status == StatusRunning {
		cp := *existing
		mu.Unlock()
		return &cp, nil
	}
	job := &Job{
		FarmID:    farmID,
		Scope:     scope,
		Status:    StatusRunning,
		StartedAt: time.Now().UTC(),
	}
	jobs[farmID] = job
	mu.Unlock()

	go run(context.WithoutCancel(ctx), q, job, repoRoot())
	return job, nil
}

func repoRoot() string {
	if v := strings.TrimSpace(os.Getenv("GR33N_REPO_ROOT")); v != "" {
		return v
	}
	return "."
}

func run(ctx context.Context, q *db.Queries, job *Job, repoRoot string) {
	ctx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	var runErr error
	defer func() {
		mu.Lock()
		defer mu.Unlock()
		if j, ok := jobs[job.FarmID]; ok && j != nil {
			fin := time.Now().UTC()
			j.FinishedAt = &fin
			if runErr != nil {
				j.Status = StatusFailed
				j.Error = runErr.Error()
			} else {
				j.Status = StatusDone
			}
		}
	}()

	emb, err := embed.NewOpenAICompatibleFromEnv()
	if err != nil {
		runErr = err
		return
	}
	w := &ingest.Worker{Q: q, Embedder: emb}

	switch job.Scope {
	case ScopeFieldGuides:
		runErr = ingestFieldGuides(ctx, w, job.FarmID, repoRoot)
	case ScopePlatformDocs:
		_, runErr = w.IngestPlatformDocs(ctx, job.FarmID, repoRoot, "")
	case ScopeOperational:
		runErr = ingestOperational(ctx, w, job.FarmID)
	case ScopeAll:
		if err := ingestFieldGuides(ctx, w, job.FarmID, repoRoot); err != nil {
			runErr = err
			return
		}
		if _, err := w.IngestPlatformDocs(ctx, job.FarmID, repoRoot, ""); err != nil {
			runErr = err
			return
		}
		runErr = ingestOperational(ctx, w, job.FarmID)
	default:
		runErr = fmt.Errorf("unsupported scope %q", job.Scope)
	}

	if runErr != nil {
		slog.Warn("rag: reingest failed", "farm_id", job.FarmID, "scope", job.Scope, "err", runErr)
		return
	}
	slog.Info("rag: reingest done", "farm_id", job.FarmID, "scope", job.Scope)
}

func ingestFieldGuides(ctx context.Context, w *ingest.Worker, farmID int64, repoRoot string) error {
	if _, err := w.IngestFieldGuides(ctx, farmID, repoRoot, ""); err != nil {
		return err
	}
	_, err := w.IngestSymptomGuidesFromDB(ctx, farmID)
	return err
}

func ingestOperational(ctx context.Context, w *ingest.Worker, farmID int64) error {
	type step struct {
		name string
		fn   func() (int, error)
	}
	steps := []step{
		{"tasks", func() (int, error) { return w.IngestFarmTasks(ctx, farmID, nil) }},
		{"crop_cycles", func() (int, error) { return w.IngestFarmCropCycles(ctx, farmID, nil) }},
		{"programs", func() (int, error) { return w.IngestFarmFertigationPrograms(ctx, farmID, nil) }},
		{"schedules", func() (int, error) { return w.IngestFarmSchedules(ctx, farmID, nil) }},
		{"automation_rules", func() (int, error) { return w.IngestFarmAutomationRules(ctx, farmID, nil) }},
		{"executable_actions", func() (int, error) { return w.IngestFarmExecutableActions(ctx, farmID) }},
		{"input_definitions", func() (int, error) { return w.IngestFarmInputDefinitions(ctx, farmID, nil) }},
		{"input_batches", func() (int, error) { return w.IngestFarmInputBatches(ctx, farmID, nil) }},
	}
	for _, s := range steps {
		if _, err := s.fn(); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}
	if _, err := w.IngestFarmAutomationRuns(ctx, farmID, 500, 0, nil); err != nil {
		return fmt.Errorf("automation_runs: %w", err)
	}
	if _, err := w.IngestFarmCostTransactions(ctx, farmID, 500, 0, nil); err != nil {
		return fmt.Errorf("cost_transactions: %w", err)
	}
	if _, err := w.IngestFarmAlertNotifications(ctx, farmID, 500, 0, nil); err != nil {
		return fmt.Errorf("alerts: %w", err)
	}
	return nil
}
