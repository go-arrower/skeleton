package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/propagation"

	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

//go:generate go run github.com/hexdigest/gowrap/cmd/gowrap gen -p github.com/go-arrower/skeleton/contexts/admin/internal/application -g -i JobsApplication -t ./templates/slog.html -o jobs.log.usecase.go
type JobsApplication interface {
	Queues(ctx context.Context) (jobs.QueueNames, error)
	ScheduleJobs(ctx context.Context, in ScheduleJobsRequest) error
	RescheduleJob(ctx context.Context, in RescheduleJobRequest) error
}

func NewJobsApplication(
	db *pgxpool.Pool,
	queries *models.Queries,
	repo jobs.Repository,
) *JobsUsecase {
	return &JobsUsecase{
		db:      db,      // todo remove and use repo instead?
		queries: queries, //  todo remove and use repo instead?

		repo: repo,
	}
}

type JobsUsecase struct {
	db      *pgxpool.Pool
	queries *models.Queries // FIXME violates clean arch checker

	repo jobs.Repository
}

var _ JobsApplication = (*JobsUsecase)(nil)

// Queues returns a list of all known Queues.
func (app *JobsUsecase) Queues(ctx context.Context) (jobs.QueueNames, error) {
	return app.repo.Queues(ctx)
}

type (
	ScheduleJobsRequest struct {
		Queue    string
		JobType  string
		Priority int16
		Payload  string
		Count    int
		RunAt    time.Time
	}
	ScheduleJobsResponse struct{}
)

func (app *JobsUsecase) ScheduleJobs(ctx context.Context, in ScheduleJobsRequest) error {
	carrier := propagation.MapCarrier{}
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

	propagator.Inject(ctx, carrier)

	_, err := app.queries.ScheduleJobs(ctx, buildJobs(in, carrier))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

type (
	RescheduleJobRequest struct {
		JobID string
	}
)

func (app *JobsUsecase) RescheduleJob(ctx context.Context, in RescheduleJobRequest) error {
	err := app.repo.RunJobAt(ctx, in.JobID, time.Now())

	return fmt.Errorf("%w", err)
}
