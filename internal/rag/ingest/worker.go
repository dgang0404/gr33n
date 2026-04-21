package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pgvector/pgvector-go"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/rag/embed"
)

// Worker batches source rows → embed text → pgvector chunks (Phase 24 WS3).
type Worker struct {
	Q        *db.Queries
	Embedder embed.Embedder
}

func emptyJSON() []byte { return []byte("{}") }

// IngestFarmTasks embeds all non-deleted tasks for a farm (chunk_index 0 per task).
func (w *Worker) IngestFarmTasks(ctx context.Context, farmID int64) (int, error) {
	tasks, err := w.Q.ListTasksByFarm(ctx, farmID)
	if err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}
	docs := make([]string, len(tasks))
	ids := make([]int64, len(tasks))
	for i := range tasks {
		docs[i] = TaskDocument(tasks[i])
		ids[i] = tasks[i].ID
	}
	return w.upsertBatch(ctx, farmID, SourceTypeTask, ids, docs)
}

// IngestFarmAutomationRuns embeds automation_runs in id order (cursor batching for large farms).
func (w *Worker) IngestFarmAutomationRuns(ctx context.Context, farmID int64, batchSize int32, startAfterID int64) (int, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	var total int
	lastID := startAfterID
	for {
		runs, err := w.Q.ListAutomationRunsByFarmAfterID(ctx, db.ListAutomationRunsByFarmAfterIDParams{
			FarmID: farmID,
			ID:     lastID,
			Limit:  batchSize,
		})
		if err != nil {
			return total, err
		}
		if len(runs) == 0 {
			break
		}
		docs := make([]string, len(runs))
		ids := make([]int64, len(runs))
		for i := range runs {
			docs[i] = AutomationRunDocument(runs[i])
			ids[i] = runs[i].ID
		}
		n, err := w.upsertBatch(ctx, farmID, SourceTypeAutomationRun, ids, docs)
		if err != nil {
			return total, err
		}
		total += n
		lastID = runs[len(runs)-1].ID
		if int32(len(runs)) < batchSize {
			break
		}
	}
	return total, nil
}

// IngestFarmCropCycles embeds all crop cycles for a farm (chunk_index 0 per row).
func (w *Worker) IngestFarmCropCycles(ctx context.Context, farmID int64) (int, error) {
	cycles, err := w.Q.ListCropCyclesByFarm(ctx, farmID)
	if err != nil {
		return 0, err
	}
	if len(cycles) == 0 {
		return 0, nil
	}
	docs := make([]string, len(cycles))
	ids := make([]int64, len(cycles))
	for i := range cycles {
		docs[i] = CropCycleDocument(cycles[i])
		ids[i] = cycles[i].ID
	}
	return w.upsertBatch(ctx, farmID, SourceTypeCropCycle, ids, docs)
}

// IngestFarmFertigationPrograms embeds fertigation programs for a farm (chunk_index 0 per row).
func (w *Worker) IngestFarmFertigationPrograms(ctx context.Context, farmID int64) (int, error) {
	progs, err := w.Q.ListProgramsByFarm(ctx, farmID)
	if err != nil {
		return 0, err
	}
	if len(progs) == 0 {
		return 0, nil
	}
	docs := make([]string, len(progs))
	ids := make([]int64, len(progs))
	for i := range progs {
		docs[i] = FertigationProgramDocument(progs[i])
		ids[i] = progs[i].ID
	}
	return w.upsertBatch(ctx, farmID, SourceTypeFertigationProgram, ids, docs)
}

// IngestFarmSchedules embeds gr33ncore.schedules for a farm.
func (w *Worker) IngestFarmSchedules(ctx context.Context, farmID int64) (int, error) {
	rows, err := w.Q.ListSchedulesByFarm(ctx, farmID)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	docs := make([]string, len(rows))
	ids := make([]int64, len(rows))
	for i := range rows {
		docs[i] = ScheduleDocument(rows[i])
		ids[i] = rows[i].ID
	}
	return w.upsertBatch(ctx, farmID, SourceTypeSchedule, ids, docs)
}

// IngestFarmAutomationRules embeds gr33ncore.automation_rules for a farm.
func (w *Worker) IngestFarmAutomationRules(ctx context.Context, farmID int64) (int, error) {
	rows, err := w.Q.ListAutomationRulesByFarm(ctx, farmID)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	docs := make([]string, len(rows))
	ids := make([]int64, len(rows))
	for i := range rows {
		docs[i] = AutomationRuleDocument(rows[i])
		ids[i] = rows[i].ID
	}
	return w.upsertBatch(ctx, farmID, SourceTypeAutomationRule, ids, docs)
}

// IngestFarmExecutableActions embeds executable_actions linked to the farm's schedules, rules, or programs.
func (w *Worker) IngestFarmExecutableActions(ctx context.Context, farmID int64) (int, error) {
	rows, err := w.Q.ListExecutableActionsByFarmForRAG(ctx, farmID)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	docs := make([]string, len(rows))
	ids := make([]int64, len(rows))
	for i := range rows {
		docs[i] = ExecutableActionDocument(rows[i])
		ids[i] = rows[i].ID
	}
	return w.upsertBatch(ctx, farmID, SourceTypeExecutableAction, ids, docs)
}

func (w *Worker) upsertBatch(ctx context.Context, farmID int64, sourceType string, sourceIDs []int64, texts []string) (int, error) {
	if len(sourceIDs) != len(texts) || len(texts) == 0 {
		return 0, nil
	}
	nonEmptyIdx := make([]int, 0, len(texts))
	for i, t := range texts {
		if t != "" {
			nonEmptyIdx = append(nonEmptyIdx, i)
		}
	}
	if len(nonEmptyIdx) == 0 {
		return 0, nil
	}
	filteredTexts := make([]string, len(nonEmptyIdx))
	filteredIDs := make([]int64, len(nonEmptyIdx))
	for j, i := range nonEmptyIdx {
		filteredTexts[j] = texts[i]
		filteredIDs[j] = sourceIDs[i]
	}
	vecs, err := w.Embedder.Embed(ctx, filteredTexts)
	if err != nil {
		return 0, err
	}
	if len(vecs) != len(filteredTexts) {
		return 0, fmt.Errorf("embed count %d != text count %d", len(vecs), len(filteredTexts))
	}
	modelID := w.Embedder.ModelID()
	n := 0
	for i := range filteredTexts {
		meta := metadataBytes(sourceType)
		if meta == nil {
			meta = emptyJSON()
		}
		v := pgvector.NewVector(vecs[i])
		_, err := w.Q.UpsertRagEmbeddingChunk(ctx, db.UpsertRagEmbeddingChunkParams{
			FarmID:      farmID,
			SourceType:  sourceType,
			SourceID:    filteredIDs[i],
			ChunkIndex:  0,
			ContentText: filteredTexts[i],
			Embedding:   v,
			ModelID:     modelID,
			Metadata:    meta,
		})
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func metadataBytes(sourceType string) []byte {
	module := metadataModuleCore
	switch sourceType {
	case SourceTypeAutomationRun, SourceTypeSchedule, SourceTypeAutomationRule, SourceTypeExecutableAction:
		module = metadataModuleAutomation
	case SourceTypeCropCycle, SourceTypeFertigationProgram:
		module = metadataModuleFertigation
	}
	m := map[string]string{"module": module}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}
