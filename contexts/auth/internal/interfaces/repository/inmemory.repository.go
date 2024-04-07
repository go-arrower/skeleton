package repository

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-arrower/arrower/tests"
	"github.com/google/uuid"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		MemoryRepository: tests.NewMemoryRepository[user.User, user.ID](),
		tokens:           make(map[uuid.UUID]user.VerificationToken),
	}
}

type MemoryRepository struct {
	*tests.MemoryRepository[user.User, user.ID]

	tokens map[uuid.UUID]user.VerificationToken
}

func (repo *MemoryRepository) All(ctx context.Context, filter user.Filter) ([]user.User, error) {
	all, err := repo.MemoryRepository.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Login < all[j].Login
	})

	users := []user.User{}
	limit := filter.Limit
	found := false

	for _, u := range all {
		// skipp Logins before the offset
		if !found && filter.Offset != "" && u.Login != filter.Offset {
			continue
		}

		// skip the found = same as offset
		found = true
		if filter.Offset == u.Login {
			continue
		}

		// append up to the limit
		users = append(users, u)
		limit--

		if limit == 0 {
			return users, nil
		}
	}

	return users, nil
}

func (repo *MemoryRepository) FindByLogin(ctx context.Context, login user.Login) (user.User, error) {
	all, _ := repo.MemoryRepository.All(ctx)

	for _, u := range all {
		if u.Login == login {
			return u, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (repo *MemoryRepository) ExistsByLogin(ctx context.Context, login user.Login) (bool, error) {
	all, _ := repo.MemoryRepository.All(ctx)

	for _, u := range all {
		if u.Login == login {
			return true, nil
		}
	}

	return false, user.ErrNotFound
}

func (repo *MemoryRepository) CreateVerificationToken(
	ctx context.Context,
	token user.VerificationToken,
) error {
	if token.Token().String() == "" {
		return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
	}

	repo.Lock()
	defer repo.Unlock()

	repo.tokens[token.Token()] = token

	return nil
}

func (repo *MemoryRepository) VerificationTokenByToken(
	ctx context.Context,
	tokenID uuid.UUID,
) (user.VerificationToken, error) {
	for _, t := range repo.tokens {
		if t.Token() == tokenID {
			return t, nil
		}
	}

	return user.VerificationToken{}, user.ErrNotFound
}

var _ user.Repository = (*MemoryRepository)(nil)
