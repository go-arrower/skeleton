package init

import (
	"context"
	"net/http"

	"github.com/go-arrower/skeleton/contexts/admin"

	"github.com/go-arrower/skeleton/shared/infrastructure"

	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/jobs/models"
	"github.com/go-arrower/arrower/mw"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
)

func NewAdminContext(di *infrastructure.Container) (*AdminContext, error) {
	di.AdminRouter.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/")
	})

	di.AdminRouter.GET("/routes", func(c echo.Context) error {
		return c.Render(http.StatusOK, "=>admin.routes", echo.Map{
			"Flashes": nil,
			"Routes":  di.WebRouter.Routes(),
		})
	})

	repo := jobs.NewPostgresJobsRepository(models.New(di.DB))

	container := application.JobsCommandContainer{
		ListAllQueues: mw.Traced(
			di.TraceProvider, mw.Metric(
				di.MeterProvider, mw.Logged(
					di.Logger.(*slog.Logger), application.ListAllQueues(repo),
				),
			),
		),
		GetQueue: mw.Traced(
			di.TraceProvider, mw.Metric(
				di.MeterProvider, mw.Logged(
					di.Logger.(*slog.Logger), application.GetQueue(repo),
				),
			),
		),
		GetWorkers: mw.Traced(
			di.TraceProvider, mw.Metric(
				di.MeterProvider, mw.Logged(
					di.Logger.(*slog.Logger), application.GetWorkers(repo),
				),
			),
		),
		ScheduleJobs: mw.TracedU(
			di.TraceProvider, mw.MetricU(
				di.MeterProvider, mw.LoggedU(
					di.Logger.(*slog.Logger), application.ScheduleJobs(di.DefaultQueue),
				),
			),
		),
		DeleteJob: mw.TracedU(
			di.TraceProvider, mw.MetricU(
				di.MeterProvider, mw.LoggedU(
					di.Logger.(*slog.Logger), application.DeleteJob(repo),
				),
			),
		),
		RescheduleJob: mw.TracedU(
			di.TraceProvider, mw.MetricU(
				di.MeterProvider, mw.LoggedU(
					di.Logger.(*slog.Logger), application.RescheduleJob(repo),
				),
			),
		),
	}

	_ = di.DefaultQueue.RegisterJobFunc(
		mw.TracedU(
			di.TraceProvider,
			mw.MetricU(
				di.MeterProvider,
				mw.LoggedU(
					di.Logger.(*slog.Logger),
					application.ProcessSomeJob(),
				),
			),
		),
	)
	_ = di.DefaultQueue.RegisterJobFunc(
		mw.TracedU(
			di.TraceProvider,
			mw.MetricU(
				di.MeterProvider,
				mw.LoggedU(
					di.Logger.(*slog.Logger),
					application.ProcessLongRunningJob(),
				),
			),
		),
	)

	cont := web.JobsController{
		Repo:   repo,
		Logger: di.Logger.(*slog.Logger),
		Cmds:   container,
	}

	{
		jobs := di.AdminRouter.Group("/jobs")
		jobs.GET("", cont.JobsHome())
		jobs.GET("/", cont.JobsHome())
		jobs.GET("/:queue", cont.JobsQueue())
		jobs.GET("/:queue/delete/:job_id", cont.DeleteJob())
		jobs.GET("/:queue/reschedule/:job_id", cont.RescheduleJob())
		jobs.GET("/workers", cont.JobsWorkers())
		jobs.GET("/settings", cont.JobsSettings())
		jobs.GET("/schedule", cont.JobsSchedule())
		jobs.POST("/schedule", cont.JobsScheduleNew())
	}

	return &AdminContext{}, nil
}

type AdminContext struct{}

func (c *AdminContext) SettingsAPI(ctx context.Context) (admin.SettingsAPI, error) {
	return application.NewMemorySettings(), nil
}

func (c *AdminContext) Shutdown(ctx context.Context) error {
	return nil
}
