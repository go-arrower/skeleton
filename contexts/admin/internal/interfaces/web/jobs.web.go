package web

import (
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"golang.org/x/exp/slog"
)

type JobsController struct {
	Repo   jobs.Repository
	Logger *slog.Logger
	Cmds   application.JobsCommandContainer
}
