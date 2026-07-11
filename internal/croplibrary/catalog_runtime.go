package croplibrary

import (
	"context"
	"fmt"
	"sync"
)

var (
	runtimeCatalogMu       sync.RWMutex
	runtimeCatalogQuerier  CatalogQuerier
)

// SetRuntimeCatalogQuerier wires the API DB pool for CROP_CATALOG_SOURCE=db.
// Call once at process startup before DefaultCatalog or Guardian registry init.
func SetRuntimeCatalogQuerier(q CatalogQuerier) {
	runtimeCatalogMu.Lock()
	runtimeCatalogQuerier = q
	runtimeCatalogMu.Unlock()
}

func runtimeCatalogQuerierOrNil() CatalogQuerier {
	runtimeCatalogMu.RLock()
	defer runtimeCatalogMu.RUnlock()
	return runtimeCatalogQuerier
}

// loadDefaultCatalog loads YAML or DB catalog (used by DefaultCatalog).
func loadDefaultCatalog() (*Catalog, error) {
	if CatalogSource() == "db" {
		q := runtimeCatalogQuerierOrNil()
		if q == nil {
			return nil, fmt.Errorf("CROP_CATALOG_SOURCE=db requires database querier")
		}
		return LoadCatalogFromDB(context.Background(), q)
	}
	root, err := FindRepoRoot()
	if err != nil {
		return nil, err
	}
	return LoadCatalog(root, DefaultCatalogPath)
}
