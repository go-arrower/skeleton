package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

func TestAuthenticationService_Authenticate(t *testing.T) {
	t.Parallel()

	t.Run("user not verified", func(t *testing.T) {
		t.Parallel()

		authenticator := user.NewAuthenticationService()

		auth := authenticator.Authenticate(ctx, newUser(), "")
		assert.False(t, auth)
	})

	t.Run("user blocked", func(t *testing.T) {
		t.Parallel()

		usr := newUser()
		usr.Verified = user.BoolFlag{}.SetTrue()
		usr.Blocked = user.BoolFlag{}.SetTrue()
		authenticator := user.NewAuthenticationService()

		auth := authenticator.Authenticate(ctx, usr, "")
		assert.False(t, auth)
	})

	t.Run("password doesn't match", func(t *testing.T) {
		t.Parallel()

		usr := newUser()
		usr.Verified = user.BoolFlag{}.SetTrue()
		usr.PasswordHash = strongPasswordHash
		authenticator := user.NewAuthenticationService()

		auth := authenticator.Authenticate(ctx, usr, "wrong-password")
		assert.False(t, auth)
	})

	t.Run("password matches", func(t *testing.T) {
		t.Parallel()

		usr := newUser()
		usr.Verified = user.BoolFlag{}.SetTrue()
		usr.PasswordHash = strongPasswordHash
		authenticator := user.NewAuthenticationService()

		auth := authenticator.Authenticate(ctx, usr, rawPassword)
		assert.True(t, auth)
	})
}
