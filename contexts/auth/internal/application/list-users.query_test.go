package application_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
)

func TestListUsersQueryHandler_H(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	u0, _ := user.NewUser(gofakeit.Email(), "abcdefA0$")
	u1, _ := user.NewUser(gofakeit.Email(), "abcdefA0$")
	users := []user.User{u0, u1}

	t.Run("no users", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewMemoryRepository()
		handler := application.NewListUsersQueryHandler(repo)

		res, err := handler.H(context.Background(), application.ListUsersQuery{})
		assert.NoError(t, err)
		assert.Empty(t, res.Users)
		assert.Equal(t, uint(0), res.Filtered)
		assert.Equal(t, uint(0), res.Total)
	})

	t.Run("success case", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewMemoryRepository()
		repo.SaveAll(ctx, users)
		handler := application.NewListUsersQueryHandler(repo)

		res, err := handler.H(context.Background(), application.ListUsersQuery{})
		assert.NoError(t, err)
		assert.NotEmpty(t, res.Users)
		assert.Equal(t, uint(2), res.Filtered)
		assert.Equal(t, uint(2), res.Total)
	})
}
