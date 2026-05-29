package ingest

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/pgvector/pgvector-go"

	db "gr33n-api/internal/db"
)

const (
	SourceTypePlatformDoc       = "platform_doc"
	metadataModulePlatformDoc   = "platform_doc"
	platformDocMaxChunkRunes    = 1800
	platformDocSplitThreshold   = 2200
	defaultPlatformDocManifest  = "docs/rag/platform-doc-manifest.yaml"
)

// PlatformDocManifest lists markdown files under docs/ to embed as platform_doc chunks.
type PlatformDocManifest struct {
	Version       int
	Include       []string
	ExcludeGlobs  []string
	DocsRoot      string // absolute path to docs/
	ManifestPath  string
}

// PlatformDocDryRun summarizes a manifest scan without embeddings.
type PlatformDocDryRun struct {
	Files       []PlatformDocFileSummary
	TotalChunks int
}

// PlatformDocFileSummary is one manifest entry with chunk estimate.
type PlatformDocFileSummary struct {
	RelPath   string
	SourceID  int64
	Bytes     int
	Chunks    int
}

// LoadPlatformDocManifest reads docs/rag/platform-doc-manifest.yaml (or manifestPath).
func LoadPlatformDocManifest(repoRoot, manifestPath string) (PlatformDocManifest, error) {
	if strings.TrimSpace(manifestPath) == "" {
		manifestPath = filepath.Join(repoRoot, defaultPlatformDocManifest)
	} else if !filepath.IsAbs(manifestPath) {
		manifestPath = filepath.Join(repoRoot, manifestPath)
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return PlatformDocManifest{}, err
	}
	include, exclude, err := parseSimpleYAMLListManifest(string(data))
	if err != nil {
		return PlatformDocManifest{}, err
	}
	docsRoot := filepath.Join(repoRoot, "docs")
	return PlatformDocManifest{
		Version:      1,
		Include:      include,
		ExcludeGlobs: exclude,
		DocsRoot:     docsRoot,
		ManifestPath: manifestPath,
	}, nil
}

// ResolvePlatformDocFiles returns readable markdown files from the manifest.
func (m PlatformDocManifest) ResolvePlatformDocFiles() ([]PlatformDocFileSummary, error) {
	var out []PlatformDocFileSummary
	seen := make(map[string]struct{})
	for _, rel := range m.Include {
		rel = strings.TrimSpace(strings.TrimPrefix(rel, "docs/"))
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
		text := string(data)
		chunks := chunkMarkdown(text)
		if len(chunks) == 0 {
			continue
		}
		out = append(out, PlatformDocFileSummary{
			RelPath:  rel,
			SourceID: PlatformDocSourceID(rel),
			Bytes:    len(data),
			Chunks:   len(chunks),
		})
		seen[rel] = struct{}{}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("manifest resolved zero readable files")
	}
	return out, nil
}

// DryRunPlatformDocs returns file list + chunk estimates.
func DryRunPlatformDocs(repoRoot, manifestPath string) (PlatformDocDryRun, error) {
	m, err := LoadPlatformDocManifest(repoRoot, manifestPath)
	if err != nil {
		return PlatformDocDryRun{}, err
	}
	files, err := m.ResolvePlatformDocFiles()
	if err != nil {
		return PlatformDocDryRun{}, err
	}
	total := 0
	for _, f := range files {
		total += f.Chunks
	}
	return PlatformDocDryRun{Files: files, TotalChunks: total}, nil
}

// IngestPlatformDocs embeds manifest markdown into rag_embedding_chunks (source_type platform_doc).
func (w *Worker) IngestPlatformDocs(ctx context.Context, farmID int64, repoRoot, manifestPath string) (int, error) {
	if w == nil || w.Q == nil || w.Embedder == nil {
		return 0, fmt.Errorf("ingest worker not configured")
	}
	m, err := LoadPlatformDocManifest(repoRoot, manifestPath)
	if err != nil {
		return 0, err
	}
	files, err := m.ResolvePlatformDocFiles()
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
		chunks := chunkMarkdown(string(data))
		n, err := w.upsertPlatformDocFile(ctx, farmID, f.RelPath, f.SourceID, chunks)
		if err != nil {
			return total, fmt.Errorf("%s: %w", f.RelPath, err)
		}
		total += n
	}
	return total, nil
}

