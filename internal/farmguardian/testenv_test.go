package farmguardian

import (
	"os"
	"testing"
)

// Unit tests in this package expect the YAML crop catalog — the default
// runtime mode is db, which needs SetRuntimeCatalogQuerier at API startup.
func TestMain(m *testing.M) {
	_ = os.Setenv("CROP_CATALOG_SOURCE", "yaml")
	os.Exit(m.Run())
}
