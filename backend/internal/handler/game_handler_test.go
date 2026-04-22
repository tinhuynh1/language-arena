package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/config"
	"github.com/michael/language-arena/backend/internal/repository"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/internal/ws"
	"github.com/stretchr/testify/assert"
)

func TestGameHandler_GetOnlineCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := ws.NewHub(nil, nil, nil, nil)
	handler := NewGameHandler(hub, nil, nil, nil, nil)

	router := gin.New()
	router.GET("/online", handler.GetOnlineCount)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/online", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var res map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &res)
	
	data := res["data"].(map[string]interface{})
	assert.Equal(t, float64(0), data["online"])
}

func TestGameHandler_GetGameHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	gameRepo := repository.NewGameRepository(db)
	handler := NewGameHandler(nil, nil, gameRepo, nil, nil)

	router := gin.New()
	router.GET("/history", func(c *gin.Context) {
		userID := uuid.New()
		c.Set("user_id", userID) // Simulate auth middleware
		handler.GetGameHistory(c)
	})

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT DISTINCT gs.id").
			WillReturnRows(sqlmock.NewRows([]string{"id", "mode", "language", "winner_id", "rounds", "avg_reaction_ms", "created_at", "finished_at"}).
				AddRow(uuid.New(), "duel", "en", nil, 5, 1200, time.Now(), time.Now()))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/history", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "duel")
	})

	t.Run("repo error", func(t *testing.T) {
		mock.ExpectQuery("SELECT DISTINCT gs.id").WillReturnError(sqlmock.ErrCancelled)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/history", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGameHandler_GetGameHistory_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewGameHandler(nil, nil, nil, nil, nil)

	router := gin.New()
	router.GET("/history", handler.GetGameHistory) // Without setting "user_id"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/history", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGameHandler_HandleWebSocket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hub := ws.NewHub(nil, nil, nil, nil)
	authSvc := service.NewAuthService(nil, nil, &config.JWTConfig{Secret: "test", Expiration: time.Hour})
	gh := NewGameHandler(hub, nil, nil, authSvc, nil)
	
	router := gin.New()
	router.GET("/ws", gh.HandleWebSocket)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ws", nil)
	// without websocket upgrade headers, it will fail and log the error, 
	// which is what we want for minimal logic coverage
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
