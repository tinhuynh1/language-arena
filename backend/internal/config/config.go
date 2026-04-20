package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	LogLevel     string
	LogFormat    string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedWSOrigins []string
}

func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
			LogLevel:     getEnv("LOG_LEVEL", "info"),
			LogFormat:    getEnv("LOG_FORMAT", "json"),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DB_URL", "postgres://lingouser:lingopass@localhost:5432/lingodb?sslmode=disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "dev-secret-change-in-prod"),
			Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getSliceEnv("CORS_ORIGINS", []string{"http://localhost:3000", "http://localhost:3001"}),
			AllowedWSOrigins: getSliceEnv("ALLOWED_WS_ORIGINS", []string{"http://localhost:3000", "http://localhost:3001"}),
		},
	}

	cfg.validate()
	return cfg
}

func (c *Config) validate() {
	if os.Getenv("GIN_MODE") == "release" && c.JWT.Secret == "dev-secret-change-in-prod" {
		slog.Error("FATAL: JWT_SECRET must be changed in production (GIN_MODE=release)")
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getSliceEnv(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return fallback
}
