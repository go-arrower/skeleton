package application

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel/propagation"

	"github.com/go-arrower/arrower/alog"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid/v2"

	"github.com/go-arrower/skeleton/contexts/admin/internal/domain"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

const defaultQueueName = "Default"

type JobsCommandContainer struct {
	ListAllQueues func(context.Context, ListAllQueuesRequest) (ListAllQueuesResponse, error)
	GetQueue      func(context.Context, GetQueueRequest) (GetQueueResponse, error)
	GetWorkers    func(context.Context, GetWorkersRequest) (GetWorkersResponse, error)
	ScheduleJobs  func(context.Context, ScheduleJobsRequest) error
	DeleteJob     func(context.Context, DeleteJobRequest) error
	RescheduleJob func(context.Context, RescheduleJobRequest) error
}

type (
	ListAllQueuesRequest  struct{}
	ListAllQueuesResponse struct {
		QueueStats map[domain.QueueName]domain.QueueStats
	}
)

// ListAllQueues returns all Queues.
func ListAllQueues(repo domain.Repository) func(context.Context, ListAllQueuesRequest) (ListAllQueuesResponse, error) {
	return func(ctx context.Context, in ListAllQueuesRequest) (ListAllQueuesResponse, error) {
		queues, _ := repo.Queues(ctx) // todo repo needs to return type []QueueName
		qWithStats := make(map[domain.QueueName]domain.QueueStats)

		for _, q := range queues {
			s, _ := repo.QueueKPIs(ctx, q) // todo accept type QueueName
			qWithStats[domain.QueueName(q)] = queueKpiToStats(q, s)
		}

		// return ListAllQueuesResponse{}, errors.New("some-error")

		return ListAllQueuesResponse{QueueStats: qWithStats}, nil
	}
}

type (
	GetQueueRequest struct {
		QueueName string // todo type QueueName?
	}
	GetQueueResponse struct {
		Jobs []domain.PendingJob
		Kpis domain.QueueKPIs
	}
)

// GetQueue returns a Queue.
func GetQueue(repo domain.Repository) func(context.Context, GetQueueRequest) (GetQueueResponse, error) {
	return func(ctx context.Context, in GetQueueRequest) (GetQueueResponse, error) {
		queue := in.QueueName
		if queue == defaultQueueName {
			queue = ""
		}

		jobs, _ := repo.PendingJobs(ctx, queue)
		kpis, _ := repo.QueueKPIs(ctx, queue)

		return GetQueueResponse{
			Jobs: jobs,
			Kpis: kpis,
		}, nil
	}
}

type (
	GetWorkersRequest  struct{}
	GetWorkersResponse struct {
		Pool []domain.WorkerPool
	}
)

func GetWorkers(repo domain.Repository) func(context.Context, GetWorkersRequest) (GetWorkersResponse, error) {
	return func(ctx context.Context, in GetWorkersRequest) (GetWorkersResponse, error) {
		wp, _ := repo.WorkerPools(ctx)

		for i, _ := range wp {
			if wp[i].Queue == "" {
				wp[i].Queue = defaultQueueName
			}
		}

		return GetWorkersResponse{Pool: wp}, nil
	}
}

func queueKpiToStats(queue string, kpis domain.QueueKPIs) domain.QueueStats {
	if queue == "" {
		queue = defaultQueueName
	}

	var errorRate float64

	if kpis.FailedJobs != 0 {
		errorRate = float64(kpis.FailedJobs * 100 / kpis.PendingJobs)
	}

	var duration time.Duration
	if kpis.AvailableWorkers != 0 {
		duration = time.Duration(kpis.PendingJobs/kpis.AvailableWorkers) * kpis.AverageTimePerJob
	}

	return domain.QueueStats{
		QueueName:            domain.QueueName(queue),
		PendingJobs:          kpis.PendingJobs,
		PendingJobsPerType:   kpis.PendingJobsPerType,
		FailedJobs:           kpis.FailedJobs,
		ProcessedJobs:        kpis.ProcessedJobs,
		AvailableWorkers:     kpis.AvailableWorkers,
		PendingJobsErrorRate: errorRate,
		AverageTimePerJob:    kpis.AverageTimePerJob,
		EstimateUntilEmpty:   duration,
	}
}

