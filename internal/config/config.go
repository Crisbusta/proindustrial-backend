package config

import (
	"log"
	"os"
)

const defaultJWTSecret = "dev-secret-change-me"

type Config struct {
	AppEnv          string
	DatabaseURL     string
	JWTSecret       string
	InitialPassword string
	Port            string
	CORSOrigin      string
	ResendAPIKey    string
	ResendFrom      string
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPPass        string
	SMTPFrom        string
	AppBaseURL      string
}

func Load() Config {
	appEnv := getEnv("APP_ENV", "development")

	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)
	if jwtSecret == defaultJWTSecret {
		if appEnv == "production" {
			log.Fatal("FATAL: JWT_SECRET must be set to a secure value in production. Refusing to start.")
		}
		log.Println("WARNING: JWT_SECRET is using the default dev value. Set a secure secret in production.")
	}

	return Config{
		AppEnv:          appEnv,
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://puntofusion:devpassword@localhost:5432/puntofusion?sslmode=disable"),
		JWTSecret:       jwtSecret,
		InitialPassword: getEnv("INITIAL_PASSWORD", ""),
		Port:            getEnv("PORT", "8080"),
		CORSOrigin:      getEnv("CORS_ORIGIN", "http://localhost:3001"),
		ResendAPIKey:    getEnv("RESEND_API_KEY", ""),
		ResendFrom:      getEnv("RESEND_FROM", ""),
		SMTPHost:        getEnv("SMTP_HOST", ""),
		SMTPPort:        getEnv("SMTP_PORT", "587"),
		SMTPUser:        getEnv("SMTP_USER", ""),
		SMTPPass:        getEnv("SMTP_PASS", ""),
		SMTPFrom:        getEnv("SMTP_FROM", ""),
		AppBaseURL:      getEnv("APP_BASE_URL", "http://localhost:3001"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
