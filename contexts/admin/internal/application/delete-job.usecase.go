package application

import (
	"context"
	"fmt"

	"github.com/go-arrower/arrower/app"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
)

func NewDeleteJobCommandHandler(repo jobs.Repository) app.Command[DeleteJobCommand] {
	return &deleteJobCommandHandler{repo: repo}
}

type deleteJobCommandHandler struct {
	repo jobs.Repository
}

type DeleteJobCommand struct {
	JobID string
}

func (h *deleteJobCommandHandler) H(ctx context.Context, cmd DeleteJobCommand) error {
	err := h.repo.Delete(ctx, cmd.JobID)

	return fmt.Errorf("%w", err)
}