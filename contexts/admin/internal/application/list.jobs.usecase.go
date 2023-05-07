package application

import (
	"context"
	"time"

	"github.com/go-arrower/arrower/jobs"
)

type JobsCommandContainer struct {
	ListAllQueues func(context.Context, ListAllQueuesRequest) (ListAllQueuesResponse, error)
	GetQueue      func(context.Context, GetQueueRequest) (GetQueueResponse, error)
	GetWorkers    func(context.Context, GetWorkersRequest) (GetWorkersResponse, error)
}

type (
	ListAllQueuesRequest  struct{}
	ListAllQueuesResponse struct {
		QueueStats map[string]QueueStats
	}
)

// ListAllQueues returns all Queues.
func ListAllQueues(repo jobs.Repository) func(ctx context.Context, in ListAllQueuesRequest) (ListAllQueuesResponse, error) {
	return func(ctx context.Context, in ListAllQueuesRequest) (ListAllQueuesResponse, error) {
		queues, _ := repo.Queues(ctx)

		qWithStats := make(map[string]QueueStats)
		for _, q := range queues {
			s, _ := repo.QueueKPIs(ctx, q)
			qWithStats[q] = queueKpiToStats(q, s)
		}

		// return ListAllQueuesResponse{}, errors.New("some-error")

		return ListAllQueuesResponse{QueueStats: qWithStats}, nil
	}
}

type (
	GetQueueRequest struct {
		QueueName string
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
		if queue == "Default" {
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

		return GetWorkersResponse{Pool: wp}, nil
	}
}

type QueueStats struct { // todo move to domain?
	QueueName            string
	PendingJobs          int
	PendingJobsPerType   map[string]int
	FailedJobs           int
	ProcessedJobs        int
	AvailableWorkers     int
	PendingJobsErrorRate float64 // can be calculated: FailedJobs * 100 / PendingJobs
	AverageTimePerJob    time.Duration
	EstimateUntilEmpty   time.Duration // can be calculated
}

func queueKpiToStats(queue string, kpis jobs.QueueKPIs) QueueStats {
	if queue == "" {
		queue = "Default"
	}

	var errorRate float64

	if kpis.FailedJobs != 0 {
		errorRate = float64(kpis.FailedJobs * 100 / kpis.PendingJobs)
	}

	var duration time.Duration = 0
	if kpis.AvailableWorkers != 0 {
		duration = time.Duration(kpis.PendingJobs/kpis.AvailableWorkers) * kpis.AverageTimePerJob
	}

	return QueueStats{
		QueueName:            queue,
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
