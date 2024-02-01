package init

import (
	"context"
	"log/slog"
	"net/http"
	"sort"

	models3 "github.com/go-arrower/arrower/alog/models"
	"github.com/go-arrower/arrower/mw"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository"
	models2 "github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
	"github.com/go-arrower/skeleton/shared/infrastructure"
	web2 "github.com/go-arrower/skeleton/shared/interfaces/web"
)

func NewAdminContext(di *infrastructure.Container) (*AdminContext, error) {
	di.AdminRouter.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/")
	})

	di.AdminRouter.GET("/routes", func(c echo.Context) error {
		routes := di.WebRouter.Routes()

		// sort routes by path and then by method
		sort.Slice(routes, func(i, j int) bool {
			if routes[i].Path < routes[j].Path {
				return true
			}

			if routes[i].Path == routes[j].Path {
				return routes[i].Method < routes[j].Method
			}

			return false
		})

		return c.Render(http.StatusOK, "admin.routes", echo.Map{
			"Flashes": nil,
			"Routes":  routes,
		})
	})

	repo := repository.NewTracedJobsRepository(repository.NewPostgresJobsRepository(di.PGx))

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
					di.Logger.(*slog.Logger), application.ScheduleJobs(models2.New(di.PGx)),
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
					application.ProcessSomeJob(di.Logger),
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
					application.ProcessNamedJob(di.Logger),
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

	{
		settingsCont := web.NewSettingsController(di.AdminRouter)
		settingsCont.List()
	}

	cont := web.NewJobsController(di.Logger, repo, web2.NewDefaultPresenter(di.Settings), application.NewJobsApplication(di.PGx))
	cont.Cmds = container
	cont.Queries = models2.New(di.PGx)

	{
		jobs := di.AdminRouter.Group("/jobs")
		jobs.GET("", cont.ListQueues())
		jobs.GET("/", cont.ListQueues())
		jobs.GET("/data/pending", cont.PendingJobsPieChartData())                // todo better htmx fruednly data URL
		jobs.GET("/data/processed/:interval", cont.ProcessedJobsLineChartData()) // todo better htmx fruednly data URL
		jobs.GET("/:queue", cont.ShowQueue())
		jobs.GET("/:queue/delete/:job_id", cont.DeleteJob())
		jobs.GET("/:queue/reschedule/:job_id", cont.RescheduleJob())
		jobs.GET("/schedule", cont.CreateJobs()).Name = "admin.jobs.schedule"
		jobs.POST("/schedule", cont.ScheduleJobs()).Name = "admin.jobs.new"
		jobs.GET("/jobTypes", cont.ShowJobTypes())
		jobs.GET("/payloads", cont.PayloadExamples())
		jobs.GET("/workers", cont.ListWorkers())
		jobs.GET("/maintenance", cont.ShowMaintenance()).Name = "admin.jobs.maintenance"
		jobs.POST("/vacuum/:table", cont.VacuumJobTables())
		jobs.POST("/history", cont.DeleteHistory())
		jobs.POST("/history/prune", cont.PruneHistory())
		jobs.GET("/history/size/", cont.EstimateHistorySize())
		jobs.GET("/history/payload/size/", cont.EstimateHistoryPayloadSize())
	}

	adminContext := &AdminContext{}

	jc := web.NewLogsController(di.Logger, models3.New(di.PGx), di.AdminRouter.Group("/logs"), web2.NewDefaultPresenter(di.Settings))
	jc.ShowLogs()

	return adminContext, nil
}

type AdminContext struct{}

func (c *AdminContext) Shutdown(_ context.Context) error {
	return nil
}
