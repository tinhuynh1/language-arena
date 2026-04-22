package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/michael/language-arena/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &model.User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hash123",
		}

		mock.ExpectQuery("INSERT INTO users").
			WithArgs(sqlmock.AnyArg(), user.Username, user.Email, user.PasswordHash).
			WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now()))

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
	})

	t.Run("error", func(t *testing.T) {
		user := &model.User{Email: "test@example.com"}
		mock.ExpectQuery("INSERT INTO users").WillReturnError(sql.ErrConnDone)

		err := repo.Create(ctx, user)
		assert.ErrorIs(t, err, sql.ErrConnDone)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		email := "test@example.com"
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "avg_reaction_ms", "total_correct", "games_played", "best_reaction_ms", "created_at"}).
			AddRow(uuid.New(), "testuser", email, "hash", 1200, 50, 10, 800, time.Now())

		mock.ExpectQuery("SELECT id, username, email").WithArgs(email).WillReturnRows(rows)

		user, err := repo.FindByEmail(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		if user != nil {
			assert.Equal(t, email, user.Email)
		}
	})

	t.Run("not found", func(t *testing.T) {
		email := "notfound@example.com"
		mock.ExpectQuery("SELECT id, username, email").WithArgs(email).WillReturnError(sql.ErrNoRows)

		user, err := repo.FindByEmail(ctx, email)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, user)
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "email", "password_hash", "avg_reaction_ms", "total_correct", "games_played", "best_reaction_ms", "created_at"}).
			AddRow(id, "testuser", "test@test.com", "hash", 1200, 50, 10, 800, time.Now())

		mock.ExpectQuery("SELECT id, username, email").WithArgs(id).WillReturnRows(rows)

		user, err := repo.FindByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		if user != nil {
			assert.Equal(t, id, user.ID)
		}
	})
}

func TestUserRepository_UpdateStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec("UPDATE users SET").
			WithArgs(id, 1000, 10, 800).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateStats(ctx, id, 1000, 10, 800)
		assert.NoError(t, err)
	})
}

func TestUserRepository_GetLeaderboard(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "avg_reaction_ms", "total_correct", "games_played", "best_reaction_ms", "total_count"}).
			AddRow(uuid.New(), "top1", 800, 100, 50, 500, 1)

		mock.ExpectQuery("SELECT id, username.*total_count").WithArgs(10, 0).WillReturnRows(rows)

		entries, total, err := repo.GetLeaderboard(ctx, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, entries, 1)
		if len(entries) > 0 {
			assert.Equal(t, "top1", entries[0].Username)
		}
	})
}
