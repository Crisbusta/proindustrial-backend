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

	// Storage
	StorageDriver   string // "local" (default) | "s3"
	StorageDir      string // local only: path to uploads dir
	S3Bucket        string
	S3Region        string
	S3Endpoint      string // R2: https://<account>.r2.cloudflarestorage.com
	S3AccessKey     string
	S3SecretKey     string
	S3PublicBase    string // CDN base URL for S3/R2
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

		StorageDriver:   getEnv("STORAGE_DRIVER", "local"),
		StorageDir:      getEnv("STORAGE_DIR", "./uploads"),
		S3Bucket:        getEnv("S3_BUCKET", ""),
		S3Region:        getEnv("S3_REGION", "auto"),
		S3Endpoint:      getEnv("S3_ENDPOINT", ""),
		S3AccessKey:     getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:     getEnv("S3_SECRET_KEY", ""),
		S3PublicBase:    getEnv("S3_PUBLIC_BASE", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
