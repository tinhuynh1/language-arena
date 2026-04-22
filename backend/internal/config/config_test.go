package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	// Unset ENV vars to test defaults
	os.Clearenv()

	cfg := Load()

	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, "info", cfg.Server.LogLevel)
	assert.Equal(t, "json", cfg.Server.LogFormat)

	assert.Equal(t, "postgres://lingouser:lingopass@localhost:5432/lingodb?sslmode=disable", cfg.Database.URL)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, cfg.Database.ConnMaxLifetime)

	assert.Equal(t, "redis://localhost:6379", cfg.Redis.URL)

	assert.Equal(t, "dev-secret-change-in-prod", cfg.JWT.Secret)
	assert.Equal(t, 24*time.Hour, cfg.JWT.Expiration)

	assert.Equal(t, []string{"http://localhost:3000", "http://localhost:3001"}, cfg.CORS.AllowedOrigins)
	assert.Equal(t, []string{"http://localhost:3000", "http://localhost:3001"}, cfg.CORS.AllowedWSOrigins)
}

func TestLoad_CustomEnv(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("READ_TIMEOUT", "15s")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("CORS_ORIGINS", "http://example.com, http://test.com ") // testing trim space

	cfg := Load()

	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, 15*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 50, cfg.Database.MaxOpenConns)
	assert.Equal(t, []string{"http://example.com", "http://test.com"}, cfg.CORS.AllowedOrigins)

	os.Clearenv()
}

func TestGetIntEnv_Invalid(t *testing.T) {
	os.Setenv("DB_MAX_OPEN_CONNS", "invalid")
	cfg := Load()
	assert.Equal(t, 25, cfg.Database.MaxOpenConns) // Fallback due to error
	os.Clearenv()
}

func TestGetDurationEnv_Invalid(t *testing.T) {
	os.Setenv("READ_TIMEOUT", "invalid")
	cfg := Load()
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout) // Fallback due to error
	os.Clearenv()
}

func TestGetSliceEnv_EmptyItems(t *testing.T) {
	os.Setenv("CORS_ORIGINS", " ")
	cfg := Load()
	// Fallback due to no valid items
	assert.Equal(t, []string{"http://localhost:3000", "http://localhost:3001"}, cfg.CORS.AllowedOrigins)
	os.Clearenv()
}
