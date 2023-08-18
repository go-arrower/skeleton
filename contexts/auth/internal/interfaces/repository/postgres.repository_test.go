//go:build integration

package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-arrower/arrower/postgres"
	"github.com/go-arrower/arrower/tests"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/testdata"
)

var (
	ctx       = context.Background()
	pgHandler *postgres.Handler
)

func TestMain(m *testing.M) {
	handler, cleanup := tests.GetDBConnectionForIntegrationTesting(ctx)
	pgHandler = handler

	//
	// Run tests
	code := m.Run()

	//
	// Cleanup
	_ = handler.Shutdown(ctx)
	_ = cleanup()

	os.Exit(code)
}

func TestNewPostgresRepository(t *testing.T) {
	t.Parallel()

	_, err := repository.NewPostgresRepository(nil)
	assert.ErrorIs(t, err, repository.ErrMissingConnection)
}

func TestPostgresRepository_All(t *testing.T) {
	t.Parallel()

	pg := tests.PrepareTestDatabase(pgHandler).PGx
	repo, _ := repository.NewPostgresRepository(pg)

	all, err := repo.All(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 3)
	assert.Len(t, all[0].Sessions, 1, "user should have its value objects returned")
}

func TestPostgresRepository_AllByIDs(t *testing.T) {
	t.Parallel()

	t.Run("valid ids", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		all, err := repo.AllByIDs(ctx, nil)
		assert.NoError(t, err)
		assert.Empty(t, all)

		all, err = repo.AllByIDs(ctx, []user.ID{testdata.UserIDZero, testdata.UserIDOne})
		assert.NoError(t, err)
		assert.Len(t, all, 2)
		assert.Len(t, all[0].Sessions, 1, "user should have its value objects returned")
	})

	t.Run("invalid ids", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		_, err := repo.AllByIDs(ctx, []user.ID{"invalid-id"})
		assert.ErrorIs(t, err, user.ErrNotFound)
	})
}

func TestPostgresRepository_FindByID(t *testing.T) {
	t.Parallel()

	t.Run("valid user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		u, err := repo.FindByID(ctx, testdata.UserIDOne)
		assert.NoError(t, err)
		assert.Equal(t, testdata.UserIDOne, u.ID)
	})

	t.Run("invalid user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		_, err := repo.FindByID(ctx, testdata.UserIDNotValid)
		assert.ErrorIs(t, err, user.ErrNotFound)
	})
}

func TestPostgresRepository_FindByLogin(t *testing.T) {
	t.Parallel()

	t.Run("valid user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		u, err := repo.FindByLogin(ctx, testdata.ValidLogin)
		assert.NoError(t, err)
		assert.Equal(t, testdata.ValidLogin, u.Login)
	})

	t.Run("invalid user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		_, err := repo.FindByLogin(ctx, testdata.NotExLogin)
		assert.ErrorIs(t, err, user.ErrNotFound)
	})
}

func TestPostgresRepository_ExistsByID(t *testing.T) {
	t.Parallel()

	t.Run("user exists", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		ex, err := repo.ExistsByID(ctx, testdata.UserIDZero)
		assert.NoError(t, err)
		assert.True(t, ex)
	})

	t.Run("user does not exist", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		ex, err := repo.ExistsByID(ctx, testdata.UserIDNotExists)
		assert.NoError(t, err)
		assert.False(t, ex)
	})

	t.Run("invalid user id", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		_, err := repo.ExistsByID(ctx, testdata.UserIDNotValid)
		assert.ErrorIs(t, err, user.ErrNotFound)
	})
}

func TestPostgresRepository_ExistsByLogin(t *testing.T) {
	t.Parallel()

	t.Run("user exists", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		ex, err := repo.ExistsByLogin(ctx, testdata.ValidLogin)
		assert.NoError(t, err)
		assert.True(t, ex)
	})

	t.Run("user does not exist", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		ex, err := repo.ExistsByLogin(ctx, testdata.NotExLogin)
		assert.NoError(t, err)
		assert.False(t, ex)
	})
}

