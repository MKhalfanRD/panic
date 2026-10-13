package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPPort             string
	AppURL               string
	AppEnv               string
	DBDSN                string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	JWTSecret            string
	JWTExpiresInHours    int
	MidtransServerKey    string
	MidtransClientKey    string
	MidtransIsProduction bool
	SeedPassword         string
}

func Load() Config {
	env := getenv("APP_ENV", "development")
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" && env != "production" {
		secret = "dev-only-change-me"
	}
	hours, _ := strconv.Atoi(getenv("JWT_EXPIRES_IN_HOURS", "24"))
	if hours <= 0 {
		hours = 24
	}
	return Config{
		HTTPPort:             getenv("HTTP_PORT", "8081"),
		AppURL:               strings.TrimRight(strings.TrimSpace(getenv("APP_URL", "")), "/"),
		AppEnv:               env,
		DBDSN:                strings.TrimSpace(os.Getenv("DB_DSN")),
		DBHost:               getenv("DB_HOST", "127.0.0.1"),
		DBPort:               getenv("DB_PORT", "3306"),
		DBUser:               getenv("DB_USER", "root"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBName:               getenv("DB_NAME", "panic"),
		JWTSecret:            secret,
		JWTExpiresInHours:    hours,
		MidtransServerKey:    strings.TrimSpace(os.Getenv("MIDTRANS_SERVER_KEY")),
		MidtransClientKey:    strings.TrimSpace(os.Getenv("MIDTRANS_CLIENT_KEY")),
		MidtransIsProduction: getenv("MIDTRANS_IS_PRODUCTION", "false") == "true",
		SeedPassword:         os.Getenv("SEED_PASSWORD"),
	}
}

func (c Config) SecureCookie() bool {
	return strings.HasPrefix(strings.ToLower(c.AppURL), "https://")
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
