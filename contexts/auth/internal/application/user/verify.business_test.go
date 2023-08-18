package user_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
)

var (
	ctx        = context.Background()
	userIDZero = user.ID("00000000-0000-0000-0000-000000000000")
)

func TestVerificationService_NewVerificationToken(t *testing.T) {
	t.Parallel()

	t.Run("new token", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewMemoryRepository()

		usr := user.User{ID: user.NewID()}
		_ = repo.Save(ctx, usr)

		verifier := user.NewVerificationService(repo)
		token, err := verifier.NewVerificationToken(ctx, usr)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// verify against the db
		tok, err := repo.VerificationTokenByToken(ctx, token.Token())
		assert.NoError(t, err)
		assert.Equal(t, token.Token(), tok.Token())
	})
}

func TestVerificationService_Verify(t *testing.T) {
	t.Parallel()

	t.Run("verify token", func(t *testing.T) {
		t.Parallel()

		// setup
		repo := repository.NewMemoryRepository()
		_ = repo.Save(ctx, user.User{
			ID: userIDZero,
		})

		usr, _ := repo.FindByID(ctx, userIDZero)
		assert.False(t, usr.IsVerified())

		verifier := user.NewVerificationService(repo)
		token, _ := verifier.NewVerificationToken(ctx, usr)

		// action
		err := verifier.Verify(ctx, &usr, token.Token())
		assert.NoError(t, err)
		assert.True(t, usr.IsVerified())

		// verify against the db
		u, _ := repo.FindByID(ctx, usr.ID)
		assert.True(t, u.IsVerified()) // todo remove as it is not an integration test any more?
	})

	t.Run("verify an unknown token", func(t *testing.T) {
		t.Parallel()

		// setup
		repo := repository.NewMemoryRepository()

		_ = repo.Save(ctx, user.User{
			ID: userIDZero,
		})

		usr, _ := repo.FindByID(ctx, userIDZero)
		assert.False(t, usr.IsVerified())

		verifier := user.NewVerificationService(repo)

		// action
		err := verifier.Verify(ctx, &usr, uuid.New())
		assert.ErrorIs(t, err, user.ErrVerificationFailed)

		// verify against the db
		u, _ := repo.FindByID(ctx, usr.ID)
		assert.False(t, u.IsVerified())
	})

	t.Run("verify expired token", func(t *testing.T) {
		t.Parallel()

		// setup
		repo := repository.NewMemoryRepository()

		_ = repo.Save(ctx, user.User{
			ID: userIDZero, // todo extract a NotVerifiedUser and reuse in all testcases
		})

		usr, _ := repo.FindByID(ctx, userIDZero)
		assert.False(t, usr.IsVerified())

		verifier := user.NewVerificationService(repo, user.WithValidTime(time.Nanosecond))
		token, _ := verifier.NewVerificationToken(ctx, usr)

		// action
		err := verifier.Verify(ctx, &usr, token.Token())
		assert.ErrorIs(t, err, user.ErrVerificationFailed)

		// verify against the db
		u, _ := repo.FindByID(ctx, usr.ID)
		assert.False(t, u.IsVerified())
	})
}
