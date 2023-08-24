package startup

import (
	"net/http"

	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/jobs/models"
	"github.com/go-arrower/arrower/mw"
	"github.com/go-arrower/arrower/postgres"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
)

func Init(logger *slog.Logger, traceProvider trace.TracerProvider, meterProvider metric.MeterProvider, e *echo.Group, pg *postgres.Handler, jq jobs.Queue) error {
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/")
	})

	repo := jobs.NewPostgresJobsRepository(models.New(pg.PGx))

	container := application.JobsCommandContainer{
		ListAllQueues: mw.Traced(
			traceProvider, mw.Metric(
				meterProvider, mw.Logged(
					logger, application.ListAllQueues(repo),
				),
			),
		),
		GetQueue: mw.Traced(
			traceProvider, mw.Metric(
				meterProvider, mw.Logged(
					logger, application.GetQueue(repo),
				),
			),
		),
		GetWorkers: mw.Traced(
			traceProvider, mw.Metric(
				meterProvider, mw.Logged(
					logger, application.GetWorkers(repo),
				),
			),
		),
		ScheduleJobs: mw.TracedU(
			traceProvider, mw.MetricU(
				meterProvider, mw.LoggedU(
					logger, application.ScheduleJobs(jq),
				),
			),
		),
		DeleteJob: mw.TracedU(
			traceProvider, mw.MetricU(
				meterProvider, mw.LoggedU(
					logger, application.DeleteJob(repo),
				),
			),
		),
		RescheduleJob: mw.TracedU(
			traceProvider, mw.MetricU(
				meterProvider, mw.LoggedU(
					logger, application.RescheduleJob(repo),
				),
			),
		),
	}

	_ = jq.RegisterJobFunc(
		mw.TracedU(
			traceProvider,
			mw.MetricU(
				meterProvider,
				mw.LoggedU(
					logger,
					application.ProcessSomeJob(),
				),
			),
		),
	)
	_ = jq.RegisterJobFunc(
		mw.TracedU(
			traceProvider,
			mw.MetricU(
				meterProvider,
				mw.LoggedU(
					logger,
					application.ProcessLongRunningJob(),
				),
			),
		),
	)

	cont := web.JobsController{
		Repo:   repo,
		Logger: logger,
		Cmds:   container,
	}

	{
		jobs := e.Group("/jobs")
		jobs.GET("", cont.JobsHome())
		jobs.GET("/", cont.JobsHome())
		jobs.GET("/:queue", cont.JobsQueue())
		jobs.GET("/:queue/delete/:job_id", cont.DeleteJob())
		jobs.GET("/:queue/reschedule/:job_id", cont.RescheduleJob())
		jobs.GET("/workers", cont.JobsWorkers())
		jobs.GET("/schedule", cont.JobsSchedule())
		jobs.POST("/schedule", cont.JobsScheduleNew())
	}

	return nil
}
