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
	ListAllQueues(ctx context.Context, in ListAllQueuesRequest) (ListAllQueuesResponse, error)
	GetWorkers(ctx context.Context, in GetWorkersRequest) (GetWorkersResponse, error)
	ScheduleJobs(ctx context.Context, in ScheduleJobsRequest) error
	RescheduleJob(ctx context.Context, in RescheduleJobRequest) error
	JobTypesForQueue(ct context.Context, queue jobs.QueueName) ([]jobs.JobType, error)
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
	ListAllQueuesRequest  struct{}
	ListAllQueuesResponse struct {
		QueueStats map[jobs.QueueName]jobs.QueueStats
	}
)

// todo how different to Queues and make that clear in the naming of the methods
// ListAllQueues returns all Queues.
func (app *JobsUsecase) ListAllQueues(ctx context.Context, in ListAllQueuesRequest) (ListAllQueuesResponse, error) {
	queues, err := app.repo.Queues(ctx)
	if err != nil {
		return ListAllQueuesResponse{}, fmt.Errorf("could not get queues: %w", err)
	}

	qWithStats := make(map[jobs.QueueName]jobs.QueueStats)
	for _, q := range queues {
		s, err := app.repo.QueueKPIs(ctx, q)
		if err != nil {
			return ListAllQueuesResponse{}, fmt.Errorf("could not get kpis for queue: %s: %w", q, err)
		}

		qWithStats[q] = queueKpiToStats(string(q), s)
	}

	return ListAllQueuesResponse{QueueStats: qWithStats}, nil
}

type (
	GetWorkersRequest  struct{}
	GetWorkersResponse struct {
		Pool []jobs.WorkerPool
	}
)

func (app *JobsUsecase) GetWorkers(ctx context.Context, in GetWorkersRequest) (GetWorkersResponse, error) {
	wp, _ := app.repo.WorkerPools(ctx)

	return GetWorkersResponse{Pool: wp}, nil
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

func (app *JobsUsecase) JobTypesForQueue(ctx context.Context, queue jobs.QueueName) ([]jobs.JobType, error) {
	if queue == jobs.DefaultQueueName { // todo move check to repo
		queue = ""
	}

	types, err := app.queries.JobTypes(ctx, string(queue))
	if err != nil {
		return nil, fmt.Errorf("could not get job types for queue: %s: %v", queue, err)
	}

	jobTypes := make([]jobs.JobType, len(types))
	for i, jt := range types {
		jobTypes[i] = jobs.JobType(jt)
	}

	return jobTypes, nil
}
