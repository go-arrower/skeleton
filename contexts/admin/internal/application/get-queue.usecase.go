package application

import (
	"context"
	"fmt"

	"github.com/go-arrower/arrower/app"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
)

func NewGetQueueQueryHandler(repo jobs.Repository) app.Query[GetQueueQuery, GetQueueResponse] {
	return &getQueueQueryHandler{repo: repo}
}

type getQueueQueryHandler struct {
	repo jobs.Repository
}

type (
	GetQueueQuery struct {
		QueueName jobs.QueueName
	}
	GetQueueResponse struct {
		Jobs []jobs.PendingJob
		Kpis jobs.QueueKPIs
	}
)

func (h *getQueueQueryHandler) H(ctx context.Context, query GetQueueQuery) (GetQueueResponse, error) {
	kpis, err := h.repo.QueueKPIs(ctx, query.QueueName)
	if err != nil {
		return GetQueueResponse{}, fmt.Errorf("could not get queue kpis: %w", err)
	}

	jobs, err := h.repo.PendingJobs(ctx, query.QueueName)
	if err != nil {
		return GetQueueResponse{}, fmt.Errorf("could not get pending jobs: %w", err)
	}

	return GetQueueResponse{
		Jobs: jobs,
		Kpis: kpis,
	}, nil
}