type (
	ScheduleJobsRequest struct {
		Queue    string
		JobType  string
		Priority int16
		Payload  string
		Count    int
	}
	ScheduleJobsResponse struct{}
)

func ScheduleJobs(queries *models.Queries) func(context.Context, ScheduleJobsRequest) error {
	return func(ctx context.Context, in ScheduleJobsRequest) error {
		_, err := queries.ScheduleJobs(ctx, buildJobs(in))
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return nil
	}
}

func buildJobs(in ScheduleJobsRequest) []models.ScheduleJobsParams {
	jobs := make([]models.ScheduleJobsParams, in.Count)

	entropy := &ulid.LockedMonotonicReader{
		MonotonicReader: ulid.Monotonic(crand.Reader, 0),
	}

	type jobPayload struct { // todo reuse the one in the jobs package
		// Carrier contains the otel tracing information.
		Carrier propagation.MapCarrier `json:"carrier"`
		// JobData is the actual data as string instead of []byte,
		// so that it is readable more easily when assessing it via psql directly.
		JobData string `json:"jobData"`
	}

	args, _ := json.Marshal(jobPayload{JobData: in.Payload})

	for i := 0; i < in.Count; i++ {
		jobID, _ := ulid.New(ulid.Now(), entropy)

		jobs[i] = models.ScheduleJobsParams{
			JobID:     jobID.String(),
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			Queue:     in.Queue,
			JobType:   in.JobType,
			Priority:  in.Priority,
			RunAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
			Args:      args,
		}

	}

	return jobs
}

type (
	SomeJob        struct{}
	NamedJob       struct{ Name string }
	LongRunningJob struct{}
)

func ProcessSomeJob(logger alog.Logger) func(context.Context, SomeJob) error {
	return func(ctx context.Context, job SomeJob) error {
		logger.InfoContext(ctx, "LOG ASYNC SIMPLE JOB")
		//panic("SOME JOB PANICS")

		time.Sleep(time.Duration(rand.Intn(10)) * time.Second) //nolint:gosec,gomnd,lll // weak numbers are ok, it is wait time

		if rand.Intn(100) > 70 { //nolint:gosec,gomndworkers,gomnd
			return errors.New("some error") //nolint:goerr113
		}

		return nil
	}
}

func ProcessNamedJob(logger alog.Logger) func(context.Context, NamedJob) error {
	return func(ctx context.Context, job NamedJob) error {
		logger.InfoContext(ctx, "named job", slog.String("name", job.Name))

		return nil
	}
}

func ProcessLongRunningJob() func(context.Context, LongRunningJob) error {
	return func(ctx context.Context, job LongRunningJob) error {
		time.Sleep(time.Duration(rand.Intn(5)) * time.Minute) //nolint:gosec,gomnd // weak numbers are ok, it is wait time

		if rand.Intn(100) > 95 { //nolint:gosec,gomnd
			return errors.New("some error") //nolint:goerr113
		}

		return nil
	}
}

type (
	DeleteJobRequest struct {
		JobID string
	}
)

func DeleteJob(repo domain.Repository) func(context.Context, DeleteJobRequest) error {
	return func(ctx context.Context, in DeleteJobRequest) error {
		err := repo.Delete(ctx, in.JobID)

		return fmt.Errorf("%w", err)
	}
}

type (
	RescheduleJobRequest struct {
		JobID string
	}
)

func RescheduleJob(repo domain.Repository) func(context.Context, RescheduleJobRequest) error {
	return func(ctx context.Context, in RescheduleJobRequest) error {
		err := repo.RunJobAt(ctx, in.JobID, time.Now())

		return fmt.Errorf("%w", err)
	}
}
