package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

var ErrVacuumFailed = errors.New("VACUUM failed")

func NewJobsApplication(db *pgxpool.Pool) *JobsApplication {
	return &JobsApplication{
		db:      db,
		queries: models.New(db),
	}
}

type JobsApplication struct {
	db      *pgxpool.Pool
	queries *models.Queries
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

func (app *JobsApplication) PruneHistory(ctx context.Context, days int) error {
	deleteBefore := time.Now().Add(-1 * time.Duration(days) * time.Hour * 24)

	err := app.queries.PruneHistory(ctx, pgtype.Timestamptz{Time: deleteBefore, Valid: true})
	if err != nil {
		return fmt.Errorf("could not delete old history: %v", err)
	}

	return nil
}
