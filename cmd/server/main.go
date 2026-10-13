package main

import (
	"log"

	"github.com/joho/godotenv"

	"panic/internal/config"
	"panic/internal/db"
	apphttp "panic/internal/http"
	"panic/internal/seed"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.AppEnv == "production" && len(cfg.JWTSecret) < 16 {
		log.Fatal("JWT_SECRET minimal 16 karakter di production")
	}

	conn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrasi: %v", err)
	}
	if err := seed.Ensure(conn, cfg.SeedPassword); err != nil {
		log.Fatalf("seed: %v", err)
	}
	if err := apphttp.Run(cfg, conn); err != nil {
		log.Fatal(err)
	}
}
