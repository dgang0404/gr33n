// generate-crop-seed reads data/crop_library.yaml and emits idempotent SQL (Phase 82 WS4a).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gr33n-api/internal/croplibrary"
)

func main() {
	repoRoot := flag.String("repo-root", ".", "repository root")
	catalogPath := flag.String("catalog", croplibrary.DefaultCatalogPath, "path to crop_library.yaml")
	output := flag.String("o", "", "write SQL to file (default stdout)")
	validateOnly := flag.Bool("validate-only", false, "validate YAML and exit")
	flag.Parse()

	root, err := filepath.Abs(*repoRoot)
	if err != nil {
		fail(err)
	}

	cat, err := croplibrary.LoadCatalog(root, *catalogPath)
	if err != nil {
		fail(err)
	}
	if *validateOnly {
		fmt.Fprintf(os.Stderr, "crop library v%d OK (%d crops, %d unsupported)\n",
			cat.Version, len(cat.Crops), len(cat.Unsupported))
		return
	}

	sql := croplibrary.GenerateSeedSQL(cat)
	if *output == "" {
		fmt.Print(sql)
		return
	}
	outPath, err := filepath.Abs(*output)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(outPath, []byte(sql), 0o644); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d crops with stages)\n", outPath, len(cat.CropsWithStages()))
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "generate-crop-seed: %v\n", err)
	os.Exit(1)
}
