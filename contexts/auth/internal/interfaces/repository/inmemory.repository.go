package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:  make(map[user.ID]user.User),
		tokens: make(map[uuid.UUID]user.VerificationToken),
		mu:     sync.Mutex{},
	}
}

type MemoryRepository struct {
	users  map[user.ID]user.User
	tokens map[uuid.UUID]user.VerificationToken
	mu     sync.Mutex
}

func (repo *MemoryRepository) All(ctx context.Context) ([]user.User, error) {
	users := []user.User{}

	for _, u := range repo.users {
		users = append(users, u)
	}

	return users, nil
}

func (repo *MemoryRepository) AllByIDs(ctx context.Context, ids []user.ID) ([]user.User, error) {
	users := []user.User{}

	for _, usr := range repo.users {
		for _, id := range ids {
			if usr.ID == "" {
				return nil, fmt.Errorf("missing ID: %w", user.ErrNotFound)
			}

			if usr.ID == id {
				users = append(users, usr)
			}
		}
	}

	return users, nil
}

func (repo *MemoryRepository) FindByID(ctx context.Context, id user.ID) (user.User, error) {
	for _, u := range repo.users {
		if u.ID == id {
			return u, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (repo *MemoryRepository) FindByLogin(ctx context.Context, login user.Login) (user.User, error) {
	for _, u := range repo.users {
		if u.Login == login {
			return u, nil
		}
	}

	return user.User{}, user.ErrNotFound
}

func (repo *MemoryRepository) ExistsByID(ctx context.Context, id user.ID) (bool, error) {
	for _, u := range repo.users {
		if u.ID == id {
			return true, nil
		}
	}

	return false, user.ErrNotFound
}

func (repo *MemoryRepository) ExistsByLogin(ctx context.Context, login user.Login) (bool, error) {
	for _, u := range repo.users {
		if u.Login == login {
			return true, nil
		}
	}

	return false, user.ErrNotFound
}

func (repo *MemoryRepository) Count(ctx context.Context) (int, error) {
	return len(repo.users), nil
}

func (repo *MemoryRepository) Save(ctx context.Context, usr user.User) error {
	if usr.ID == "" {
		return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.users[usr.ID] = usr

	return nil
}

func (repo *MemoryRepository) SaveAll(ctx context.Context, users []user.User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	var errs []error

	for _, usr := range users {
		if usr.ID == "" {
			errs = append(errs, fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed))

			continue
		}

		repo.users[usr.ID] = usr
	}

	return errors.Join(errs...)
}

func (repo *MemoryRepository) Delete(ctx context.Context, usr user.User) error {
	if usr.ID == "" {
		return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, u := range repo.users {
		if u.ID == usr.ID {
			delete(repo.users, usr.ID)

			return nil
		}
	}

	return user.ErrPersistenceFailed
}

func (repo *MemoryRepository) DeleteByID(ctx context.Context, id user.ID) error {
	if id == "" {
		return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	var errs []error

	for _, usr := range repo.users {
		if usr.ID == id {
			if usr.ID == "" {
				errs = append(errs, fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed))

				continue
			}

			delete(repo.users, usr.ID)

			return nil
		}
	}

	return errors.Join(errs...)
}

func (repo *MemoryRepository) DeleteByIDs(ctx context.Context, ids []user.ID) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, usr := range repo.users {
		for _, id := range ids {
			if id == "" {
				return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
			}

			if usr.ID == id {
				delete(repo.users, usr.ID)
			}
		}
	}

	return nil
}

func (repo *MemoryRepository) DeleteAll(ctx context.Context) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.users = make(map[user.ID]user.User)

	return nil
}

func (repo *MemoryRepository) CreateVerificationToken(
	ctx context.Context,
	token user.VerificationToken,
) error {
	if token.Token().String() == "" {
		return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

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
