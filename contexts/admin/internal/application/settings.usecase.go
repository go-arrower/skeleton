package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-arrower/skeleton/contexts/admin"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain"
)

var ErrUpdateFailed = errors.New("update setting failed")

func NewSettingsApp(repo domain.SettingRepository) *SettingsApp {
	return &SettingsApp{repo: repo}
}

type SettingsApp struct {
	repo domain.SettingRepository
}

func (app *SettingsApp) Setting(ctx context.Context, setting admin.SettingKey) (admin.SettingValue, error) {
	s, err := app.repo.FindByID(ctx, setting)

	return s.Value, err //nolint:wrapcheck
}

func (app *SettingsApp) Settings(ctx context.Context, settings ...admin.SettingKey) ([]admin.SettingValue, error) {
	return app.repo.Settings(ctx, settings...)
}

func (app *SettingsApp) SettingsByContext(ctx context.Context, context string) ([]admin.SettingValue, error) {
	return app.repo.SettingsByContext(ctx, context)
}

func (app *SettingsApp) Add(ctx context.Context, setting admin.Setting) error {
	if setting.Key == "" {
		return fmt.Errorf("%w: missing setting key", admin.ErrInvalidSetting)
	}

	ex, err := app.repo.Exists(ctx, setting.Key)
	if err != nil {
		return fmt.Errorf("could check existance of setting: %w", err)
	}

	if ex {
		return fmt.Errorf("%w: setting key already exists", admin.ErrInvalidSetting)
	}

	return app.repo.Create(ctx, setting)
}

// func (app *SettingsApp) UpdateIgnoreOptions(ctx context.Context, key admin.SettingKey, value admin.SettingValue) error {

func (app *SettingsApp) UpdateAndGet(ctx context.Context, key admin.SettingKey, value admin.SettingValue) (admin.Setting, error) {
	setting, err := app.repo.FindByID(ctx, key)
	if err != nil {
		return admin.Setting{}, fmt.Errorf("could not find setting: %w", err)
	}

	if setting.UIOptions.ReadOnly {
		return admin.Setting{}, ErrUpdateFailed
	}

	setting.Value = value // FIXME the value from the POST form is always string. That is an issue, if the key is supposed to be something else like bool or float or complex !!!
	// s.SetValueBySettingsType(val)
	err = app.repo.Update(ctx, setting)
	if err != nil {
		return admin.Setting{}, fmt.Errorf("could not update setting: %w", err)
	}

	return setting, nil
}
