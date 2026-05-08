package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	JWTSecret      string
	GoogleClientID string
	DatabaseURL    string
}

func Load() *Config {
	// Best-effort load of .env (ignored if file absent in prod)
	_ = godotenv.Load("../../.env")

	return &Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      mustEnv("JWT_SECRET"),
		GoogleClientID: mustEnv("GOOGLE_CLIENT_ID"),
		DatabaseURL:    buildDSN(),
	}
}

func buildDSN() string {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "mtracker")
	pass := getEnv("POSTGRES_PASSWORD", "mtracker_secret")
	db := getEnv("POSTGRES_DB", "mtracker")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, pass, db)
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
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}
