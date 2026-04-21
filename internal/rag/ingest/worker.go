package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
// If since is non-nil, only rows with updated_at strictly after since are embedded (incremental poll).
func (w *Worker) IngestFarmTasks(ctx context.Context, farmID int64, since *time.Time) (int, error) {
	var tasks []db.Gr33ncoreTask
	var err error
	if since != nil {
		tasks, err = w.Q.ListTasksByFarmUpdatedAfter(ctx, db.ListTasksByFarmUpdatedAfterParams{
			FarmID:       farmID,
			UpdatedAfter: *since,
		})
	} else {
		tasks, err = w.Q.ListTasksByFarm(ctx, farmID)
	}
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

// IngestFarmAutomationRuns embeds automation_runs (cursor batching for large farms).
// If since is nil, scans by id starting after startAfterID.
// If since is set, scans by executed_at > since with (executed_at, id) keyset paging; startAfterID is ignored.
func (w *Worker) IngestFarmAutomationRuns(ctx context.Context, farmID int64, batchSize int32, startAfterID int64, since *time.Time) (int, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	if since != nil {
		var total int
		page, err := w.Q.ListAutomationRunsByFarmExecutedAfterFirst(ctx, db.ListAutomationRunsByFarmExecutedAfterFirstParams{
			FarmID: farmID,
			Since:  *since,
			Limit:  batchSize,
		})
		for {
			if err != nil {
				return total, err
			}
			if len(page) == 0 {
				break
			}
			docs := make([]string, len(page))
			ids := make([]int64, len(page))
			for i := range page {
				docs[i] = AutomationRunDocument(page[i])
				ids[i] = page[i].ID
			}
			n, err := w.upsertBatch(ctx, farmID, SourceTypeAutomationRun, ids, docs)
			if err != nil {
				return total, err
			}
			total += n
			if int32(len(page)) < batchSize {
				break
			}
			last := page[len(page)-1]
			page, err = w.Q.ListAutomationRunsByFarmExecutedAfterNext(ctx, db.ListAutomationRunsByFarmExecutedAfterNextParams{
				FarmID:           farmID,
				CursorExecutedAt: last.ExecutedAt,
				CursorID:         last.ID,
				Limit:            batchSize,
			})
		}
		return total, nil
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
func (w *Worker) IngestFarmCropCycles(ctx context.Context, farmID int64, since *time.Time) (int, error) {
	var cycles []db.Gr33nfertigationCropCycle
	var err error
	if since != nil {
		cycles, err = w.Q.ListCropCyclesByFarmUpdatedAfter(ctx, db.ListCropCyclesByFarmUpdatedAfterParams{
			FarmID:       farmID,
			UpdatedAfter: *since,
		})
	} else {
		cycles, err = w.Q.ListCropCyclesByFarm(ctx, farmID)
	}
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
func (w *Worker) IngestFarmFertigationPrograms(ctx context.Context, farmID int64, since *time.Time) (int, error) {
	var progs []db.Gr33nfertigationProgram
	var err error
	if since != nil {
		progs, err = w.Q.ListProgramsByFarmUpdatedAfter(ctx, db.ListProgramsByFarmUpdatedAfterParams{
			FarmID:       farmID,
			UpdatedAfter: *since,
		})
	} else {
		progs, err = w.Q.ListProgramsByFarm(ctx, farmID)
	}
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
func (w *Worker) IngestFarmSchedules(ctx context.Context, farmID int64, since *time.Time) (int, error) {
	var rows []db.Gr33ncoreSchedule
	var err error
	if since != nil {
		rows, err = w.Q.ListSchedulesByFarmUpdatedAfter(ctx, db.ListSchedulesByFarmUpdatedAfterParams{
			FarmID:       farmID,
			UpdatedAfter: *since,
		})
	} else {
		rows, err = w.Q.ListSchedulesByFarm(ctx, farmID)
	}
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
func (w *Worker) IngestFarmAutomationRules(ctx context.Context, farmID int64, since *time.Time) (int, error) {
	var rows []db.Gr33ncoreAutomationRule
	var err error
	if since != nil {
		rows, err = w.Q.ListAutomationRulesByFarmUpdatedAfter(ctx, db.ListAutomationRulesByFarmUpdatedAfterParams{
			FarmID:       farmID,
			UpdatedAfter: *since,
		})
	} else {
		rows, err = w.Q.ListAutomationRulesByFarm(ctx, farmID)
	}
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

// IngestFarmCostTransactions embeds cost_transactions (amount/currency omitted in document text).
// If since is nil, scans by id after startAfterID. If since is set, scans updated_at > since with keyset paging; startAfterID is ignored.
func (w *Worker) IngestFarmCostTransactions(ctx context.Context, farmID int64, batchSize int32, startAfterID int64, since *time.Time) (int, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	if since != nil {
		var total int
		page, err := w.Q.ListCostTransactionsByFarmUpdatedAfterFirst(ctx, db.ListCostTransactionsByFarmUpdatedAfterFirstParams{
			FarmID: farmID,
			Since:  *since,
			Limit:  batchSize,
		})
		for {
			if err != nil {
				return total, err
			}
			if len(page) == 0 {
				break
			}
			docs := make([]string, len(page))
			ids := make([]int64, len(page))
			for i := range page {
				docs[i] = CostTransactionDocument(page[i])
				ids[i] = page[i].ID
			}
			n, err := w.upsertBatch(ctx, farmID, SourceTypeCostTransaction, ids, docs)
			if err != nil {
				return total, err
			}
			total += n
			if int32(len(page)) < batchSize {
				break
			}
			last := page[len(page)-1]
			page, err = w.Q.ListCostTransactionsByFarmUpdatedAfterNext(ctx, db.ListCostTransactionsByFarmUpdatedAfterNextParams{
				FarmID:          farmID,
				CursorUpdatedAt: last.UpdatedAt,
				CursorID:        last.ID,
				Limit:           batchSize,
			})
		}
		return total, nil
	}
	var total int
	lastID := startAfterID
	for {
		rows, err := w.Q.ListCostTransactionsByFarmAfterID(ctx, db.ListCostTransactionsByFarmAfterIDParams{
			FarmID: farmID,
			ID:     lastID,
			Limit:  batchSize,
		})
		if err != nil {
			return total, err
		}
		if len(rows) == 0 {
			break
		}
		docs := make([]string, len(rows))
		ids := make([]int64, len(rows))
		for i := range rows {
			docs[i] = CostTransactionDocument(rows[i])
			ids[i] = rows[i].ID
		}
		n, err := w.upsertBatch(ctx, farmID, SourceTypeCostTransaction, ids, docs)
		if err != nil {
			return total, err
		}
		total += n
		lastID = rows[len(rows)-1].ID
		if int32(len(rows)) < batchSize {
			break
		}
	}
	return total, nil
}

// IngestFarmInputDefinitions embeds natural-farming input definitions (no unit cost in text).
func (w *Worker) IngestFarmInputDefinitions(ctx context.Context, farmID int64, since *time.Time) (int, error) {
	var rows []db.Gr33nnaturalfarmingInputDefinition
	var err error
	if since != nil {
		rows, err = w.Q.ListInputDefinitionsByFarmUpdatedAfter(ctx, db.ListInputDefinitionsByFarmUpdatedAfterParams{
			FarmID:       farmID,
			UpdatedAfter: *since,
		})
	} else {
		rows, err = w.Q.ListInputDefinitionsByFarm(ctx, farmID)
	}
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	docs := make([]string, len(rows))
	ids := make([]int64, len(rows))
	for i := range rows {
		docs[i] = InputDefinitionDocument(rows[i])
		ids[i] = rows[i].ID
	}
	return w.upsertBatch(ctx, farmID, SourceTypeInputDefinition, ids, docs)
}

// IngestFarmInputBatches embeds input batches (no quantity / commercial numerics in text).
func (w *Worker) IngestFarmInputBatches(ctx context.Context, farmID int64, since *time.Time) (int, error) {
	var rows []db.Gr33nnaturalfarmingInputBatch
	var err error
	if since != nil {
		rows, err = w.Q.ListInputBatchesByFarmUpdatedAfter(ctx, db.ListInputBatchesByFarmUpdatedAfterParams{
			FarmID:       farmID,
			UpdatedAfter: *since,
		})
	} else {
		rows, err = w.Q.ListInputBatchesByFarm(ctx, farmID)
	}
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	docs := make([]string, len(rows))
	ids := make([]int64, len(rows))
	for i := range rows {
		docs[i] = InputBatchDocument(rows[i])
		ids[i] = rows[i].ID
	}
	return w.upsertBatch(ctx, farmID, SourceTypeInputBatch, ids, docs)
}

// IngestFarmAlertNotifications embeds alerts_notifications.
// If since is nil, scans by id after startAfterID. If since is set, scans created_at > since with keyset paging; startAfterID is ignored.
func (w *Worker) IngestFarmAlertNotifications(ctx context.Context, farmID int64, batchSize int32, startAfterID int64, since *time.Time) (int, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	if since != nil {
		var total int
		page, err := w.Q.ListAlertsByFarmCreatedAfterFirst(ctx, db.ListAlertsByFarmCreatedAfterFirstParams{
			FarmID: farmID,
			Since:  *since,
			Limit:  batchSize,
		})
		for {
			if err != nil {
				return total, err
			}
			if len(page) == 0 {
				break
			}
			docs := make([]string, len(page))
			ids := make([]int64, len(page))
			for i := range page {
				docs[i] = AlertNotificationDocument(page[i])
				ids[i] = page[i].ID
			}
			n, err := w.upsertBatch(ctx, farmID, SourceTypeAlertNotification, ids, docs)
			if err != nil {
				return total, err
			}
			total += n
			if int32(len(page)) < batchSize {
				break
			}
			last := page[len(page)-1]
			page, err = w.Q.ListAlertsByFarmCreatedAfterNext(ctx, db.ListAlertsByFarmCreatedAfterNextParams{
				FarmID:          farmID,
				CursorCreatedAt: last.CreatedAt,
				CursorID:        last.ID,
				Limit:           batchSize,
			})
		}
		return total, nil
	}
	var total int
	lastID := startAfterID
	for {
		rows, err := w.Q.ListAlertsByFarmAfterID(ctx, db.ListAlertsByFarmAfterIDParams{
			FarmID: farmID,
			ID:     lastID,
			Limit:  batchSize,
		})
		if err != nil {
			return total, err
		}
		if len(rows) == 0 {
			break
		}
		docs := make([]string, len(rows))
		ids := make([]int64, len(rows))
		for i := range rows {
			docs[i] = AlertNotificationDocument(rows[i])
			ids[i] = rows[i].ID
		}
		n, err := w.upsertBatch(ctx, farmID, SourceTypeAlertNotification, ids, docs)
		if err != nil {
			return total, err
		}
		total += n
		lastID = rows[len(rows)-1].ID
		if int32(len(rows)) < batchSize {
			break
		}
	}
	return total, nil
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
	case SourceTypeCostTransaction:
		module = metadataModuleCost
	case SourceTypeInputDefinition, SourceTypeInputBatch:
		module = metadataModuleInventory
	case SourceTypeAlertNotification:
		module = metadataModuleAlerts
	}
	m := map[string]string{"module": module}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return b
}
