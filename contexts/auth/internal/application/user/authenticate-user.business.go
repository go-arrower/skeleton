package user

import (
	"context"

	"github.com/go-arrower/skeleton/contexts/admin"
)

func NewAuthenticationService(settingsService admin.SettingsAPI) *AuthenticationService {
	return &AuthenticationService{settingsService: settingsService}
}

type AuthenticationService struct {
	settingsService admin.SettingsAPI
}

func (s *AuthenticationService) Authenticate(ctx context.Context, usr User, password string) bool {
	if isLoginActive, err := s.settingsService.Setting(ctx, admin.SettingLogin); !isLoginActive.Bool() || err != nil {
		return false
	}

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
