package init

import (
	"context"
	"log/slog"
	"net/http"
	"sort"

	models3 "github.com/go-arrower/arrower/alog/models"
	"github.com/go-arrower/arrower/mw"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/admin"
	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/models"
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

		return c.Render(http.StatusOK, "=>admin.routes", echo.Map{
			"Flashes": nil,
			"Routes":  routes,
		})
	})

	repo := domain.NewTracedJobsRepository(domain.NewPostgresJobsRepository(models.New(di.DB)))
	settingsRepo := repository.NewSettingsMemoryRepository()

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
					di.Logger.(*slog.Logger), application.ScheduleJobs(models2.New(di.DB)),
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
		settingsCont := web.NewSettingsController(di.AdminRouter, settingsRepo)
		settingsCont.List()
		settingsCont.Update()
	}

	cont := web.NewJobsController(di.Logger, repo, web2.NewDefaultPresenter(application.NewSettingsApp(settingsRepo)))
	cont.Cmds = container
	cont.Queries = models2.New(di.DB)
	cont.DB = di.DB

	{
		jobs := di.AdminRouter.Group("/jobs")
		jobs.GET("", cont.ListQueues())
		jobs.GET("/", cont.ListQueues())
		jobs.GET("/data/pending", cont.PendingJobsPieChartData())                // todo better htmx fruednly data URL
		jobs.GET("/data/processed/:interval", cont.ProcessedJobsLineChartData()) // todo better htmx fruednly data URL
		jobs.GET("/:queue", cont.ShowQueue())
		jobs.GET("/:queue/delete/:job_id", cont.DeleteJob())
		jobs.GET("/:queue/reschedule/:job_id", cont.RescheduleJob())
		jobs.GET("/workers", cont.ListWorkers())
		jobs.GET("/settings", cont.ShowSettings())
		jobs.POST("/vacuum/:table", cont.VacuumJobTables())
		jobs.GET("/schedule", cont.CreateJobs())
		jobs.GET("/jobTypes", cont.ShowJobTypes())
		jobs.POST("/schedule", cont.ScheduleJobs())
	}

	adminContext := &AdminContext{
		settingsRepo: settingsRepo,
	}

	api, _ := adminContext.SettingsAPI(context.Background())
	jc := web.NewLogsController(di.Logger, models3.New(di.DB), di.AdminRouter.Group("/logs"), web2.NewDefaultPresenter(api))
	jc.ShowLogs()

	return adminContext, nil
}

type AdminContext struct {
	settingsRepo domain.SettingRepository
}

func (c *AdminContext) SettingsAPI(_ context.Context) (admin.SettingsAPI, error) {
	return application.NewSettingsApp(c.settingsRepo), nil
}

func (c *AdminContext) Shutdown(_ context.Context) error {
	return nil
}

func NewMemorySettingsAPI() admin.SettingsAPI {
	return application.NewSettingsApp(repository.NewSettingsMemoryRepository())
}
