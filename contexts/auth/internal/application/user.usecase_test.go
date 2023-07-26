//go:build integration

package application_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-arrower/arrower/postgres"
	"github.com/go-arrower/arrower/tests"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

var (
	ctx       = context.Background()
	pgHandler *postgres.Handler
)

func TestMain(m *testing.M) {
	handler, cleanup := tests.GetDBConnectionForIntegrationTesting(ctx)
	pgHandler = handler

	//
	// Run tests
	code := m.Run()

	//
	// Cleanup
	_ = handler.Shutdown(ctx)
	_ = cleanup()

	os.Exit(code)
}

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("login already in use", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		_, _ = queries.CreateUser(ctx, models.CreateUserParams{
			UserLogin:        userLogin,
			UserPasswordHash: "xxxxxx",
		})

		cmd := application.RegisterUser(queries)

		_, err := cmd(ctx, application.RegisterUserRequest{RegisterEmail: userLogin})
		assert.Error(t, err)
		assert.ErrorIs(t, err, application.ErrUserAlreadyExists)
	})

	t.Run("password weak", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)
		cmd := application.RegisterUser(queries)

		tests := []struct {
			testName string
			password string
		}{
			{
				"too short",
				"123456",
			},
			{
				"missing lower case letter",
				"1234567890",
			},
			{
				"missing upper case letter",
				"123456abc",
			},
			{
				"missing number",
				"abcdefghi",
			},
			{
				"missing special character",
				"123456abCD",
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.testName, func(t *testing.T) {
				t.Parallel()

				_, err := cmd(ctx, application.RegisterUserRequest{RegisterEmail: userLogin, Password: tt.password})
				assert.Error(t, err)
				assert.ErrorIs(t, err, application.ErrPasswordTooWeak)
			})
		}
	})

	t.Run("register new user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.RegisterUser(queries)

		_, err := cmd(ctx, application.RegisterUserRequest{RegisterEmail: userLogin, Password: strongPassword})
		assert.NoError(t, err)

		users, err := queries.AllUsers(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
	})
}
