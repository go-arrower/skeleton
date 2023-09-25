package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

func TestAuthenticationService_Authenticate(t *testing.T) {
	t.Parallel()

	t.Run("login setting disabled", func(t *testing.T) {
		t.Parallel()

		authenticator := user.NewAuthenticationService(settingsService(false))

		auth := authenticator.Authenticate(ctx, newVerifiedUser(), rawPassword)
		assert.False(t, auth)
	})

	t.Run("user not verified", func(t *testing.T) {
		t.Parallel()

		authenticator := user.NewAuthenticationService(settingsService(true))

		auth := authenticator.Authenticate(ctx, newUser(), "")
		assert.False(t, auth)
	})

	t.Run("user blocked", func(t *testing.T) {
		t.Parallel()

		usr := newVerifiedUser()
		usr.Blocked = user.BoolFlag{}.SetTrue()
		authenticator := user.NewAuthenticationService(settingsService(true))

		auth := authenticator.Authenticate(ctx, usr, "")
		assert.False(t, auth)
	})

	t.Run("password doesn't match", func(t *testing.T) {
		t.Parallel()

		authenticator := user.NewAuthenticationService(settingsService(true))

		auth := authenticator.Authenticate(ctx, newVerifiedUser(), "wrong-password")
		assert.False(t, auth)
	})

	t.Run("password matches", func(t *testing.T) {
		t.Parallel()

		authenticator := user.NewAuthenticationService(settingsService(true))

		auth := authenticator.Authenticate(ctx, newVerifiedUser(), rawPassword)
		assert.True(t, auth)
	})
}
