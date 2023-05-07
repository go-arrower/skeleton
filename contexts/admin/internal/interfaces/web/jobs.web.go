package web

import (
	"github.com/go-arrower/arrower/jobs"
	"golang.org/x/exp/slog"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

type JobsController struct {
	Repo   jobs.Repository
	Logger *slog.Logger
	Cmds   application.JobsCommandContainer
}
