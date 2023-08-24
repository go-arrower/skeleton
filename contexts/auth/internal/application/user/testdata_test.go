package user_test

import (
	"context"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

var (
	ctx                   = context.Background()
	rawPassword           = "secret"
	strongPasswordHash, _ = user.NewPasswordHash(rawPassword)
)

// newUser returns a new User, so you don't have to worry about changing fields when verifying.
func newUser() user.User {
	return user.User{
		ID:       user.NewID(),
		Verified: user.BoolFlag{}.SetFalse(),
	}
}
