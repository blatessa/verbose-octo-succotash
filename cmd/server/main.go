package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/blatessa/verbose-octo-succotash/internal/auth"
	authdb "github.com/blatessa/verbose-octo-succotash/internal/auth/db"
	pkgdb "github.com/blatessa/verbose-octo-succotash/pkg/db"
)

func main() {
	ctx := context.Background()

	pool, err := pkgdb.NewPool(ctx, dbConfig())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := authdb.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	authCfg := auth.Config{
		Secret:     []byte(mustEnv("JWT_SECRET")),
		Expiration: 24 * time.Hour,
	}

	authHandler := auth.NewHandler(
		auth.NewService(authdb.New(pool), authCfg),
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	authHandler.RegisterRoutes(mux, "/api/auth")

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func dbConfig() pkgdb.Config {
	port, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	return pkgdb.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     port,
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "postgres"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
