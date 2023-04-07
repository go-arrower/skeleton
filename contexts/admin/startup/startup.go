package startup

import (
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/jobs/models"
	"github.com/go-arrower/arrower/postgres"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
	"github.com/labstack/echo/v4"
)

func Init(e *echo.Echo, pg *postgres.Handler) error {

	admin := e.Group("/admin")

	cont := web.JobsController{
		Repo: jobs.NewPostgresJobsRepository(models.New(pg.PGx)),
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
