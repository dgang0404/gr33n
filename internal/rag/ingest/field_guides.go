package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pgvector/pgvector-go"

	db "gr33n-api/internal/db"
	"gr33n-api/internal/farmguardian"
)

const (
	SourceTypeFieldGuide      = "field_guide"
	metadataModuleFieldGuide  = "field_guide"
	defaultFieldGuideManifest = "docs/rag/field-guide-manifest.yaml"
)

// FieldGuideManifest lists authored field/trades markdown under docs/field-guides/.
type FieldGuideManifest struct {
	Version      int
	Include      []string
	ExcludeGlobs []string
	DocsRoot     string
	ManifestPath string
}

// FieldGuideDryRun summarizes a manifest scan without embeddings.
type FieldGuideDryRun struct {
	Files       []FieldGuideFileSummary
	TotalChunks int
}

// FieldGuideFileSummary is one manifest entry with chunk estimate.
type FieldGuideFileSummary struct {
	RelPath  string
	SourceID int64
	Bytes    int
	Chunks   int
	Domain   string
	Safety   string
}

// LoadFieldGuideManifest reads docs/rag/field-guide-manifest.yaml (or manifestPath).
func LoadFieldGuideManifest(repoRoot, manifestPath string) (FieldGuideManifest, error) {
	if strings.TrimSpace(manifestPath) == "" {
		manifestPath = filepath.Join(repoRoot, defaultFieldGuideManifest)
	} else if !filepath.IsAbs(manifestPath) {
		manifestPath = filepath.Join(repoRoot, manifestPath)
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return FieldGuideManifest{}, err
	}
	include, exclude, err := parseSimpleYAMLListManifest(string(data))
	if err != nil {
		return FieldGuideManifest{}, err
	}
	docsRoot := filepath.Join(repoRoot, "docs", "field-guides")
	return FieldGuideManifest{
		Version:      1,
		Include:      include,
		ExcludeGlobs: exclude,
		DocsRoot:     docsRoot,
		ManifestPath: manifestPath,
	}, nil
}

// ResolveFieldGuideFiles returns readable markdown files from the manifest.
func (m FieldGuideManifest) ResolveFieldGuideFiles() ([]FieldGuideFileSummary, error) {
	var out []FieldGuideFileSummary
	seen := make(map[string]struct{})
	for _, rel := range m.Include {
		rel = strings.TrimSpace(strings.TrimPrefix(rel, "field-guides/"))
		if rel == "" {
			continue
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		if excludedByGlob(rel, m.ExcludeGlobs) {
			continue
		}
		abs := filepath.Join(m.DocsRoot, rel)
		data, err := os.ReadFile(abs)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", rel, err)
		}
		body, meta := splitYAMLFrontmatter(string(data))
		text := strings.TrimSpace(body)
		chunks := chunkMarkdown(text)
		if len(chunks) == 0 {
			continue
		}
		domain, safety := fieldGuideMetaDefaults(rel, meta)
		out = append(out, FieldGuideFileSummary{
			RelPath:  rel,
			SourceID: FieldGuideSourceID(rel),
			Bytes:    len(data),
			Chunks:   len(chunks),
			Domain:   domain,
			Safety:   safety,
		})
		seen[rel] = struct{}{}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("manifest resolved zero readable files")
	}
	return out, nil
}

// FieldGuideQuerier loads published agronomy field guides from DB.
type FieldGuideQuerier interface {
	ListAgronomyFieldGuides(ctx context.Context) ([]db.Gr33ncropsAgronomyFieldGuide, error)
}

// DryRunFieldGuides returns chunk estimates from files or DB (AGRONOMY_FIELD_GUIDES_SOURCE).
func DryRunFieldGuides(ctx context.Context, q FieldGuideQuerier, repoRoot, manifestPath string) (FieldGuideDryRun, error) {
	if FieldGuidesSource() == "db" {
		if q == nil {
			return FieldGuideDryRun{}, fmt.Errorf("AGRONOMY_FIELD_GUIDES_SOURCE=db requires database querier")
		}
		return DryRunFieldGuidesFromDB(ctx, q)
	}
	logFieldGuideFileDeprecation()
	m, err := LoadFieldGuideManifest(repoRoot, manifestPath)
	if err != nil {
		return FieldGuideDryRun{}, err
	}
	files, err := m.ResolveFieldGuideFiles()
	if err != nil {
		return FieldGuideDryRun{}, err
	}
	total := 0
	for _, f := range files {
		total += f.Chunks
	}
	return FieldGuideDryRun{Files: files, TotalChunks: total}, nil
}

// IngestFieldGuides embeds field guides into rag_embedding_chunks (source_type field_guide).
func (w *Worker) IngestFieldGuides(ctx context.Context, farmID int64, repoRoot, manifestPath string) (int, error) {
	if w == nil || w.Q == nil || w.Embedder == nil {
		return 0, fmt.Errorf("ingest worker not configured")
	}
	if FieldGuidesSource() == "db" {
		return w.IngestFieldGuidesFromDB(ctx, farmID)
	}
	logFieldGuideFileDeprecation()
	m, err := LoadFieldGuideManifest(repoRoot, manifestPath)
	if err != nil {
		return 0, err
	}
	files, err := m.ResolveFieldGuideFiles()
	if err != nil {
		return 0, err
	}
	total := 0
	for _, f := range files {
		abs := filepath.Join(m.DocsRoot, f.RelPath)
		data, err := os.ReadFile(abs)
		if err != nil {
			return total, fmt.Errorf("read %s: %w", f.RelPath, err)
		}
		body, meta := splitYAMLFrontmatter(string(data))
		chunks := chunkMarkdown(strings.TrimSpace(body))
		domain, safety := fieldGuideMetaDefaults(f.RelPath, meta)
		n, err := w.upsertFieldGuideFile(ctx, farmID, f.RelPath, f.SourceID, chunks, domain, safety)
		if err != nil {
			return total, fmt.Errorf("%s: %w", f.RelPath, err)
		}
		total += n
	}
	return total, nil
}

