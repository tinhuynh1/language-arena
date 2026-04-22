package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/config"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/michael/language-arena/backend/internal/service"
	"github.com/michael/language-arena/backend/pkg/response"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ── Mocks ──────────────────────────────────────────────

type mockAuthUserReader struct {
	findByEmailFn func(ctx context.Context, email string) (*model.User, error)
	findByIDFn    func(ctx context.Context, id uuid.UUID) (*model.User, error)
}

func (m *mockAuthUserReader) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.findByEmailFn(ctx, email)
}

func (m *mockAuthUserReader) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, errors.New("not implemented")
}

type mockAuthUserWriter struct {
	createFn func(ctx context.Context, user *model.User) error
}

func (m *mockAuthUserWriter) Create(ctx context.Context, user *model.User) error {
	return m.createFn(ctx, user)
}

func testAuthService() *service.AuthService {
	reader := &mockAuthUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}
	writer := &mockAuthUserWriter{
		createFn: func(_ context.Context, u *model.User) error {
			u.ID = uuid.New()
			return nil
		},
	}
	return service.NewAuthService(reader, writer, &config.JWTConfig{
		Secret:     "test-secret",
		Expiration: time.Hour,
	})
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) response.APIResponse {
	var resp response.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return resp
}

// ── Register Handler Tests ──────────────────────────────

func TestRegisterHandler_Success(t *testing.T) {
	reader := &mockAuthUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}
	writer := &mockAuthUserWriter{
		createFn: func(_ context.Context, u *model.User) error {
			u.ID = uuid.New()
			return nil
		},
	}
	svc := service.NewAuthService(reader, writer, &config.JWTConfig{Secret: "s", Expiration: time.Hour})
	h := NewAuthHandler(svc)

	body := `{"username":"newuser","email":"new@test.com","password":"password123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp.Success)
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	h := NewAuthHandler(testAuthService())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{invalid`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	resp := parseResponse(t, w)
	assert.False(t, resp.Success)
}

func TestRegisterHandler_EmailExists(t *testing.T) {
	reader := &mockAuthUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{Email: "exists@test.com"}, nil
		},
	}
	writer := &mockAuthUserWriter{createFn: func(_ context.Context, _ *model.User) error { return nil }}
	svc := service.NewAuthService(reader, writer, &config.JWTConfig{Secret: "s", Expiration: time.Hour})
	h := NewAuthHandler(svc)

	body := `{"username":"u","email":"exists@test.com","password":"pass123"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── Login Handler Tests ─────────────────────────────────

func TestLoginHandler_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	reader := &mockAuthUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{ID: uuid.New(), Username: "u", Email: "u@t.com", PasswordHash: string(hash)}, nil
		},
	}
	svc := service.NewAuthService(reader, nil, &config.JWTConfig{Secret: "s", Expiration: time.Hour})
	h := NewAuthHandler(svc)

	body := `{"email":"u@t.com","password":"correct"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := parseResponse(t, w)
	assert.True(t, resp.Success)
}

func TestLoginHandler_InvalidBody(t *testing.T) {
	h := NewAuthHandler(testAuthService())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{bad`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.MinCost)
	reader := &mockAuthUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{ID: uuid.New(), PasswordHash: string(hash)}, nil
		},
	}
	svc := service.NewAuthService(reader, nil, &config.JWTConfig{Secret: "s", Expiration: time.Hour})
	h := NewAuthHandler(svc)

	body := `{"email":"u@t.com","password":"wrong"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
