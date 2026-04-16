package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"gr33n-api/internal/filestorage"
)

func main() {
	var (
		sourceDir = flag.String("source-dir", "", "existing FILE_STORAGE_DIR root to backfill from")
		dryRun    = flag.Bool("dry-run", false, "report what would be copied without writing target blobs")
		fileType  = flag.String("file-type", "", "optional file_type filter (for example: cost_receipt)")
	)
	flag.Parse()

	if strings.TrimSpace(*sourceDir) == "" {
		log.Fatal("source-dir is required")
	}

	ctx := context.Background()

	pool, err := connectDB(getEnv("DATABASE_URL", "postgres://davidg@/gr33n?host=/var/run/postgresql"))
	if err != nil {
		log.Fatalf("connect DB: %v", err)
	}
	defer pool.Close()

	source, err := filestorage.NewLocal(*sourceDir)
	if err != nil {
		log.Fatalf("source store: %v", err)
	}
	target, cfg, err := filestorage.NewFromEnv(ctx)
	if err != nil {
		log.Fatalf("target store: %v", err)
	}
	if cfg.Backend == "local" && samePath(cfg.LocalRoot, *sourceDir) {
		log.Fatal("target local storage matches source-dir; choose a different target backend/root")
	}

	attachments, err := listAttachments(ctx, pool, strings.TrimSpace(*fileType))
	if err != nil {
		log.Fatalf("list attachments: %v", err)
	}
	if len(attachments) == 0 {
		fmt.Println("No matching file attachments found.")
		return
	}

	log.Printf("backfill starting: source=%s target_backend=%s target_root=%s attachments=%d dry_run=%v", *sourceDir, cfg.Backend, cfg.LocalRoot, len(attachments), *dryRun)
	summary, err := filestorage.MigrateAttachments(ctx, source, target, attachments, *dryRun)
	if err != nil {
		log.Fatalf("backfill failed after %d scanned / %d copied: %v", summary.Scanned, summary.Copied, err)
	}
	log.Printf("backfill complete: scanned=%d copied=%d failed=%d", summary.Scanned, summary.Copied, summary.Failed)
}

func listAttachments(ctx context.Context, pool *pgxpool.Pool, fileType string) ([]filestorage.MigrationAttachment, error) {
	q := `
SELECT id, storage_path, file_name
FROM gr33ncore.file_attachments
WHERE storage_path <> ''
`
	var args []any
	if fileType != "" {
		q += " AND file_type = $1"
		args = append(args, fileType)
	}
	q += " ORDER BY id ASC"

	rows, err := pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []filestorage.MigrationAttachment
	for rows.Next() {
		var att filestorage.MigrationAttachment
		if err := rows.Scan(&att.ID, &att.StoragePath, &att.FileName); err != nil {
			return nil, err
		}
		out = append(out, att)
	}
	return out, rows.Err()
}

func connectDB(dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	return pgxpool.NewWithConfig(context.Background(), config)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func samePath(a, b string) bool {
	aa := strings.TrimRight(strings.TrimSpace(a), "/")
	bb := strings.TrimRight(strings.TrimSpace(b), "/")
	return aa != "" && aa == bb
}
