package application

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-arrower/arrower/jobs"

	"github.com/go-arrower/skeleton/contexts/admin/internal/domain"
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
func ListAllQueues(repo jobs.Repository) func(context.Context, ListAllQueuesRequest) (ListAllQueuesResponse, error) {
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
		Jobs []jobs.PendingJob
		Kpis jobs.QueueKPIs
	}
)

// GetQueue returns a Queue.
func GetQueue(repo jobs.Repository) func(context.Context, GetQueueRequest) (GetQueueResponse, error) {
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
		Pool []jobs.WorkerPool
	}
)

func GetWorkers(repo jobs.Repository) func(context.Context, GetWorkersRequest) (GetWorkersResponse, error) {
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

func queueKpiToStats(queue string, kpis jobs.QueueKPIs) domain.QueueStats {
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
		Queue   string
		JobType string
		Count   int
	}
	ScheduleJobsResponse struct{}
)

func ScheduleJobs(jq jobs.Enqueuer) func(context.Context, ScheduleJobsRequest) error {
	return func(ctx context.Context, in ScheduleJobsRequest) error {
		err := jq.Enqueue(ctx, buildJobs(in.JobType, in.Count))
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return nil
	}
}

func buildJobs(jobType string, count int) []any {
	jobs := make([]any, count)

	for i := 0; i < count; i++ {
		switch jobType {
		case "SomeJob":
			jobs[i] = SomeJob{}
		case "LongRunningJob":
			jobs[i] = LongRunningJob{}
		}
	}

	return jobs
}

type (
	SomeJob        struct{}
	LongRunningJob struct{}
)

func ProcessSomeJob() func(context.Context, SomeJob) error {
	return func(ctx context.Context, job SomeJob) error {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second) //nolint:gosec,gomnd,lll // weak numbers are ok, it is wait time

		if rand.Intn(100) > 80 { //nolint:gosec,gomndworkers,gomnd
			return errors.New("some error") //nolint:goerr113
		}

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

func DeleteJob(repo jobs.Repository) func(context.Context, DeleteJobRequest) error {
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

func RescheduleJob(repo jobs.Repository) func(context.Context, RescheduleJobRequest) error {
	return func(ctx context.Context, in RescheduleJobRequest) error {
		err := repo.RunJobAt(ctx, in.JobID, time.Now())

		return fmt.Errorf("%w", err)
	}
}
