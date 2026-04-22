package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/config"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ── Mocks ──────────────────────────────────────────────

type mockUserReader struct {
	findByEmailFn func(ctx context.Context, email string) (*model.User, error)
	findByIDFn    func(ctx context.Context, id uuid.UUID) (*model.User, error)
}

func (m *mockUserReader) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.findByEmailFn(ctx, email)
}

func (m *mockUserReader) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, errors.New("not implemented")
}

type mockUserWriter struct {
	createFn func(ctx context.Context, user *model.User) error
}

func (m *mockUserWriter) Create(ctx context.Context, user *model.User) error {
	return m.createFn(ctx, user)
}

func testJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		Secret:     "test-secret-key-for-unit-tests",
		Expiration: 1 * time.Hour,
	}
}

// ── Register Tests ─────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	reader := &mockUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}
	writer := &mockUserWriter{
		createFn: func(_ context.Context, user *model.User) error {
			user.ID = uuid.New()
			return nil
		},
	}

	svc := NewAuthService(reader, writer, testJWTConfig())
	resp, err := svc.Register(context.Background(), model.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "testuser", resp.User.Username)
	assert.Equal(t, "test@example.com", resp.User.Email)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	reader := &mockUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{Email: "existing@example.com"}, nil
		},
	}
	writer := &mockUserWriter{
		createFn: func(_ context.Context, _ *model.User) error { return nil },
	}

	svc := NewAuthService(reader, writer, testJWTConfig())
	resp, err := svc.Register(context.Background(), model.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
	})

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrUserExists)
}

func TestRegister_CreateFails(t *testing.T) {
	reader := &mockUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}
	dbErr := errors.New("unique constraint violation: username")
	writer := &mockUserWriter{
		createFn: func(_ context.Context, _ *model.User) error {
			return dbErr
		},
	}

	svc := NewAuthService(reader, writer, testJWTConfig())
	resp, err := svc.Register(context.Background(), model.RegisterRequest{
		Username: "taken",
		Email:    "new@example.com",
		Password: "password123",
	})

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, dbErr)
}

// ── Login Tests ────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	userID := uuid.New()

	reader := &mockUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{
				ID:           userID,
				Username:     "loginuser",
				Email:        "login@example.com",
				PasswordHash: string(hash),
			}, nil
		},
	}

	svc := NewAuthService(reader, nil, testJWTConfig())
	resp, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "login@example.com",
		Password: "correctpassword",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, userID, resp.User.ID)
	assert.Equal(t, "loginuser", resp.User.Username)
}

func TestLogin_UserNotFound(t *testing.T) {
	reader := &mockUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, errors.New("not found")
		},
	}

	svc := NewAuthService(reader, nil, testJWTConfig())
	resp, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "noone@example.com",
		Password: "password",
	})

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("realpassword"), bcrypt.MinCost)

	reader := &mockUserReader{
		findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{
				ID:           uuid.New(),
				PasswordHash: string(hash),
			}, nil
		},
	}

	svc := NewAuthService(reader, nil, testJWTConfig())
	resp, err := svc.Login(context.Background(), model.LoginRequest{
		Email:    "user@example.com",
		Password: "wrongpassword",
	})

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// ── Token Tests ────────────────────────────────────────

func TestValidateToken_RoundTrip(t *testing.T) {
	svc := NewAuthService(nil, nil, testJWTConfig())
	userID := uuid.New()

	token, err := svc.generateToken(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedID, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, parsedID)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:     "test-secret",
		Expiration: -1 * time.Hour, // already expired
	}
	svc := NewAuthService(nil, nil, cfg)
	userID := uuid.New()

	token, err := svc.generateToken(userID)
	require.NoError(t, err)

	_, err = svc.ValidateToken(token)
	assert.Error(t, err)
}

func TestValidateToken_TamperedToken(t *testing.T) {
	svc := NewAuthService(nil, nil, testJWTConfig())
	userID := uuid.New()

	token, err := svc.generateToken(userID)
	require.NoError(t, err)

	// Tamper with the token by modifying a character
	tampered := token[:len(token)-4] + "XXXX"

	_, err = svc.ValidateToken(tampered)
	assert.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewAuthService(nil, nil, &config.JWTConfig{Secret: "secret-1", Expiration: time.Hour})
	svc2 := NewAuthService(nil, nil, &config.JWTConfig{Secret: "secret-2", Expiration: time.Hour})

	token, err := svc1.generateToken(uuid.New())
	require.NoError(t, err)

	_, err = svc2.ValidateToken(token)
	assert.Error(t, err)
}

func TestValidateToken_Malformed(t *testing.T) {
	svc := NewAuthService(nil, nil, testJWTConfig())

	_, err := svc.ValidateToken("not.a.valid.jwt.token")
	assert.Error(t, err)

	_, err = svc.ValidateToken("")
	assert.Error(t, err)
}
