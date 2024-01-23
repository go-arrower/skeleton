package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrVacuumFailed = errors.New("VACUUM failed")

func NewJobsApplication(db *pgxpool.Pool) *JobsApplication {
	return &JobsApplication{db: db}
}

type JobsApplication struct {
	db *pgxpool.Pool
}

func (app *JobsApplication) VacuumJobsTable(ctx context.Context, table string) error {
	if !isValidTable(table) {
		return fmt.Errorf("%w: invalid table: %s", ErrVacuumFailed, table)
	}

	_, err := app.db.Exec(ctx, fmt.Sprintf(`VACUUM FULL arrower.%s`, validTables[table]))
	if err != nil {
		return fmt.Errorf("%w for table: %s: %v", ErrVacuumFailed, table, err)
	}

	return nil
}

var validTables = map[string]string{
	"jobs":    "gue_jobs",
	"history": "gue_jobs_history",
}

func isValidTable(table string) bool {
	var validTable bool
	for k := range validTables {
		if k == table {
			validTable = true
		}
	}

	return validTable
}
