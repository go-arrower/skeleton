package application

import (
	"context"
	"fmt"

	"github.com/go-arrower/arrower/app"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

func NewJobTypesForQueueQueryHandler(queries *models.Queries) app.Query[JobTypesForQueueQuery, []jobs.JobType] {
	return &jobTypesForQueueQueryHandler{queries: queries}
}

type jobTypesForQueueQueryHandler struct {
	queries *models.Queries
}

type (
	JobTypesForQueueQuery struct {
		Queue jobs.QueueName
	}

	//JobTypesForQueueResponse struct{}
)

func (h jobTypesForQueueQueryHandler) H(ctx context.Context, query JobTypesForQueueQuery) ([]jobs.JobType, error) {
	queue := query.Queue
	if query.Queue == jobs.DefaultQueueName { // todo move check to repo
		queue = ""
	}

	types, err := h.queries.JobTypes(ctx, string(queue))
	if err != nil {
		return nil, fmt.Errorf("could not get job types for queue: %s: %v", queue, err)
	}

	jobTypes := make([]jobs.JobType, len(types))
	for i, jt := range types {
		jobTypes[i] = jobs.JobType(jt)
	}

	return jobTypes, nil
}
