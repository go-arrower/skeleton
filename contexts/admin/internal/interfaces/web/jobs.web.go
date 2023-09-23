package web

import (
	"log/slog"

	"github.com/go-arrower/arrower/jobs"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

type JobsController struct {
	Repo   jobs.Repository
	Logger *slog.Logger
	Cmds   application.JobsCommandContainer
}
