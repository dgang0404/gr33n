// generate-crop-catalog-seed reads crop_library.yaml + field guides and emits catalog seed SQL.
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
	manifestPath := flag.String("manifest", croplibrary.DefaultFieldGuideManifest, "field guide manifest")
	output := flag.String("o", "", "write SQL to file (default stdout)")
	validateOnly := flag.Bool("validate-only", false, "validate sources and exit")
	flag.Parse()

	root, err := filepath.Abs(*repoRoot)
	if err != nil {
		fail(err)
	}

	cat, err := croplibrary.LoadCatalog(root, *catalogPath)
	if err != nil {
		fail(err)
	}
	guides, err := croplibrary.LoadFieldGuideSeeds(root, *manifestPath, cat)
	if err != nil {
		fail(err)
	}
	if *validateOnly {
		fmt.Fprintf(os.Stderr, "crop catalog v%d OK (%d crops, %d unsupported, %d field guides)\n",
			cat.Version, len(cat.Crops), len(cat.Unsupported), len(guides))
		return
	}

	sql := croplibrary.GenerateCatalogSeedSQL(cat, guides)
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
	fmt.Fprintf(os.Stderr, "wrote %s (%d catalog entries, %d guides)\n",
		outPath, len(cat.Crops)+len(cat.Unsupported), len(guides))
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "generate-crop-catalog-seed: %v\n", err)
	os.Exit(1)
}
