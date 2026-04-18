package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/michael/language-arena/backend/internal/config"
	"github.com/michael/language-arena/backend/internal/handler"
	"github.com/michael/language-arena/backend/internal/middleware"
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
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			total_score BIGINT DEFAULT 0,
			games_played INT DEFAULT 0,
			best_reaction_ms INT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_total_score ON users(total_score DESC)`,
		`CREATE TABLE IF NOT EXISTS vocabularies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			word VARCHAR(100) NOT NULL,
			meaning VARCHAR(255) NOT NULL,
			language VARCHAR(5) NOT NULL,
			level VARCHAR(10) NOT NULL DEFAULT 'A1',
			difficulty INT DEFAULT 1 CHECK (difficulty BETWEEN 1 AND 3),
			category VARCHAR(50)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_vocabularies_language ON vocabularies(language)`,
		`CREATE INDEX IF NOT EXISTS idx_vocabularies_level ON vocabularies(language, level)`,
		`CREATE TABLE IF NOT EXISTS game_sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mode VARCHAR(10) NOT NULL CHECK (mode IN ('solo', 'duel', 'battle')),
			language VARCHAR(5) NOT NULL,
			winner_id UUID REFERENCES users(id),
			rounds INT DEFAULT 10,
			avg_reaction_ms INT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			finished_at TIMESTAMPTZ
		)`,
		// Add level column if table already exists (idempotent)
		`DO $$ BEGIN
			ALTER TABLE vocabularies ADD COLUMN IF NOT EXISTS level VARCHAR(10) NOT NULL DEFAULT 'A1';
		EXCEPTION WHEN duplicate_column THEN NULL;
		END $$`,
		`CREATE TABLE IF NOT EXISTS game_session_players (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id),
			score INT DEFAULT 0,
			avg_reaction_ms INT DEFAULT 0,
			best_reaction_ms INT DEFAULT 0,
			rank INT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_gsp_session ON game_session_players(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gsp_user ON game_session_players(user_id)`,
		// Drop redundant player columns from game_sessions (now in game_session_players)
		`DO $$ BEGIN
			ALTER TABLE game_sessions DROP COLUMN IF EXISTS player1_id;
			ALTER TABLE game_sessions DROP COLUMN IF EXISTS player2_id;
			ALTER TABLE game_sessions DROP COLUMN IF EXISTS player1_score;
			ALTER TABLE game_sessions DROP COLUMN IF EXISTS player2_score;
		END $$`,
		// Fix CHECK constraint to include 'battle' mode
		`DO $$ BEGIN
			ALTER TABLE game_sessions DROP CONSTRAINT IF EXISTS game_sessions_mode_check;
			ALTER TABLE game_sessions ADD CONSTRAINT game_sessions_mode_check CHECK (mode IN ('solo', 'duel', 'battle'));
		END $$`,
	}

	for i, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}
