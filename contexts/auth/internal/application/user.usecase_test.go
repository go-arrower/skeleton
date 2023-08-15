//go:build integration

package application_test

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-arrower/arrower/alog"
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/postgres"
	"github.com/go-arrower/arrower/tests"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/testdata"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
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

		buf := bytes.Buffer{}
		logger := alog.NewTest(&buf)
		alog.Unwrap(logger).SetLevel(alog.LevelInfo)

		cmd := application.LoginUser(logger, queries, nil)

		_, err := cmd(ctx, application.LoginUserRequest{
			LoginEmail: testdata.ValidUserLogin,
			Password:   "wrong-password",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, application.ErrLoginFailed)

		// assert failed attempt is logged, e.g. for monitoring or fail2ban etc.
		assert.Contains(t, buf.String(), "login failed")
	})

	t.Run("login fails - user not verified", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.LoginUser(alog.NewTest(nil), queries, nil)

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

		cmd := application.LoginUser(alog.NewTest(nil), queries, nil)

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
		queue := jobs.NewInMemoryJobs()

		cmd := application.LoginUser(alog.NewTest(nil), queries, queue)

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
		assert.NotEmpty(t, sessions[1].UserID)

		queue.Assert(t).Queued(application.SendConfirmationNewDeviceLoggedIn{}, 0)
	})

	t.Run("unknown device - send email about login to user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)
		queue := jobs.NewInMemoryJobs()

		cmd := application.LoginUser(alog.NewTest(nil), queries, queue)

		_, err := cmd(ctx, application.LoginUserRequest{
			LoginEmail:  testdata.ValidUserLogin,
			Password:    testdata.StrongPassword,
			UserAgent:   testdata.UserAgent,
			IP:          testdata.IP,
			SessionKey:  "new-session-key",
			IsNewDevice: true,
		})

		// assert return values
		assert.NoError(t, err)
		queue.Assert(t).Queued(application.SendConfirmationNewDeviceLoggedIn{}, 1)
		job := queue.GetFirstOf(application.SendConfirmationNewDeviceLoggedIn{}).(application.SendConfirmationNewDeviceLoggedIn)
		assert.NotEmpty(t, job.UserID)
		assert.NotEmpty(t, job.OccurredAt)
		assert.Equal(t, testdata.IP, job.IP)
		assert.Equal(t, user.NewDevice(testdata.UserAgent), job.Device)
	})
}

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	t.Run("login already in use", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		buf := bytes.Buffer{}
		logger := alog.NewTest(&buf)
		alog.Unwrap(logger).SetLevel(alog.LevelInfo)

		cmd := application.RegisterUser(logger, queries, nil)

		_, err := cmd(ctx, application.RegisterUserRequest{RegisterEmail: testdata.ValidUserLogin})
		assert.Error(t, err)
		assert.ErrorIs(t, err, application.ErrUserAlreadyExists)

		// assert failed attempt is logged, e.g. for monitoring or fail2ban etc.
		assert.Contains(t, buf.String(), "register new user failed")
	})

	t.Run("password weak", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)
		cmd := application.RegisterUser(alog.NewTest(nil), queries, nil)

		tests := []struct {
			testName string
			password string
		}{
			{
				"weak pw",
				"123",
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.testName, func(t *testing.T) {
				t.Parallel()

				_, err := cmd(ctx, application.RegisterUserRequest{RegisterEmail: testdata.NewUserLogin, Password: tt.password})
				assert.Error(t, err)
				assert.ErrorIs(t, err, user.ErrPasswordTooWeak)
			})
		}
	})

	t.Run("register new user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)
		queue := jobs.NewInMemoryJobs()

		cmd := application.RegisterUser(alog.NewTest(nil), queries, queue)

		usr, err := cmd(ctx, application.RegisterUserRequest{
			RegisterEmail: testdata.NewUserLogin,
			Password:      testdata.StrongPassword,
			UserAgent:     testdata.UserAgent,
			IP:            testdata.IP,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, usr.User.ID)

		dbUser, err := queries.FindUserByLogin(ctx, testdata.NewUserLogin)
		assert.NoError(t, err)
		assert.Empty(t, dbUser.VerifiedAt)
		assert.Empty(t, dbUser.BlockedAt)

		// assert session got updated with device info
		sessions, _ := queries.AllSessions(ctx)
		assert.Len(t, sessions, 1+1) // 1 session is already created via _common.yaml fixtures
		assert.Equal(t, testdata.UserAgent, sessions[1].UserAgent)
		assert.NotEmpty(t, sessions[1].UserID)

		queue.Assert(t).Queued(application.NewUserVerificationEmail{}, 1)
		job := queue.GetFirstOf(application.NewUserVerificationEmail{}).(application.NewUserVerificationEmail)
		assert.NotEmpty(t, job.UserID)
		assert.NotEmpty(t, job.OccurredAt)
		assert.Equal(t, testdata.IP, job.IP)
		assert.Equal(t, user.NewDevice(testdata.UserAgent), job.Device)
	})
}

func TestSendNewUserVerificationEmail(t *testing.T) {
	t.Parallel()

	t.Run("send new verification email", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.SendNewUserVerificationEmail(alog.NewDevelopment(), queries)
		err := cmd(ctx, application.NewUserVerificationEmail{
			UserID:     testdata.UserNotVerifiedUserID,
			OccurredAt: time.Now().UTC(),
			IP:         testdata.IP,
			Device:     user.NewDevice(testdata.UserAgent),
		})
		assert.NoError(t, err)

		// later: assert the email has been sent via the email interface
	})
}

func TestShowUser(t *testing.T) {
	t.Parallel()

	t.Run("invalid userID", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.ShowUser(queries)
		res, err := cmd(ctx, application.ShowUserRequest{})
		assert.Error(t, err)
		assert.Empty(t, res)
	})

	t.Run("show user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.ShowUser(queries)
		res, err := cmd(ctx, application.ShowUserRequest{
			UserID: testdata.UserIDZero,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, res)

		assert.Equal(t, testdata.UserIDZero, res.User.ID)
		assert.Len(t, res.User.Sessions, 1)
	})
}

func TestBlockUser(t *testing.T) {
	t.Parallel()

	t.Run("block user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.BlockUser(queries)
		_, err := cmd(ctx, application.BlockUserRequest{UserID: testdata.UserIDZero})
		assert.NoError(t, err)

		// verify
		usr, err := repository.GetUserByID(ctx, queries, testdata.UserIDZero)
		assert.NoError(t, err)
		assert.True(t, usr.IsBlocked())
	})
}

func TestUnblockUser(t *testing.T) {
	t.Parallel()

	t.Run("unblock user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		queries := models.New(pg)

		cmd := application.UnblockUser(queries)
		_, err := cmd(ctx, application.BlockUserRequest{UserID: testdata.UserBlockedUserID})
		assert.NoError(t, err)

		// verify
		usr, err := repository.GetUserByID(ctx, queries, testdata.UserIDZero)
		assert.NoError(t, err)
		assert.True(t, !usr.IsBlocked())
	})
}
