package repository

import (
	"context"

	"github.com/go-arrower/arrower/tests"

	"github.com/go-arrower/skeleton/contexts/admin"
)

func NewSettingsMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		MemoryRepository: tests.NewMemoryRepository[admin.Setting, admin.SettingKey](tests.WithIDField("Key")),
	}
}

type MemoryRepository struct {
	*tests.MemoryRepository[admin.Setting, admin.SettingKey]
}

func (repo *MemoryRepository) Settings(ctx context.Context, settings ...admin.SettingKey) ([]admin.SettingValue, error) { //nolint:lll
	setts, err := repo.FindByIDs(ctx, settings)

	ret := make([]admin.SettingValue, len(setts))
	for i, s := range setts {
		ret[i] = s.Value
	}

	return ret, err //nolint:wrapcheck
}

func (repo *MemoryRepository) SettingsByContext(ctx context.Context, context string) ([]admin.SettingValue, error) {
	all, _ := repo.All(ctx)

	var ret []admin.SettingValue

	for _, s := range all {
		if s.Key.Context() == context {
			ret = append(ret, s.Value)
		}
	}

	return ret, nil
}
