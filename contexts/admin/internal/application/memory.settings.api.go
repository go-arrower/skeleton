package application

import (
	"context"

	"github.com/go-arrower/skeleton/contexts/admin"

	"github.com/go-arrower/arrower/tests"
)

func NewMemorySettings() *MemorySettings {
	return &MemorySettings{
		MemoryRepository: tests.NewMemoryRepository[admin.Setting, admin.SettingKey](tests.WithIDField("Key")),
	}
}

type MemorySettings struct {
	*tests.MemoryRepository[admin.Setting, admin.SettingKey]
}

func (repo *MemorySettings) Setting(ctx context.Context, setting admin.SettingKey) (admin.SettingValue, error) {
	s, err := repo.FindByID(ctx, setting)

	return s.Value, err //nolint:wrapcheck
}

func (repo *MemorySettings) Settings(ctx context.Context, settings ...admin.SettingKey) ([]admin.SettingValue, error) {
	setts, err := repo.FindByIDs(ctx, settings)

	ret := make([]admin.SettingValue, len(setts))
	for i, s := range setts {
		ret[i] = s.Value
	}

	return ret, err //nolint:wrapcheck
}

func (repo *MemorySettings) SettingsByContext(ctx context.Context, context string) ([]admin.SettingValue, error) {
	all, _ := repo.All(ctx)

	var ret []admin.SettingValue

	for _, s := range all {
		if s.Key.Context() == context {
			ret = append(ret, s.Value)
		}
	}

	return ret, nil
}

var _ admin.SettingsAPI = (*MemorySettings)(nil)
