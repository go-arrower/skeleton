package startup

import (
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/jobs/models"
	"github.com/go-arrower/arrower/postgres"
	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

func Init(e *echo.Echo, pg *postgres.Handler, logger *slog.Logger) error {

	admin := e.Group("/admin")

	repo := jobs.NewPostgresJobsRepository(models.New(pg.PGx))

	container := application.JobsCommandContainer{
		ListAllQueues: application.Logged(logger, application.ListAllQueues(repo)),
	}

	cont := web.JobsController{
		Repo:   repo,
		Logger: logger,
		Cmds:   container,
	}

	{
		jobs := admin.Group("/jobs")
		jobs.GET("", cont.JobsHome())
		jobs.GET("/", cont.JobsHome())
		jobs.GET("/:queue", cont.JobsQueue())
		jobs.GET("/workers", cont.JobsWorkers())
	}

	return nil
}
