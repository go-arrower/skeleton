package user_test

import (
	"testing"

	"github.com/go-arrower/skeleton/contexts/auth"

	"github.com/go-arrower/arrower/setting"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
)

func TestRegistrationService_RegisterNewUser(t *testing.T) {
	t.Parallel()

	t.Run("register setting disabled", func(t *testing.T) {
		t.Parallel()

		settings := setting.NewInMemorySettings()
		settings.Save(ctx, auth.SettingAllowRegistration, setting.NewValue(false))

		rs := user.NewRegistrationService(settings, nil)

		_, err := rs.RegisterNewUser(ctx, "", "")
		assert.ErrorIs(t, err, user.ErrRegistrationFailed)
	})

	t.Run("login already in use", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewMemoryRepository()
		_ = repo.Save(ctx, userVerified)

		settings := setting.NewInMemorySettings()
		settings.Save(ctx, auth.SettingAllowRegistration, setting.NewValue(true))

		rs := user.NewRegistrationService(settings, repo)

		_, err := rs.RegisterNewUser(ctx, userLogin, "")
		assert.ErrorIs(t, err, user.ErrUserAlreadyExists)
	})

	t.Run("register new user", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewMemoryRepository()

		settings := setting.NewInMemorySettings()
		settings.Save(ctx, auth.SettingAllowRegistration, setting.NewValue(true))

		rs := user.NewRegistrationService(settings, repo)

		usr, err := rs.RegisterNewUser(ctx, userLogin, rawPassword)
		assert.NoError(t, err)
		assert.NotEmpty(t, usr)
	})
}
