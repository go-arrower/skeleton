package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-arrower/arrower/app"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

var ErrVacuumFailed = errors.New("VACUUM failed")

func NewVacuumJobTableRequestHandler(db *pgxpool.Pool) app.Request[VacuumJobTableRequest, VacuumJobTableResponse] {
	return &vacuumJobTableRequestHandler{
		db:      db,
		queries: models.New(db),
	}
}

type vacuumJobTableRequestHandler struct {
	db      *pgxpool.Pool
	queries *models.Queries
}

type (
	VacuumJobTableRequest struct {
		Table string // todo require this field
	}

	VacuumJobTableResponse struct {
		Jobs    string
		History string
	}
)

func (h *vacuumJobTableRequestHandler) H(
	ctx context.Context,
	req VacuumJobTableRequest,
) (VacuumJobTableResponse, error) {
	if !isValidTable(req.Table) {
		return VacuumJobTableResponse{}, fmt.Errorf("%w: invalid table: %s", ErrVacuumFailed, req.Table)
	}

	_, err := h.db.Exec(ctx, fmt.Sprintf(`VACUUM FULL arrower.%s`, validTables()[req.Table]))
	if err != nil {
		return VacuumJobTableResponse{}, fmt.Errorf("%w for table: %s: %v", ErrVacuumFailed, req.Table, err)
	}

	size, err := h.queries.JobTableSize(ctx)
	if err != nil {
		return VacuumJobTableResponse{}, fmt.Errorf("could not get new job table size: %v", err)
	}

	return VacuumJobTableResponse{
		Jobs:    size.Jobs,
		History: size.History,
	}, nil
}

func validTables() map[string]string {
	return map[string]string{
		"jobs":    "gue_jobs",
		"history": "gue_jobs_history",
	}
}

func isValidTable(table string) bool {
	var validTable bool

	for k := range validTables() {
		if k == table {
			validTable = true
		}
	}

	return validTable
}
