package croplibrary

import (
	"sort"
	"strings"
	"sync"
	"unicode"
)

// MentionKind classifies a resolved crop mention.
type MentionKind int

const (
	MentionCrop MentionKind = iota
	MentionUnsupported
	MentionUnknown
)

// ResolvedMention is a canonical crop or unsupported key found in text.
type ResolvedMention struct {
	Key          string
	DisplayName  string
	Kind         MentionKind
	Reason       string
	CousinOf     *string
	MatchedTerm  string
}

var (
	defaultCatalogOnce sync.Once
	defaultCatalog     *Catalog
	defaultCatalogErr  error
)

// DefaultCatalog loads the platform crop catalog once per process (YAML or DB).
func DefaultCatalog() (*Catalog, error) {
	defaultCatalogOnce.Do(func() {
		defaultCatalog, defaultCatalogErr = loadDefaultCatalog()
	})
	return defaultCatalog, defaultCatalogErr
}

// Registry indexes crop keys, aliases, and unsupported mentions for Guardian/UI.
type Registry struct {
	catalog      *Catalog
	termToCrop   map[string]string
	termToUnsup  map[string]string
	cropByKey    map[string]Crop
	unsupByKey   map[string]UnsupportedCrop
	termsByLen   []string
}

// NewRegistry builds a mention index from a validated catalog.
func NewRegistry(cat *Catalog) *Registry {
	r := &Registry{
		catalog:     cat,
		termToCrop:  make(map[string]string),
		termToUnsup: make(map[string]string),
		cropByKey:   make(map[string]Crop, len(cat.Crops)),
		unsupByKey:  make(map[string]UnsupportedCrop, len(cat.Unsupported)),
	}
	for _, crop := range cat.Crops {
		r.cropByKey[crop.Key] = crop
		r.addCropTerm(crop.Key, crop.Key)
		for _, a := range crop.Aliases {
			r.addCropTerm(a, crop.Key)
		}
	}
	for _, u := range cat.Unsupported {
		r.unsupByKey[u.Key] = u
		r.addUnsupTerm(u.Key, u.Key)
		for _, a := range u.Aliases {
			r.addUnsupTerm(a, u.Key)
		}
	}
	for alias, target := range cat.Aliases {
		if _, cropOK := r.cropByKey[target]; cropOK {
			r.addCropTerm(alias, target)
		} else if _, unsupOK := r.unsupByKey[target]; unsupOK {
			r.addUnsupTerm(alias, target)
		}
	}
	seen := make(map[string]struct{})
	for term := range r.termToCrop {
		seen[term] = struct{}{}
	}
	for term := range r.termToUnsup {
		seen[term] = struct{}{}
	}
	r.termsByLen = make([]string, 0, len(seen))
	for term := range seen {
		r.termsByLen = append(r.termsByLen, term)
	}
	sort.Slice(r.termsByLen, func(i, j int) bool {
		if len(r.termsByLen[i]) != len(r.termsByLen[j]) {
			return len(r.termsByLen[i]) > len(r.termsByLen[j])
		}
		return r.termsByLen[i] < r.termsByLen[j]
	})
	return r
}

func (r *Registry) addCropTerm(term, cropKey string) {
	term = strings.ToLower(strings.TrimSpace(term))
	if term != "" {
		r.termToCrop[term] = cropKey
	}
}

func (r *Registry) addUnsupTerm(term, unsupKey string) {
	term = strings.ToLower(strings.TrimSpace(term))
	if term != "" {
		r.termToUnsup[term] = unsupKey
	}
}

// ResolveTerm maps one mention to crop or unsupported key.
func (r *Registry) ResolveTerm(term string) (ResolvedMention, bool) {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return ResolvedMention{}, false
	}
	if key, ok := r.termToCrop[term]; ok {
		c := r.cropByKey[key]
		return ResolvedMention{
			Key:         key,
			DisplayName: c.DisplayName,
			Kind:        MentionCrop,
			MatchedTerm: term,
		}, true
	}
	if key, ok := r.termToUnsup[term]; ok {
		u := r.unsupByKey[key]
		display := strings.TrimSpace(u.DisplayName)
		if display == "" {
			display = key
		}
		return ResolvedMention{
			Key:         key,
			DisplayName: display,
			Kind:        MentionUnsupported,
			Reason:      u.Reason,
			CousinOf:    u.CousinOf,
			MatchedTerm: term,
		}, true
	}
	return ResolvedMention{}, false
}

// FindMentions scans text for catalog crop keys, aliases, and unsupported names.
func (r *Registry) FindMentions(text string) []ResolvedMention {
	lower := strings.ToLower(text)
	var out []ResolvedMention
	seen := make(map[string]struct{})
	for _, term := range r.termsByLen {
		if !containsTerm(lower, term) {
			continue
		}
		m, ok := r.ResolveTerm(term)
		if !ok {
			continue
		}
		id := m.Key + "|" + string(rune(m.Kind))
		if m.Kind == MentionCrop {
			id = "crop:" + m.Key
		} else {
			id = "unsup:" + m.Key
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, m)
	}
	return out
}

func containsTerm(lowerText, term string) bool {
	if term == "" {
		return false
	}
	idx := 0
	for {
		i := strings.Index(lowerText[idx:], term)
		if i < 0 {
			return false
		}
		start := idx + i
		end := start + len(term)
		if isTermBoundary(lowerText, start, end) {
			return true
		}
		idx = start + 1
		if idx >= len(lowerText) {
			return false
		}
	}
}

func isTermBoundary(s string, start, end int) bool {
	if start > 0 {
		if r := rune(s[start-1]); unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return false
		}
	}
	if end < len(s) {
		if r := rune(s[end]); unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return false
		}
	}
	return true
}
