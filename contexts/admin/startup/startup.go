package startup

import (
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
	"github.com/labstack/echo/v4"
)

func Init(e *echo.Echo) error {

	admin := e.Group("/admin")

	cont := web.JobsController{}

	{
		jobs := admin.Group("/jobs")
		jobs.GET("", cont.JobsHome())
		jobs.GET("/", cont.JobsHome())
		jobs.GET("/queue", cont.JobsQueue())
	}

	return nil
}
