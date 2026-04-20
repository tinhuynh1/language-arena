package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/config"
	"github.com/michael/language-arena/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// UserReader abstracts read operations on users.
type UserReader interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// UserWriter abstracts write operations on users.
type UserWriter interface {
	Create(ctx context.Context, user *model.User) error
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user with this email already exists")
	ErrUsernameExists     = errors.New("username already taken")
)

type AuthService struct {
	userReader UserReader
	userWriter UserWriter
	cfg        *config.JWTConfig
	log        *slog.Logger
}

func NewAuthService(reader UserReader, writer UserWriter, cfg *config.JWTConfig) *AuthService {
	return &AuthService{
		userReader: reader,
		userWriter: writer,
		cfg:        cfg,
		log:        slog.Default().With("component", "SVC.Auth"),
	}
}

func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	start := time.Now()

	existing, _ := s.userReader.FindByEmail(ctx, req.Email)
	if existing != nil {
		s.log.Info("register rejected: email exists", "op", "Register", "email", req.Email)
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("bcrypt hash failed", "op", "Register", "err", err)
		return nil, err
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	if err := s.userWriter.Create(ctx, user); err != nil {
		s.log.Error("user creation failed", "op", "Register", "email", req.Email, "username", req.Username, "err", err)
		return nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		s.log.Error("token generation failed", "op", "Register", "user_id", user.ID, "err", err)
		return nil, err
	}

	s.log.Info("user registered", "op", "Register", "user_id", user.ID, "email", req.Email, "duration_ms", time.Since(start).Milliseconds())
	return &model.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	start := time.Now()

	user, err := s.userReader.FindByEmail(ctx, req.Email)
	if err != nil {
		s.log.Debug("login failed: user not found", "op", "Login", "email", req.Email)
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.log.Info("login failed: wrong password", "op", "Login", "email", req.Email, "user_id", user.ID)
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		s.log.Error("token generation failed", "op", "Login", "user_id", user.ID, "err", err)
		return nil, err
	}

	s.log.Info("user logged in", "op", "Login", "user_id", user.ID, "duration_ms", time.Since(start).Milliseconds())
	return &model.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(s.cfg.Expiration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		s.log.Debug("token validation failed", "op", "ValidateToken", "err", err)
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid user_id in token")
	}

	return uuid.Parse(userIDStr)
}
