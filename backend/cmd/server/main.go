package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("✅ Database connected")

	if err := runMigrations(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("✅ Migrations applied")

	userRepo := repository.NewUserRepository(db)
	vocabRepo := repository.NewVocabRepository(db)
	gameRepo := repository.NewGameRepository(db)

	authService := service.NewAuthService(userRepo, &cfg.JWT)
	vocabService := service.NewVocabService(vocabRepo)
	leaderboardService := service.NewLeaderboardService(userRepo, gameRepo)

	hub := ws.NewHub(vocabService, gameRepo, userRepo)
	go hub.Run()

	authHandler := handler.NewAuthHandler(authService)
	vocabHandler := handler.NewVocabHandler(vocabService)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardService)
	gameHandler := handler.NewGameHandler(hub, userRepo, authService)

	r := setupRouter(cfg, authService, authHandler, vocabHandler, leaderboardHandler, gameHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

func setupRouter(
	cfg *config.Config,
	authService *service.AuthService,
	authHandler *handler.AuthHandler,
	vocabHandler *handler.VocabHandler,
	leaderboardHandler *handler.LeaderboardHandler,
	gameHandler *handler.GameHandler,
) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	r.Use(rateLimiter.Middleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

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

func runMigrations(db *sql.DB) error {
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
		log.Printf("  applied: %s", name)
	}

	return nil
}
