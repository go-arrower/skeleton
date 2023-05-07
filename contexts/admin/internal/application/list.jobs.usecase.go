package application

import (
	"context"
	"time"

	"github.com/go-arrower/arrower/jobs"
)

type JobsCommandContainer struct {
	ListAllQueues func(ctx context.Context, in ListAllQueuesRequest) (ListAllQueuesResponse, error)
}

type (
	ListAllQueuesRequest  struct{}
	ListAllQueuesResponse struct {
		QueueStats map[string]QueueStats
	}
)

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
