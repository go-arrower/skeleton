package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-arrower/skeleton/contexts/admin"
)

var (
	ErrRegistrationFailed = errors.New("registration failed")
	ErrUserAlreadyExists  = fmt.Errorf("%w: user already exists", ErrRegistrationFailed)
)

func NewRegistrationService(settingsService admin.SettingsAPI, repo Repository) *RegistrationService {
	return &RegistrationService{
		settingsService: settingsService,
		repo:            repo,
	}
}

type RegistrationService struct {
	settingsService admin.SettingsAPI
	repo            Repository
}

func (s *RegistrationService) RegisterNewUser(
	ctx context.Context,
	registerEmail string,
	password string,
) (User, error) {
	isRegistrationActive, err := s.settingsService.Setting(ctx, admin.SettingRegistration)
	if err != nil {
		return User{}, fmt.Errorf("%w: could not load settings: %v", ErrRegistrationFailed, err)
	}

	if !isRegistrationActive.Bool() {
		return User{}, fmt.Errorf("%w: registration is disabled", ErrRegistrationFailed)
	}

	ex, err := s.repo.ExistsByLogin(ctx, Login(registerEmail))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return User{}, fmt.Errorf("could not check if user exists: %w", err)
	}

	if ex {
		return User{}, ErrUserAlreadyExists
	}

	usr, err := NewUser(registerEmail, password)
	if err != nil {
		return User{}, fmt.Errorf("%w: could not create new user", ErrRegistrationFailed)
	}

	return usr, nil
}
