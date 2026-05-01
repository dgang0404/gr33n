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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	automationworker "gr33n-api/internal/automation"
	"gr33n-api/internal/filestorage"
	"gr33n-api/internal/pgxutil"
	"gr33n-api/internal/pushnotify"
)

// loadDotEnv reads optional .env then .env.local from the current working directory
// (later file overrides). Shell exports still win — offline / local setups set secrets once in .env.
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

func main() {
	loadDotEnv()

	dbURL := getEnv("DATABASE_URL", "postgres://davidg@/gr33n?host=/var/run/postgresql")
	port := getEnv("PORT", "8080")
	jwtSecret = []byte(getEnv("JWT_SECRET", ""))
	piAPIKey = getEnv("PI_API_KEY", "")
	corsOrigin = getEnv("CORS_ORIGIN", "http://localhost:5173")
	authMode = strings.ToLower(strings.TrimSpace(getEnv("AUTH_MODE", "production")))
	if authMode == "" {
		authMode = "production"
	}
	switch authMode {
	case "dev", "auth_test", "production":
	default:
		log.Fatalf("AUTH_MODE must be dev, auth_test, or production (got %q)", authMode)
	}
	adminUser := getEnv("ADMIN_USERNAME", "admin")
	simulationMode := strings.EqualFold(getEnv("AUTOMATION_SIMULATION_MODE", "true"), "true")
	hashFilePath := filepath.Join(os.Getenv("HOME"), ".gr33n", "admin.hash")
	adminHash := loadPasswordHash(hashFilePath)
	adminBindUserID, adminBindEmail := loadAdminBindIdentity()

	if authMode == "dev" && !devBypassAllowed {
		log.Fatal("AUTH_MODE=dev is not allowed in this binary. " +
			"Rebuild with `-tags dev` for local development, or set AUTH_MODE=production.")
	}
	if authMode == "auth_test" && !devBypassAllowed {
		log.Fatal("AUTH_MODE=auth_test is only for local binaries built with `-tags dev`. " +
			"Use AUTH_MODE=production in QA/production.")
	}

	if authMode != "dev" {
		if len(jwtSecret) == 0 {
			log.Fatal("JWT_SECRET must be set when AUTH_MODE != dev")
		}
		if piAPIKey == "" {
			log.Fatal("PI_API_KEY must be set when AUTH_MODE != dev")
		}
	}

	pool, err := connectDB(dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to gr33n database")
	log.Printf("AUTH_MODE=%s  (dev_bypass_compiled=%v)", authMode, devBypassAllowed)
	switch authMode {
	case "dev":
		log.Println("⚠️  DEV MODE — auth bypass ACTIVE. Do NOT deploy this binary to QA/production.")
	case "auth_test":
		log.Println("🧪 AUTH_TEST — JWT + API-key enforced (local dev binary). Use for login / regression against real auth.")
	default:
		log.Println("🔒 Auth enforced (JWT + API-key)")
	}
	log.Printf("CORS_ORIGIN=%s", corsOrigin)
	mux := http.NewServeMux()
	var workerOpts []automationworker.WorkerOption
	if cs := getEnv("AUTOMATION_COOLDOWN_SECONDS", ""); cs != "" {
		if n, err := strconv.Atoi(cs); err == nil && n > 0 {
			workerOpts = append(workerOpts, automationworker.WithCooldown(time.Duration(n)*time.Second))
			log.Printf("⏱  Automation cooldown set to %ds", n)
		}
	}
	pushDispatch := pushnotify.NewDispatcher(pool)
	workerOpts = append(workerOpts, automationworker.WithPushNotifier(pushDispatch))
	worker := automationworker.NewWorker(pool, simulationMode, workerOpts...)
	go worker.Start(context.Background())
	log.Printf("🧠 Automation worker started (simulation_mode=%v)", simulationMode)
	fileStore, fileCfg, err := filestorage.NewFromEnv(context.Background())
	if err != nil {
		log.Fatalf("file storage init: %v", err)
	}
	registerRoutes(mux, pool, worker, pushDispatch, adminUser, adminHash, hashFilePath, fileStore, fileCfg, adminBindUserID, adminBindEmail)
	log.Printf("FILE_STORAGE_BACKEND=%s", fileCfg.Backend)
	if fileCfg.Backend == "local" {
		log.Printf("FILE_STORAGE_DIR=%s", fileCfg.LocalRoot)
	} else {
		log.Printf("S3_BUCKET=%s S3_REGION=%s S3_ENDPOINT=%s S3_PREFIX=%s", fileCfg.S3Bucket, fileCfg.S3Region, fileCfg.S3Endpoint, fileCfg.S3Prefix)
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("🌱 gr33n API running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}

// loadAdminBindIdentity sets JWT claims for env-admin (ADMIN_USERNAME + admin.hash).
// Farm routes require authctx.UserID; without user_id the dashboard gets 401 on every /farms/... call.
// Defaults match db/seeds/master_seed.sql demo user.
func loadAdminBindIdentity() (uuid.UUID, string) {
	s := getEnv("ADMIN_BIND_USER_ID", "00000000-0000-0000-0000-000000000001")
	id, err := uuid.Parse(s)
	if err != nil {
		log.Fatalf("ADMIN_BIND_USER_ID: invalid UUID %q", s)
	}
	email := strings.TrimSpace(getEnv("ADMIN_BIND_EMAIL", "dev@gr33n.local"))
	log.Printf("🔑 Env-admin login binds JWT user_id=%s (farm RBAC)", id)
	return id, email
}

func loadPasswordHash(filePath string) []byte {
	if data, err := os.ReadFile(filePath); err == nil {
		hash := []byte(strings.TrimSpace(string(data)))
		if len(hash) > 0 {
			log.Printf("🔒 Loaded password hash from %s", filePath)
			return hash
		}
	}
	if h := getEnv("ADMIN_PASSWORD_HASH", ""); h != "" {
		log.Println("🔒 Loaded password hash from env")
		return []byte(h)
	}
	return nil
}

func connectDB(dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	pgxutil.RegisterVectorTypes(config)
	var pool *pgxpool.Pool
	var lastErr error
	for i := range 5 {
		log.Printf("⏳ Waiting for database... attempt %d/5", i+1)
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

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
