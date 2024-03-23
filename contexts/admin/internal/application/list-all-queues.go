package application

import (
	"context"
	"fmt"

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
func (h *listAllQueuesQueryHandler) H(ctx context.Context, query ListAllQueuesQuery) (ListAllQueuesResponse, error) {
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