func (w *Worker) upsertFieldGuideFile(ctx context.Context, farmID int64, relPath string, sourceID int64, chunks []string, domain, safety string) (int, error) {
	if len(chunks) == 0 {
		return 0, nil
	}
	if err := w.Q.DeleteRagChunksByFarmSource(ctx, db.DeleteRagChunksByFarmSourceParams{
		FarmID:     farmID,
		SourceType: SourceTypeFieldGuide,
		SourceID:   sourceID,
	}); err != nil {
		return 0, err
	}
	texts := make([]string, len(chunks))
	for i, ch := range chunks {
		texts[i] = FieldGuideDocument(relPath, ch, i, len(chunks))
	}
	vecs, err := w.Embedder.Embed(ctx, texts)
	if err != nil {
		return 0, err
	}
	if len(vecs) != len(texts) {
		return 0, fmt.Errorf("embed count %d != chunk count %d", len(vecs), len(texts))
	}
	meta := fieldGuideMetadata(relPath, domain, safety)
	modelID := w.Embedder.ModelID()
	n := 0
	for i, text := range texts {
		_, err := w.Q.UpsertRagEmbeddingChunk(ctx, db.UpsertRagEmbeddingChunkParams{
			FarmID:      farmID,
			SourceType:  SourceTypeFieldGuide,
			SourceID:    sourceID,
			ChunkIndex:  int32(i),
			ContentText: text,
			Embedding:   pgvector.NewVector(vecs[i]),
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

// FieldGuideSourceID returns a stable positive int64 id for a field-guides-relative path.
func FieldGuideSourceID(relPath string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte("field_guide:" + strings.ToLower(strings.TrimSpace(relPath))))
	id := int64(h.Sum64() & 0x7fffffffffffffff)
	if id == 0 {
		return 1
	}
	return id
}

// FieldGuideDocument formats one chunk for embedding/retrieval.
func FieldGuideDocument(relPath, chunk string, chunkIndex, chunkTotal int) string {
	var b strings.Builder
	b.WriteString("field_guide\n")
	b.WriteString("doc_path: field-guides/")
	b.WriteString(relPath)
	if chunkTotal > 1 {
		fmt.Fprintf(&b, "\nchunk: %d/%d", chunkIndex+1, chunkTotal)
	}
	b.WriteString("\n\n")
	b.WriteString(strings.TrimSpace(chunk))
	return b.String()
}

func fieldGuideMetadata(relPath, domain, safety string) []byte {
	m := map[string]string{
		"module":      metadataModuleFieldGuide,
		"doc_path":    "field-guides/" + relPath,
		"domain":      domain,
		"safety_tier": safety,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return emptyJSON()
	}
	return b
}

func fieldGuideMetaDefaults(relPath string, front map[string]string) (domain, safety string) {
	domain = strings.TrimSpace(front["domain"])
	safety = normalizeSafetyTier(front["safety_tier"])
	if domain == "" {
		switch {
		case strings.Contains(relPath, "electrical"):
			domain = "electrical"
		case strings.Contains(relPath, "irrigation"), strings.Contains(relPath, "plumb"):
			domain = "plumbing"
		case strings.Contains(relPath, "sensor"):
			domain = "sensor"
		case strings.Contains(relPath, "relay"), strings.Contains(relPath, "actuator"):
			domain = "actuator"
		case strings.Contains(relPath, "pi"):
			domain = "pi"
		default:
			domain = "general"
		}
	}
	if safety == "" {
		if domain == "electrical" || strings.Contains(relPath, "electrical-safety") {
			safety = farmguardian.SafetyTierCaution
		} else {
			safety = farmguardian.SafetyTierSafe
		}
	}
	return domain, safety
}

func normalizeSafetyTier(s string) string {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case farmguardian.SafetyTierSafe, farmguardian.SafetyTierCaution, farmguardian.SafetyTierQualifiedPersonRequired:
		return strings.TrimSpace(strings.ToLower(s))
	default:
		return ""
	}
}

// splitYAMLFrontmatter parses optional --- delimited YAML key: value headers.
func splitYAMLFrontmatter(raw string) (body string, meta map[string]string) {
	raw = strings.TrimPrefix(raw, "\ufeff")
	if !strings.HasPrefix(raw, "---") {
		return raw, nil
	}
	rest := strings.TrimPrefix(raw, "---")
	rest = strings.TrimPrefix(rest, "\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return raw, nil
	}
	header := rest[:end]
	body = strings.TrimSpace(rest[end+len("\n---"):])
	meta = make(map[string]string)
	for _, line := range strings.Split(header, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		meta[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return body, meta
}

var fieldGuideFileDeprecationOnce sync.Once

func logFieldGuideFileDeprecation() {
	fieldGuideFileDeprecationOnce.Do(func() {
		log.Printf("warning: AGRONOMY_FIELD_GUIDES_SOURCE=file is deprecated — migrate and use db (default). See docs/crop-catalog-db-cutover-runbook.md")
	})
}
