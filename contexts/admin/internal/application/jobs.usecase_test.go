//go:build integration

package application_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-arrower/arrower/tests"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

var pgHandler *tests.PostgresDocker

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
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := app.VacuumJobsTable(context.Background(), tc.table)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
