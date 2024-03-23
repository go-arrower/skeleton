package application

import (
	"context"
	"fmt"

	"github.com/go-arrower/arrower/app"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
)

func NewGetWorkersQueryHandler(repo jobs.Repository) app.Query[GetWorkersQuery, GetWorkersResponse] {
	return &getWorkersQueryHandler{repo: repo}
}

type getWorkersQueryHandler struct {
	repo jobs.Repository
}

type (
	GetWorkersQuery    struct{}
	GetWorkersResponse struct {
		Pool []jobs.WorkerPool
	}
)

func (h *getWorkersQueryHandler) H(ctx context.Context, query GetWorkersQuery) (GetWorkersResponse, error) {
	wp, err := h.repo.WorkerPools(ctx)
	if err != nil {
		return GetWorkersResponse{}, fmt.Errorf("could not get workers: %w", err)
	}

	return GetWorkersResponse{Pool: wp}, nil
}
