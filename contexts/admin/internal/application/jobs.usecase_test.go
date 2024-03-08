//go:build integration

package application_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-arrower/arrower/tests"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

var (
	ctx       = context.Background()
	pgHandler *tests.PostgresDocker
)

func TestMain(m *testing.M) {
	pgHandler = tests.GetPostgresDockerForIntegrationTestingInstance()

	//
	// Run tests
	code := m.Run()

	pgHandler.Cleanup()
	os.Exit(code)
}

func TestJobsApplication_VacuumJobsTable(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		table string
		err   error
	}{
		"empty table": {
			table: "",
			err:   application.ErrVacuumFailed,
		},
		"non existing table": {
			table: "non-existing-table",
			err:   application.ErrVacuumFailed,
		},
		"jobs": {
			table: "jobs",
			err:   nil,
		},
		"history": {
			table: "history",
			err:   nil,
		},
	}

	// share one database for all tests, as it is about vacuum and not modifying data
	app := application.NewJobsApplication(pgHandler.PGx())

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := app.VacuumJobsTable(context.Background(), tc.table)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestJobsApplication_PruneHistory(t *testing.T) {
	t.Parallel()

	t.Run("older than", func(t *testing.T) {
		t.Parallel()

		pg := pgHandler.NewTestDatabase("testdata/fixtures/prune_history.yaml")
		app := application.NewJobsApplication(pg)

		err := app.PruneHistory(ctx, 7)
		assert.NoError(t, err)

		assertTableNumberOfRows(t, pg, "arrower.gue_jobs_history", 1)
	})

	t.Run("all", func(t *testing.T) {
		t.Parallel()

		pg := pgHandler.NewTestDatabase("testdata/fixtures/prune_history.yaml")
		app := application.NewJobsApplication(pg)

		err := app.PruneHistory(ctx, 0)
		assert.NoError(t, err)

		assertTableNumberOfRows(t, pg, "arrower.gue_jobs_history", 0)
	})

}

func assertTableNumberOfRows(t *testing.T, db *pgxpool.Pool, table string, num int) {
	t.Helper()

	var c int
	_ = db.QueryRow(context.Background(), fmt.Sprintf(`SELECT COUNT(*) FROM %s;`, table)).Scan(&c)

	assert.Equal(t, num, c)
}
