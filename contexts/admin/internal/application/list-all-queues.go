package application

import (
	"context"
	"fmt"
	"time"

	"github.com/go-arrower/arrower/app"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
)

func NewListAllQueuesQueryHandler(repo jobs.Repository) app.Query[ListAllQueuesQuery, ListAllQueuesResponse] {
	return &listAllQueuesQueryHandler{repo: repo}
}

type listAllQueuesQueryHandler struct {
	repo jobs.Repository
}

type (
	ListAllQueuesQuery struct{}

	ListAllQueuesResponse struct {
		QueueStats map[jobs.QueueName]jobs.QueueStats
	}
)

// todo how different to Queues and make that clear in the naming of the methods
// ListAllQueues returns all Queues.
func (h *listAllQueuesQueryHandler) H(ctx context.Context, _ ListAllQueuesQuery) (ListAllQueuesResponse, error) {
	queues, err := h.repo.Queues(ctx)
	if err != nil {
		return ListAllQueuesResponse{}, fmt.Errorf("could not get queues: %w", err)
	}

	qWithStats := make(map[jobs.QueueName]jobs.QueueStats)

	for _, q := range queues {
		s, err := h.repo.QueueKPIs(ctx, q)
		if err != nil {
			return ListAllQueuesResponse{}, fmt.Errorf("could not get kpis for queue: %s: %w", q, err)
		}

		qWithStats[q] = queueKpiToStats(string(q), s)
	}

	return ListAllQueuesResponse{QueueStats: qWithStats}, nil
}

func queueKpiToStats(queue string, kpis jobs.QueueKPIs) jobs.QueueStats {
	var errorRate float64

	if kpis.FailedJobs != 0 {
		errorRate = float64(kpis.FailedJobs * 100 / kpis.PendingJobs)
	}

	var duration time.Duration
	if kpis.AvailableWorkers != 0 {
		duration = time.Duration(kpis.PendingJobs/kpis.AvailableWorkers) * kpis.AverageTimePerJob
	}

	return jobs.QueueStats{
		QueueName:            jobs.QueueName(queue),
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
