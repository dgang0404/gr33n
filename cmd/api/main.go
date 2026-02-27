package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://davidg@/gr33n?host=/var/run/postgresql")
	port  := getEnv("PORT", "8080")

	pool, err := connectDB(dbURL)
	if err != nil {
		log.Fatalf("❌ Could not connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("✅ Connected to gr33n database")

	mux := http.NewServeMux()
	registerRoutes(mux, pool)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("🌱 gr33n API running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

func connectDB(dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	config.MaxConns        = 20
	config.MinConns        = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	var pool *pgxpool.Pool
	var lastErr error
	for i := range 5 {
		log.Printf("⏳ Waiting for database... attempt %d/5", i+1)
		pool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			lastErr = err
			log.Printf("   ↳ Pool create failed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if pingErr := pool.Ping(context.Background()); pingErr != nil {
			lastErr = pingErr
			log.Printf("   ↳ Ping failed: %v", pingErr)
			time.Sleep(2 * time.Second)
			continue
		}
		return pool, nil
	}
	return nil, fmt.Errorf("could not reach database after 5 attempts: %w", lastErr)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
