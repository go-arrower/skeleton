//go:build integration

package application_test

import (
	"testing"
	"time"

	"github.com/go-arrower/arrower/tests"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/testdata"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

func TestVerificationService_NewVerificationToken(t *testing.T) {
	t.Parallel()

	t.Run("new token", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		usr := user.User{ID: user.NewID()}
		_ = repository.SaveUser(ctx, queries, usr)

		verify := application.NewVerificationService(queries)
		token, err := verify.NewVerificationToken(ctx, usr)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// verify against the db
		tok, err := queries.VerificationTokenByToken(ctx, token.Token())
		assert.NoError(t, err)
		assert.Equal(t, token.Token(), tok.Token)
	})
}

func TestVerificationService_Verify(t *testing.T) {
	t.Parallel()

	t.Run("verify token", func(t *testing.T) {
		t.Parallel()

		// setup
		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		usr, _ := repository.GetUserByID(ctx, queries, testdata.UserNotVerifiedUserID)
		assert.False(t, usr.IsVerified())

		verify := application.NewVerificationService(queries)
		token, _ := verify.NewVerificationToken(ctx, usr)

		// action
		err := verify.Verify(ctx, &usr, token.Token())
		assert.NoError(t, err)

		// verify against the db
		u, _ := repository.GetUserByID(ctx, queries, usr.ID)
		assert.True(t, u.IsVerified())
	})

	t.Run("verify an unknown token", func(t *testing.T) {
		t.Parallel()

		// setup
		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		usr, _ := repository.GetUserByID(ctx, queries, testdata.UserNotVerifiedUserID)
		assert.False(t, usr.IsVerified())

		verify := application.NewVerificationService(queries)

		// action
		err := verify.Verify(ctx, &usr, uuid.New())
		assert.ErrorIs(t, err, application.ErrVerificationFailed)

		// verify against the db
		u, _ := repository.GetUserByID(ctx, queries, usr.ID)
		assert.False(t, u.IsVerified())
	})

	t.Run("verify expired token", func(t *testing.T) {
		t.Parallel()

		// setup
		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		usr, _ := repository.GetUserByID(ctx, queries, testdata.UserNotVerifiedUserID)
		assert.False(t, usr.IsVerified())

		verify := application.NewVerificationService(queries, application.WithValidTime(time.Nanosecond))
		token, _ := verify.NewVerificationToken(ctx, usr)

		// action
		err := verify.Verify(ctx, &usr, token.Token())
		assert.ErrorIs(t, err, application.ErrVerificationFailed)

		// verify against the db
		u, _ := repository.GetUserByID(ctx, queries, usr.ID)
		assert.False(t, u.IsVerified())
	})
}
