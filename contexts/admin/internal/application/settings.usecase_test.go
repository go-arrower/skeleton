package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin"
	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository"
)

var ctx = context.Background()

func TestSettingsApp_Add(t *testing.T) {
	t.Parallel()

	t.Run("missing key", func(t *testing.T) {
		t.Parallel()

		uc := application.NewSettingsApp(repository.NewSettingsMemoryRepository())

		err := uc.Add(ctx, admin.Setting{})
		assert.ErrorIs(t, err, admin.ErrInvalidSetting)
	})

	t.Run("add new setting", func(t *testing.T) {
		t.Parallel()

		uc := application.NewSettingsApp(repository.NewSettingsMemoryRepository())

		err := uc.Add(ctx, admin.Setting{
			Key:       admin.NewSettingKey("", "someKey"),
			Value:     "",
			UIOptions: admin.Options{},
		})
		assert.NoError(t, err)
	})

	t.Run("setting already exists", func(t *testing.T) {
		t.Parallel()

		uc := application.NewSettingsApp(repository.NewSettingsMemoryRepository())

		setting := admin.Setting{
			Key:       admin.NewSettingKey("", "someKey"),
			Value:     admin.NewSettingValue("someValue"),
			UIOptions: admin.Options{},
		}

		err := uc.Add(ctx, setting)
		assert.NoError(t, err)

		err = uc.Add(ctx, setting)
		assert.ErrorIs(t, err, admin.ErrInvalidSetting)
	})

	// todo validate setting & options ect
}

func TestSettingsApp_UpdateAndGet(t *testing.T) {
	t.Parallel()

	// todo setting key does not exist => error

	t.Run("setting is read only", func(t *testing.T) {
		t.Parallel()

		uc := application.NewSettingsApp(repository.NewSettingsMemoryRepository())
		opt := defaultSetting
		opt.UIOptions.ReadOnly = true
		uc.Add(ctx, opt)

		_, err := uc.UpdateAndGet(ctx, defaultKey, admin.NewSettingValue("someVal"))
		assert.ErrorIs(t, err, application.ErrUpdateFailed)
	})

	// todo ensure validator funcs are applied

	t.Run("update succeeds", func(t *testing.T) {
		t.Parallel()

		uc := application.NewSettingsApp(repository.NewSettingsMemoryRepository())
		uc.Add(ctx, defaultSetting)

		newVal := admin.NewSettingValue("someVal")
		setting, err := uc.UpdateAndGet(ctx, defaultKey, newVal)
		assert.NoError(t, err)
		assert.Equal(t, setting.Value, newVal)
	})
}

// --- --- --- TEST DATA --- --- ---

var (
	defaultKey     = admin.NewSettingKey("", "someKey")
	defaultSetting = admin.Setting{
		Key:       defaultKey,
		Value:     "",
		UIOptions: admin.Options{},
	}
)
