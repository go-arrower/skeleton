package user_test

import (
	"context"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

var (
	ctx                   = context.Background()
	rawPassword           = "0Secret!"
	strongPasswordHash, _ = user.NewPasswordHash(rawPassword)
)

// newUser returns a new User, so you don't have to worry about changing fields when verifying.
func newUser() user.User {
	return user.User{
		ID:       user.NewID(),
		Verified: user.BoolFlag{}.SetFalse(),
	}
}

// used by RegistrationService
const (
	userID    = user.ID("00000000-0000-0000-0000-000000000000")
	userLogin = "0@test.com"
)

var (
	userVerified = user.User{
		ID:           userID,
		Login:        userLogin,
		PasswordHash: strongPasswordHash,
		Verified:     user.BoolFlag{}.SetTrue(),
	}
)
