package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"gr33n-api/internal/insertcommonsreceiver"
)

func loadDotEnv() {
	var paths []string
	for _, p := range []string{".env", ".env.local"} {
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, p)
		}
	}
	if len(paths) == 0 {
		return
	}
	if err := godotenv.Load(paths...); err != nil {
		log.Printf("warning: env files %v: %v", paths, err)
		return
	}
	log.Printf("Loaded config from %s", strings.Join(paths, ", "))
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func connectDB(dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	config.MaxConns = 10
	config.MinConns = 1
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	var pool *pgxpool.Pool
	var lastErr error
	for i := range 5 {
		log.Printf("waiting for database... attempt %d/5", i+1)
		pool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			lastErr = err
			time.Sleep(2 * time.Second)
			continue
		}
		if pingErr := pool.Ping(context.Background()); pingErr != nil {
			lastErr = pingErr
			time.Sleep(2 * time.Second)
			continue
		}
		return pool, nil
	}
	return nil, fmt.Errorf("could not reach database after 5 attempts: %w", lastErr)
}

func main() {
	loadDotEnv()

	dbURL := getEnv("DATABASE_URL", "postgres://"+os.Getenv("USER")+"@/gr33n?host=/var/run/postgresql")
	listen := strings.TrimSpace(getEnv("INSERT_COMMONS_RECEIVER_LISTEN", ":8765"))
	secret := strings.TrimSpace(getEnv("INSERT_COMMONS_SHARED_SECRET", ""))
	allowInsecure := strings.EqualFold(getEnv("INSERT_COMMONS_RECEIVER_ALLOW_INSECURE_NO_AUTH", ""), "true")
	retentionDays := 90
	if s := strings.TrimSpace(getEnv("INSERT_COMMONS_RECEIVER_RETENTION_DAYS", "90")); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			log.Fatalf("INSERT_COMMONS_RECEIVER_RETENTION_DAYS must be a non-negative integer (got %q)", s)
		}
		retentionDays = n
	}

	if secret == "" && !allowInsecure {
		log.Fatal("Set INSERT_COMMONS_SHARED_SECRET to match the farm API, or for local pilots only set INSERT_COMMONS_RECEIVER_ALLOW_INSECURE_NO_AUTH=true")
	}
	if allowInsecure {
		log.Printf("warning: INSERT_COMMONS_RECEIVER_ALLOW_INSECURE_NO_AUTH is set; ingest is not authenticated")
	}

	pool, err := connectDB(dbURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	h := insertcommonsreceiver.NewHandler(pool, secret, allowInsecure, retentionDays)
	http.Handle("/", h)

	log.Printf("gr33n Insert Commons receiver on http://localhost%s (POST /v1/ingest, GET /health)", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
