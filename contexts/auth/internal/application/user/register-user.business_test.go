package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin"
	admin_init "github.com/go-arrower/skeleton/contexts/admin/init"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
)

func TestRegistrationService_RegisterNewUser(t *testing.T) {
	t.Parallel()

	t.Run("register setting disabled", func(t *testing.T) {
		t.Parallel()

		settingsService := admin_init.NewMemorySettingsAPI()
		settingsService.Add(ctx, admin.Setting{
			Key:   admin.SettingRegistration,
			Value: admin.NewSettingValue(false),
		})

		rs := user.NewRegistrationService(settingsService, nil)

		_, err := rs.RegisterNewUser(ctx, "", "")
		assert.ErrorIs(t, err, user.ErrRegistrationFailed)
	})

	t.Run("login already in use", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewMemoryRepository()
		_ = repo.Save(ctx, userVerified)

		settingsService := admin_init.NewMemorySettingsAPI()
		settingsService.Add(ctx, admin.Setting{
			Key:   admin.SettingRegistration,
			Value: admin.NewSettingValue(true),
		})

		rs := user.NewRegistrationService(settingsService, repo)

		_, err := rs.RegisterNewUser(ctx, userLogin, "")
		assert.ErrorIs(t, err, user.ErrUserAlreadyExists)
	})

	t.Run("register new user", func(t *testing.T) {
		t.Parallel()

		repo := repository.NewMemoryRepository()

		settingsService := admin_init.NewMemorySettingsAPI()
		settingsService.Add(ctx, admin.Setting{
			Key:   admin.SettingRegistration,
			Value: admin.NewSettingValue(true),
		})

		rs := user.NewRegistrationService(settingsService, repo)

		usr, err := rs.RegisterNewUser(ctx, userLogin, rawPassword)
		assert.NoError(t, err)
		assert.NotEmpty(t, usr)
	})
}
