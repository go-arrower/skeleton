package user

import (
	"context"

	"github.com/go-arrower/arrower/setting"

	"github.com/go-arrower/skeleton/contexts/admin"
)

func NewAuthenticationService(settings setting.Settings) *AuthenticationService {
	return &AuthenticationService{settingsService: settings}
}

type AuthenticationService struct {
	settingsService setting.Settings
}

func (s *AuthenticationService) Authenticate(ctx context.Context, usr User, password string) bool {
	if isLoginActive, err := s.settingsService.Setting(ctx, admin.SettingLogin); !isLoginActive.MustBool() || err != nil {
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
