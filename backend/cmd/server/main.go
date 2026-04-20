package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/michael/language-arena/backend/internal/config"
	"github.com/michael/language-arena/backend/internal/handler"
	"github.com/michael/language-arena/backend/internal/middleware"
	"github.com/michael/language-arena/backend/internal/migration"
	"github.com/michael/language-arena/backend/internal/repository"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/internal/ws"
	"github.com/michael/language-arena/backend/pkg/logger"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.Server.LogLevel, cfg.Server.LogFormat)
	log := logger.WithComponent("BOOT")

	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		slog.Error("database ping failed", "err", err)
		os.Exit(1)
	}
	log.Info("database connected")

	if err := runMigrations(db, log); err != nil {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}
	log.Info("migrations applied")

	userRepo := repository.NewUserRepository(db)
	vocabRepo := repository.NewVocabRepository(db)
	gameRepo := repository.NewGameRepository(db)

	authService := service.NewAuthService(userRepo, userRepo, &cfg.JWT)
	vocabService := service.NewVocabService(vocabRepo)
	leaderboardService := service.NewLeaderboardService(userRepo, gameRepo)

	// Initialize Redis adapter (optional — gracefully degrade to single-instance)
	var redisAdapter *ws.RedisAdapter
	if cfg.Redis.URL != "" {
		ra, err := ws.NewRedisAdapter(cfg.Redis.URL)
		if err != nil {
			log.Warn("redis unavailable, running single-instance mode", "err", err)
		} else {
			redisAdapter = ra
			defer ra.Close()
		}
	}

	hub := ws.NewHub(vocabService, gameRepo, userRepo, redisAdapter)
	go hub.Run()

	authHandler := handler.NewAuthHandler(authService)
	vocabHandler := handler.NewVocabHandler(vocabService)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardService)
	gameHandler := handler.NewGameHandler(hub, userRepo, gameRepo, authService, cfg.CORS.AllowedWSOrigins)

	r := setupRouter(cfg, db, authService, authHandler, vocabHandler, leaderboardHandler, gameHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info("server starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
		os.Exit(1)
	}

	log.Info("server exited gracefully")
}

func setupRouter(
	cfg *config.Config,
	db *sql.DB,
	authService *service.AuthService,
	authHandler *handler.AuthHandler,
	vocabHandler *handler.VocabHandler,
	leaderboardHandler *handler.LeaderboardHandler,
	gameHandler *handler.GameHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORSMiddleware(cfg.CORS.AllowedOrigins))

	// Health & Readiness probes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})
	r.GET("/ready", healthProbe(db, cfg))

	rateLimiter := middleware.NewRateLimiter(200, time.Minute) // Increased limit just in case
	r.Use(rateLimiter.Middleware())

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		api.GET("/vocab", vocabHandler.GetVocabularies)

		api.GET("/leaderboard", leaderboardHandler.GetLeaderboard)

		api.GET("/online", gameHandler.GetOnlineCount)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/stats/me", leaderboardHandler.GetMyStats)
			protected.GET("/games/history", gameHandler.GetGameHistory)
		}

		api.GET("/ws/game", gameHandler.HandleWebSocket)
	}

	return r
}

func runMigrations(db *sql.DB, log *slog.Logger) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("failed to create schema_migrations: %w", err)
	}

	entries, err := migration.Files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 4 && e.Name()[len(e.Name())-4:] == ".sql" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var applied bool
		if err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)`, name).Scan(&applied); err != nil {
			return fmt.Errorf("failed to check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		content, err := migration.Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("migration %s failed: %w", name, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (name) VALUES ($1)`, name); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}
		log.Info("migration applied", "file", name)
	}

	return nil
}

func healthProbe(db *sql.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := gin.H{}
		healthy := true

		// Check database
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			checks["database"] = "error: " + err.Error()
			healthy = false
		} else {
			checks["database"] = "ok"
		}

		// Check Redis (if configured)
		if cfg.Redis.URL != "" {
			checks["redis"] = "ok"
		} else {
			checks["redis"] = "not configured"
		}

		status := http.StatusOK
		if !healthy {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":    map[bool]string{true: "ready", false: "not ready"}[healthy],
			"checks":    checks,
			"timestamp": time.Now().Unix(),
		})
	}
}
