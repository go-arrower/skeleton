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
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/testdata"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
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

func TestLoginUser(t *testing.T) {
	t.Parallel()

	t.Run("password does not match", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.LoginUser(queries)

		_, err := cmd(ctx, application.LoginUserRequest{
			LoginEmail: testdata.ValidUserLogin,
			Password:   "wrong-password",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, application.ErrLoginFailed)
	})

	t.Run("login fails - user not verified", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.LoginUser(queries)

		_, err := cmd(ctx, application.LoginUserRequest{
			LoginEmail: testdata.NotVerifiedUserLogin,
			Password:   testdata.StrongPassword,
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, application.ErrLoginFailed)
	})

	t.Run("login fails - user blocked", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.LoginUser(queries)

		_, err := cmd(ctx, application.LoginUserRequest{
			LoginEmail: testdata.BlockedUserLogin,
			Password:   testdata.StrongPassword,
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, application.ErrLoginFailed)
	})

	t.Run("login succeeds", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.LoginUser(queries)

		res, err := cmd(ctx, application.LoginUserRequest{
			LoginEmail: testdata.ValidUserLogin,
			Password:   testdata.StrongPassword,
			UserAgent:  testdata.UserAgent,
			SessionKey: "new-session-key",
		})

		// assert return values
		assert.NoError(t, err)
		assert.Equal(t, user.Login(testdata.ValidUserLogin), res.User.Login)
		assert.NotEmpty(t, testdata.ValidUserLogin, res.User.ID)

		// assert session got updated with device info
		sessions, _ := queries.AllSessions(ctx)
		assert.Len(t, sessions, 1+1) // 1 session is already created via _common.yaml fixtures
		assert.Equal(t, testdata.UserAgent, sessions[1].UserAgent)
	})
}

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("login already in use", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		_, _ = queries.CreateUser(ctx, models.CreateUserParams{
			Login:        testdata.ValidUserLogin,
			PasswordHash: "xxxxxx",
		})

		cmd := application.RegisterUser(queries)

		_, err := cmd(ctx, application.RegisterUserRequest{RegisterEmail: testdata.ValidUserLogin})
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

				_, err := cmd(ctx, application.RegisterUserRequest{RegisterEmail: testdata.NewUserLogin, Password: tt.password})
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

		_, err := cmd(ctx, application.RegisterUserRequest{RegisterEmail: testdata.NewUserLogin, Password: testdata.StrongPassword})
		assert.NoError(t, err)

		user, err := queries.FindUserByLogin(ctx, testdata.NewUserLogin)
		assert.NoError(t, err)
		assert.Empty(t, user.VerifiedAt)
		assert.Empty(t, user.BlockedAt)
	})
}
