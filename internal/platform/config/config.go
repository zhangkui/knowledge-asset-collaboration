package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	UploadLimit int64
	Environment string
}

func Load() Config {
	return Config{HTTPAddr: env("HTTP_ADDR", ":8080"), DatabaseURL: env("DATABASE_URL", "postgres://knowledge:knowledge@localhost:5432/knowledge?sslmode=disable"), RedisURL: env("REDIS_URL", "redis://localhost:6379/0"), JWTSecret: env("JWT_SECRET", "development-secret-change-me"), AccessTTL: duration("ACCESS_TOKEN_TTL", 15*time.Minute), RefreshTTL: duration("REFRESH_TOKEN_TTL", 7*24*time.Hour), UploadLimit: int64(integer("UPLOAD_LIMIT_MB", 100)) * 1024 * 1024, Environment: env("APP_ENV", "development")}
}
func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTPAddr) == "" {
		return errors.New("HTTP_ADDR is required")
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return errors.New("DATABASE_URL is required")
	}
	if strings.TrimSpace(c.RedisURL) == "" {
		return errors.New("REDIS_URL is required")
	}
	if len(c.JWTSecret) < 16 {
		return errors.New("JWT_SECRET must contain at least 16 characters")
	}
	if c.UploadLimit <= 0 {
		return errors.New("UPLOAD_LIMIT_MB must be positive")
	}
	return nil
}
func env(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}
func integer(k string, f int) int {
	v, err := strconv.Atoi(env(k, strconv.Itoa(f)))
	if err != nil || v <= 0 {
		return f
	}
	return v
}
func duration(k string, f time.Duration) time.Duration {
	v := env(k, "")
	if v == "" {
		return f
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return f
	}
	return d
}
