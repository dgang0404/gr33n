package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	automationworker "gr33n-api/internal/automation"
)

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://davidg@/gr33n?host=/var/run/postgresql")
	port  := getEnv("PORT", "8080")
	jwtSecret = []byte(getEnv("JWT_SECRET", ""))
	piAPIKey  = getEnv("PI_API_KEY", "")
	corsOrigin = getEnv("CORS_ORIGIN", "http://localhost:5173")
	authMode  = getEnv("AUTH_MODE", "dev")
	adminUser    := getEnv("ADMIN_USERNAME", "admin")
	simulationMode := strings.EqualFold(getEnv("AUTOMATION_SIMULATION_MODE", "true"), "true")
	hashFilePath := filepath.Join(os.Getenv("HOME"), ".gr33n", "admin.hash")
	adminHash    := loadPasswordHash(hashFilePath)

	if authMode == "production" {
		if len(jwtSecret) == 0 {
			log.Fatal("AUTH_MODE=production requires JWT_SECRET to be set")
		}
		if piAPIKey == "" {
			log.Fatal("AUTH_MODE=production requires PI_API_KEY to be set")
		}
	}

	pool, err := connectDB(dbURL)
	if err != nil { log.Fatalf("Could not connect to database: %v", err) }
	defer pool.Close()
	log.Println("Connected to gr33n database")
	log.Printf("AUTH_MODE=%s", authMode)
	if len(jwtSecret) == 0 { log.Println("JWT_SECRET not set — JWT auth disabled (dev mode)") } else { log.Println("JWT auth enabled") }
	if piAPIKey == "" { log.Println("PI_API_KEY not set — Pi API key auth disabled (dev mode)") } else { log.Println("Pi API key auth enabled") }
	log.Printf("CORS_ORIGIN=%s", corsOrigin)
	mux := http.NewServeMux()
	var workerOpts []automationworker.WorkerOption
	if cs := getEnv("AUTOMATION_COOLDOWN_SECONDS", ""); cs != "" {
		if n, err := strconv.Atoi(cs); err == nil && n > 0 {
			workerOpts = append(workerOpts, automationworker.WithCooldown(time.Duration(n)*time.Second))
			log.Printf("⏱  Automation cooldown set to %ds", n)
		}
	}
	worker := automationworker.NewWorker(pool, simulationMode, workerOpts...)
	go worker.Start(context.Background())
	log.Printf("🧠 Automation worker started (simulation_mode=%v)", simulationMode)
	registerRoutes(mux, pool, worker, adminUser, adminHash, hashFilePath)
	addr := fmt.Sprintf(":%s", port)
	log.Printf("🌱 gr33n API running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil { log.Fatalf("❌ Server error: %v", err) }
}

func loadPasswordHash(filePath string) []byte {
	if data, err := os.ReadFile(filePath); err == nil {
		hash := []byte(strings.TrimSpace(string(data)))
		if len(hash) > 0 { log.Printf("🔒 Loaded password hash from %s", filePath); return hash }
	}
	if h := getEnv("ADMIN_PASSWORD_HASH", ""); h != "" { log.Println("🔒 Loaded password hash from env"); return []byte(h) }
	return nil
}

func connectDB(dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil { return nil, fmt.Errorf("invalid DATABASE_URL: %w", err) }
	config.MaxConns = 20; config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour; config.MaxConnIdleTime = 30 * time.Minute
	var pool *pgxpool.Pool; var lastErr error
	for i := range 5 {
		log.Printf("⏳ Waiting for database... attempt %d/5", i+1)
		pool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err != nil { lastErr = err; time.Sleep(2 * time.Second); continue }
		if pingErr := pool.Ping(context.Background()); pingErr != nil { lastErr = pingErr; time.Sleep(2 * time.Second); continue }
		return pool, nil
	}
	return nil, fmt.Errorf("could not reach database after 5 attempts: %w", lastErr)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok { return val }
	return fallback
}
