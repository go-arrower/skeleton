package user

import (
	"context"
)

func NewAuthenticationService() *AuthenticationService {
	return &AuthenticationService{}
}

type AuthenticationService struct{}

func (s *AuthenticationService) Authenticate(ctx context.Context, usr User, password string) bool {
	if !usr.IsVerified() {
		return false
	}

	if usr.IsBlocked() {
		return false
	}

	if !usr.PasswordHash.Matches(password) {
		return false
	}

	return true
}
