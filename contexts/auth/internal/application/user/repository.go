package user

import (
	"context"
	"errors"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrPersistenceFailed = errors.New("persistence operation failed")
)

type Repository interface {
	All(context.Context) ([]User, error)
	AllByIDs(context.Context, []ID) ([]User, error)

	FindByID(context.Context, ID) (User, error)
	FindByLogin(context.Context, Login) (User, error)
	ExistsByID(context.Context, ID) (bool, error)
	ExistsByLogin(context.Context, Login) (bool, error)

	Count(context.Context) (int, error)

	Save(context.Context, User) error
	SaveAll(context.Context, []User) error

	Delete(context.Context, User) error
	DeleteByID(context.Context, ID) error
	DeleteByIDs(context.Context, []ID) error
	DeleteAll(context.Context) error
}
