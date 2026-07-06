// Phase 135 — RAG corpus freshness and staleness tiers.

package farmguardian

import (
	"time"

	db "gr33n-api/internal/db"
)

const (
	FreshnessFresh = "fresh" // < 24h since last ingest
	FreshnessAging = "aging" // 24h–7d
	FreshnessStale = "stale" // > 7d or unknown age
	FreshnessEmpty = "empty" // zero chunks

	StalenessOK                = "ok"
	StalenessFieldGuideEmpty   = "field_guide_empty"
	StalenessOperationalAging  = "operational_aging"
	StalenessOperationalStale  = "operational_stale"
)

// CorpusHealth is the operator-facing RAG corpus block (Phase 135).
type CorpusHealth struct {
	FieldGuideChunks          int64      `json:"field_guide_chunks"`
	FieldGuideLastIngestedAt  *time.Time `json:"field_guide_last_ingested_at,omitempty"`
	PlatformDocChunks         int64      `json:"platform_doc_chunks"`
	PlatformLastIngestedAt    *time.Time `json:"platform_last_ingested_at,omitempty"`
	OperationalChunks         int64      `json:"operational_chunks"`
	OperationalLastIngestedAt *time.Time `json:"operational_last_ingested_at,omitempty"`
	FieldGuideFreshness       string     `json:"field_guide_freshness"`
	PlatformFreshness         string     `json:"platform_freshness"`
	OperationalFreshness      string     `json:"operational_freshness"`
	Staleness                 string     `json:"staleness"`
}

// CorpusStatsInput carries DB aggregates for BuildCorpusHealth.
type CorpusStatsInput struct {
	FieldGuideChunks          int64
	FieldGuideLastIngestedAt  *time.Time
	PlatformDocChunks         int64
	PlatformLastIngestedAt    *time.Time
	OperationalChunks         int64
	OperationalLastIngestedAt *time.Time
}

// CorpusStatsFromRow maps sqlc GetRagCorpusStatsByFarm into CorpusStatsInput.
func CorpusStatsFromRow(row db.GetRagCorpusStatsByFarmRow) CorpusStatsInput {
	return CorpusStatsInput{
		FieldGuideChunks:          row.FieldGuideChunks,
		FieldGuideLastIngestedAt:  coerceTimePtr(row.FieldGuideLastIngestedAt),
		PlatformDocChunks:         row.PlatformDocChunks,
		PlatformLastIngestedAt:    coerceTimePtr(row.PlatformLastIngestedAt),
		OperationalChunks:         row.OperationalChunks,
		OperationalLastIngestedAt: coerceTimePtr(row.OperationalLastIngestedAt),
	}
}

func coerceTimePtr(v any) *time.Time {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case time.Time:
		return timestamptzPtr(t)
	default:
		return nil
	}
}

func timestamptzPtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	utc := t.UTC()
	return &utc
}

// TierFreshness classifies a corpus tier by chunk count and last ingest time.
func TierFreshness(chunkCount int64, lastIngested *time.Time, now time.Time) string {
	if chunkCount == 0 {
		return FreshnessEmpty
	}
	if lastIngested == nil {
		return FreshnessStale
	}
	age := now.Sub(*lastIngested)
	if age < 24*time.Hour {
		return FreshnessFresh
	}
	if age < 7*24*time.Hour {
		return FreshnessAging
	}
	return FreshnessStale
}

// BuildCorpusHealth applies staleness rules for GET /v1/chat/health.
func BuildCorpusHealth(in CorpusStatsInput, now time.Time) CorpusHealth {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	out := CorpusHealth{
		FieldGuideChunks:          in.FieldGuideChunks,
		FieldGuideLastIngestedAt:  in.FieldGuideLastIngestedAt,
		PlatformDocChunks:         in.PlatformDocChunks,
		PlatformLastIngestedAt:    in.PlatformLastIngestedAt,
		OperationalChunks:         in.OperationalChunks,
		OperationalLastIngestedAt: in.OperationalLastIngestedAt,
	}
	out.FieldGuideFreshness = TierFreshness(in.FieldGuideChunks, in.FieldGuideLastIngestedAt, now)
	out.PlatformFreshness = TierFreshness(in.PlatformDocChunks, in.PlatformLastIngestedAt, now)
	out.OperationalFreshness = TierFreshness(in.OperationalChunks, in.OperationalLastIngestedAt, now)

	switch {
	case in.FieldGuideChunks == 0 && in.PlatformDocChunks == 0:
		out.Staleness = StalenessFieldGuideEmpty
	case out.OperationalFreshness == FreshnessStale:
		out.Staleness = StalenessOperationalStale
	case out.OperationalFreshness == FreshnessAging:
		out.Staleness = StalenessOperationalAging
	default:
		out.Staleness = StalenessOK
	}
	return out
}

// CorpusWarningMessages returns operator hints for farm counsel mode.
func CorpusWarningMessages(c CorpusHealth, mode string) []string {
	if normalizeWarmupMode(mode) != WarmupModeFarmCounsel {
		return nil
	}
	var msgs []string
	if c.FieldGuideChunks == 0 && c.PlatformDocChunks == 0 {
		msgs = append(msgs, "Field memories not loaded — run guardian-bootstrap-farm or re-ingest from Settings.")
	} else if c.FieldGuideChunks == 0 {
		msgs = append(msgs, "Field guide memories are empty — re-ingest field guides from Settings.")
	}
	switch c.Staleness {
	case StalenessOperationalStale:
		msgs = append(msgs, "Operational memories are stale (>7d) — re-ingest from Settings for fresher farm context.")
	case StalenessOperationalAging:
		msgs = append(msgs, "Operational memories are aging — consider re-ingest from Settings.")
	}
	return msgs
}
