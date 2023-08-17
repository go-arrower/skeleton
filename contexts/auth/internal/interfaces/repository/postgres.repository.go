package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-arrower/arrower/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

var ErrMissingConnection = errors.New("missing db connection")

func NewPostgresRepository(pg *pgxpool.Pool) (*PostgresRepository, error) {
	if pg == nil {
		return nil, ErrMissingConnection
	}

	return &PostgresRepository{
		db: postgres.NewPostgresBaseRepository(models.New(pg)),
	}, nil
}

type PostgresRepository struct {
	db postgres.BaseRepository[*models.Queries]
}

func (repo *PostgresRepository) All(ctx context.Context) ([]user.User, error) {
	dbUser, err := repo.db.Conn(ctx).AllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", user.ErrNotFound, err)
	}

	return usersFromModel(ctx, repo.db.Conn(ctx), dbUser)
}

func (repo *PostgresRepository) AllByIDs(ctx context.Context, ids []user.ID) ([]user.User, error) {
	dbIDs := make([]uuid.UUID, len(ids))

	var err error
	for i, id := range ids {
		dbIDs[i], err = uuid.Parse(string(id))
		if err != nil {
			return nil, fmt.Errorf("%w: could not parse as uuid: %s: %w", user.ErrNotFound, id, err)
		}
	}

	dbUser, err := repo.db.Conn(ctx).AllUsersByIDs(ctx, dbIDs)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", user.ErrNotFound, err)
	}

	return usersFromModel(ctx, repo.db.Conn(ctx), dbUser)
}

func (repo *PostgresRepository) FindByID(ctx context.Context, id user.ID) (user.User, error) {
	dbID, err := uuid.Parse(string(id))
	if err != nil {
		return user.User{}, fmt.Errorf("%w: could not parse as uuid: %s: %w", user.ErrNotFound, id, err)
	}

	dbUser, err := repo.db.Conn(ctx).FindUserByID(ctx, dbID)
	if err != nil {
		return user.User{}, fmt.Errorf("%w: could not find user by id: %s : %w", user.ErrNotFound, dbID, err)
	}

	return userFromModel(ctx, repo.db.Conn(ctx), dbUser)
}

func (repo *PostgresRepository) FindByLogin(ctx context.Context, login user.Login) (user.User, error) {
	dbUser, err := repo.db.Conn(ctx).FindUserByLogin(ctx, string(login))
	if err != nil {
		return user.User{}, fmt.Errorf("%w: could not find user by login: %s : %w", user.ErrNotFound, login, err)
	}

	return userFromModel(ctx, repo.db.Conn(ctx), dbUser)
}

func (repo *PostgresRepository) ExistsByID(ctx context.Context, id user.ID) (bool, error) {
	dbID, err := uuid.Parse(string(id))
	if err != nil {
		return false, fmt.Errorf("%w: could not parse as uuid: %s: %w", user.ErrNotFound, id, err)
	}

	ex, err := repo.db.Conn(ctx).UserExistsByID(ctx, dbID)
	if err != nil {
		return false, fmt.Errorf("%w: %w", user.ErrNotFound, err)
	}

	return ex, nil
}

func (repo *PostgresRepository) ExistsByLogin(ctx context.Context, login user.Login) (bool, error) {
	ex, err := repo.db.Conn(ctx).UserExistsByLogin(ctx, string(login))
	if err != nil {
		return false, fmt.Errorf("%w: %w", user.ErrNotFound, err)
	}

	return ex, nil
}

func (repo *PostgresRepository) Count(ctx context.Context) (int, error) {
	c, err := repo.db.Conn(ctx).CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w: could not count users: %w", user.ErrNotFound, err)
	}

	return int(c), nil
}

func (repo *PostgresRepository) Save(ctx context.Context, usr user.User) error {
	if usr.ID == "" {
		return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
	}

	_, err := repo.db.ConnOrTX(ctx).UpsertUser(ctx, userToModel(usr))
	if err != nil {
		return fmt.Errorf("%w: could not save user: %s: %w", user.ErrPersistenceFailed, usr.ID, err)
	}

	return nil
}

func (repo *PostgresRepository) SaveAll(ctx context.Context, users []user.User) error {
	for _, usr := range users {
		if usr.ID == "" {
			return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
		}

		_, err := repo.db.ConnOrTX(ctx).UpsertUser(ctx, userToModel(usr))
		if err != nil {
			return fmt.Errorf("%w: could not save user: %s: %w", user.ErrPersistenceFailed, usr.ID, err)
		}
	}

	return nil
}

func (repo *PostgresRepository) Delete(ctx context.Context, usr user.User) error {
	if usr.ID == "" {
		return fmt.Errorf("missing ID: %w", user.ErrPersistenceFailed)
	}

	id, err := uuid.Parse(string(usr.ID))
	if err != nil {
		return fmt.Errorf("%w: could not parse as uuid: %s: %w", user.ErrPersistenceFailed, id, err)
	}

	err = repo.db.ConnOrTX(ctx).DeleteUser(ctx, []uuid.UUID{id})
	if err != nil {
		return fmt.Errorf("%w: could not delete user: %s: %w", user.ErrPersistenceFailed, usr.ID, err)
	}

	return nil
}

func (repo *PostgresRepository) DeleteByID(ctx context.Context, userID user.ID) error {
	id, err := uuid.Parse(string(userID))
	if err != nil {
		return fmt.Errorf("%w: could not parse as uuid: %s: %w", user.ErrPersistenceFailed, id, err)
	}

	err = repo.db.ConnOrTX(ctx).DeleteUser(ctx, []uuid.UUID{id})
	if err != nil {
		return fmt.Errorf("%w: could not delete user: %s: %w", user.ErrPersistenceFailed, userID, err)
	}

	return nil
}

func (repo *PostgresRepository) DeleteByIDs(ctx context.Context, ids []user.ID) error {
	dbIDs := make([]uuid.UUID, len(ids))

	var err error
	for i, id := range ids {
		dbIDs[i], err = uuid.Parse(string(id))
		if err != nil {
			return fmt.Errorf("%w: could not parse as uuid: %s: %w", user.ErrPersistenceFailed, id, err)
		}
	}

	err = repo.db.ConnOrTX(ctx).DeleteUser(ctx, dbIDs)
	if err != nil {
		return fmt.Errorf("%w: could not delete users: %w", user.ErrPersistenceFailed, err)
	}

	return nil
}

func (repo *PostgresRepository) DeleteAll(ctx context.Context) error {
	err := repo.db.ConnOrTX(ctx).DeleteAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("%w: could not delete all users: %w", user.ErrPersistenceFailed, err)
	}

	return nil
}

var _ user.Repository = (*PostgresRepository)(nil)