func TestPostgresRepository_Count(t *testing.T) {
	t.Parallel()

	pg := tests.PrepareTestDatabase(pgHandler).PGx
	repo, _ := repository.NewPostgresRepository(pg)

	c, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, c)
}

func TestPostgresRepository_Save(t *testing.T) {
	t.Parallel()

	t.Run("save new user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		err := repo.Save(ctx, user.User{
			ID: testdata.UserIDNew,
			Sessions: []user.Session{
				{
					ID:        "some-new-session-key",
					Device:    user.NewDevice(testdata.UserAgent),
					CreatedAt: time.Now().UTC(),
				},
			},
		})
		assert.NoError(t, err)

		c, _ := repo.Count(ctx)
		assert.Equal(t, 4, c)

		queries := models.New(pg)
		sessions, _ := queries.AllSessions(ctx)
		assert.Len(t, sessions, 1+1) // 1 session is already created via _common.yaml fixtures
	})

	t.Run("save existing user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		usr, _ := repo.FindByID(ctx, testdata.UserIDZero)
		assert.Empty(t, usr.Name)

		usr.Name = user.NewName("firstName", "", "")
		err := repo.Save(ctx, usr)
		assert.NoError(t, err)

		usr, _ = repo.FindByID(ctx, testdata.UserIDZero)
		assert.NotEmpty(t, usr.Name)
	})

	t.Run("save empty user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		err := repo.Save(ctx, user.User{})
		assert.ErrorIs(t, err, user.ErrPersistenceFailed)
	})
}

func TestPostgresRepository_SaveAll(t *testing.T) {
	t.Parallel()

	t.Run("save multiple", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		newUser := user.User{ID: testdata.UserIDNew}
		updatedUser := testdata.UserZero
		updatedUser.Name = user.NewName("firstName", "", "")

		err := repo.SaveAll(ctx, []user.User{
			newUser,
			updatedUser,
		})
		assert.NoError(t, err)

		c, _ := repo.Count(ctx)
		assert.Equal(t, 4, c)
		u, _ := repo.FindByID(ctx, testdata.UserIDZero)
		assert.NotEmpty(t, u.Name)
	})
}

func TestPostgresRepository_Delete(t *testing.T) {
	t.Parallel()

	t.Run("delete user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		err := repo.Delete(ctx, testdata.UserZero)
		assert.NoError(t, err)

		c, _ := repo.Count(ctx)
		assert.Equal(t, 2, c)
	})
}

func TestPostgresRepository_DeleteByID(t *testing.T) {
	t.Parallel()

	t.Run("delete user", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		err := repo.DeleteByID(ctx, testdata.UserIDZero)
		assert.NoError(t, err)

		c, _ := repo.Count(ctx)
		assert.Equal(t, 2, c)
	})

	t.Run("invalid id", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		err := repo.DeleteByID(ctx, testdata.UserIDNotValid)
		assert.ErrorIs(t, err, user.ErrPersistenceFailed)
	})
}

func TestPostgresRepository_DeleteByIDs(t *testing.T) {
	t.Parallel()

	t.Run("delete users", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		err := repo.DeleteByIDs(ctx, []user.ID{
			testdata.UserIDZero,
			testdata.UserIDOne,
		})
		assert.NoError(t, err)

		c, _ := repo.Count(ctx)
		assert.Equal(t, 1, c)
	})
}

func TestPostgresRepository_DeleteAll(t *testing.T) {
	t.Parallel()

	t.Run("delete all users", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		err := repo.DeleteAll(ctx)
		assert.NoError(t, err)

		c, _ := repo.Count(ctx)
		assert.Equal(t, 0, c)
	})
}

func TestPostgresRepository_CreateVerificationToken(t *testing.T) {
	t.Parallel()

	t.Run("create new token", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx
		repo, _ := repository.NewPostgresRepository(pg)

		err := repo.CreateVerificationToken(ctx, testdata.ValidToken)
		assert.NoError(t, err)

		tok, err := repo.VerificationTokenByToken(ctx, testdata.ValidToken.Token())
		assert.NoError(t, err)
		assert.Equal(t, testdata.ValidToken.Token(), tok.Token())
	})
}