func (w *Worker) upsertPlatformDocFile(ctx context.Context, farmID int64, relPath string, sourceID int64, chunks []string) (int, error) {
	if len(chunks) == 0 {
		return 0, nil
	}
	if err := w.Q.DeleteRagChunksByFarmSource(ctx, db.DeleteRagChunksByFarmSourceParams{
		FarmID:     farmID,
		SourceType: SourceTypePlatformDoc,
		SourceID:   sourceID,
	}); err != nil {
		return 0, err
	}
	texts := make([]string, len(chunks))
	for i, ch := range chunks {
		texts[i] = PlatformDocument(relPath, ch, i, len(chunks))
	}
	vecs, err := w.Embedder.Embed(ctx, texts)
	if err != nil {
		return 0, err
	}
	if len(vecs) != len(texts) {
		return 0, fmt.Errorf("embed count %d != chunk count %d", len(vecs), len(texts))
	}
	meta := platformDocMetadata(relPath)
	modelID := w.Embedder.ModelID()
	n := 0
	for i, text := range texts {
		_, err := w.Q.UpsertRagEmbeddingChunk(ctx, db.UpsertRagEmbeddingChunkParams{
			FarmID:      farmID,
			SourceType:  SourceTypePlatformDoc,
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

// PlatformDocSourceID returns a stable positive int64 id for a docs-relative path.
func PlatformDocSourceID(relPath string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(relPath))))
	id := int64(h.Sum64() & 0x7fffffffffffffff)
	if id == 0 {
		return 1
	}
	return id
}

// PlatformDocument formats one chunk for embedding/retrieval.
func PlatformDocument(relPath, chunk string, chunkIndex, chunkTotal int) string {
	var b strings.Builder
	b.WriteString("platform_doc\n")
	b.WriteString("doc_path: ")
	b.WriteString(relPath)
	if chunkTotal > 1 {
		fmt.Fprintf(&b, "\nchunk: %d/%d", chunkIndex+1, chunkTotal)
	}
	b.WriteString("\n\n")
	b.WriteString(strings.TrimSpace(chunk))
	return b.String()
}

func platformDocMetadata(relPath string) []byte {
	m := map[string]string{
		"module":   metadataModulePlatformDoc,
		"doc_path": relPath,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return emptyJSON()
	}
	return b
}

func chunkMarkdown(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if utf8.RuneCountInString(text) <= platformDocMaxChunkRunes {
		return []string{text}
	}
	sections := splitMarkdownSections(text)
	var chunks []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		chunks = append(chunks, strings.TrimSpace(cur.String()))
		cur.Reset()
	}
	for _, sec := range sections {
		sec = strings.TrimSpace(sec)
		if sec == "" {
			continue
		}
		if utf8.RuneCountInString(sec) > platformDocSplitThreshold {
			flush()
			chunks = append(chunks, splitByRunes(sec, platformDocMaxChunkRunes)...)
			continue
		}
		sep := ""
		if cur.Len() > 0 {
			sep = "\n\n"
		}
		candidate := cur.String() + sep + sec
		if utf8.RuneCountInString(candidate) > platformDocMaxChunkRunes && cur.Len() > 0 {
			flush()
			cur.WriteString(sec)
		} else {
			cur.Reset()
			cur.WriteString(candidate)
		}
	}
	flush()
	if len(chunks) == 0 {
		return []string{text}
	}
	return chunks
}

func splitMarkdownSections(text string) []string {
	lines := strings.Split(text, "\n")
	var sections []string
	var b strings.Builder
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") && b.Len() > 0 {
			sections = append(sections, b.String())
			b.Reset()
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}
	if b.Len() > 0 {
		sections = append(sections, b.String())
	}
	return sections
}

func splitByRunes(text string, maxRunes int) []string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return []string{text}
	}
	var out []string
	for i := 0; i < len(runes); {
		end := i + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, strings.TrimSpace(string(runes[i:end])))
		i = end
	}
	return out
}

func parseSimpleYAMLListManifest(raw string) (include, exclude []string, err error) {
	var section string
	sc := bufio.NewScanner(strings.NewReader(raw))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, "- ") {
			section = strings.TrimSuffix(line, ":")
			continue
		}
		if strings.HasPrefix(line, "- ") {
			item := strings.TrimSpace(strings.TrimPrefix(line, "- "))
			item = strings.Trim(item, `"'`)
			switch section {
			case "include":
				include = append(include, item)
			case "exclude_globs":
				exclude = append(exclude, item)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	if len(include) == 0 {
		return nil, nil, fmt.Errorf("manifest include list is empty")
	}
	return include, exclude, nil
}

func excludedByGlob(relPath string, patterns []string) bool {
	for _, pat := range patterns {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			continue
		}
		if ok, _ := filepath.Match(pat, relPath); ok {
			return true
		}
		if strings.Contains(pat, "**") {
			suffix := strings.TrimPrefix(pat, "**/")
			if suffix != pat && strings.HasSuffix(relPath, suffix) {
				return true
			}
		}
	}
	return false
}
