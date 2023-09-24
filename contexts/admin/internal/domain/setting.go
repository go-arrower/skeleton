package domain

import (
	"context"

	"github.com/go-arrower/skeleton/contexts/admin"
)

type SettingRepository interface {
	Create(context.Context, admin.Setting) error
	Update(context.Context, admin.Setting) error

	Settings(ctx context.Context, settings ...admin.SettingKey) ([]admin.SettingValue, error)
	SettingsByContext(ctx context.Context, context string) ([]admin.SettingValue, error)
	FindByID(context.Context, admin.SettingKey) (admin.Setting, error)
	Exists(context.Context, admin.SettingKey) (bool, error)
	All(context.Context) ([]admin.Setting, error)
}

type Setting struct{}

func (s *Setting) Reset() {}

func (s *Setting) SetValue() {}
