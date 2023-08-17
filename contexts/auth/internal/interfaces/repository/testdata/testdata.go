package testdata

import "github.com/go-arrower/skeleton/contexts/auth/internal/application/user"

const (
	UserIDZero = user.ID("00000000-0000-0000-0000-000000000000")
	UserIDOne  = user.ID("00000000-0000-0000-0000-000000000001")

	UserIDNew       = user.ID("00000000-0000-0000-0000-000000000010")
	UserIDNotExists = user.ID("00000000-0000-0000-0000-999999999999")
	UserIDNotValid  = user.ID("invalid-id")

	ValidLogin = user.Login("0@test.com")
	NotExLogin = user.Login("invalid-login")
)

var (
	UserZero = user.User{
		ID:    UserIDZero,
		Login: "0@test.com",
	}
)
