package startup

import (
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/jobs/models"
	"github.com/go-arrower/arrower/postgres"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
)

func Init(
	logger *slog.Logger,
	traceProvider trace.TracerProvider,
	meterProvider metric.MeterProvider,
	e *echo.Echo,
	pg *postgres.Handler,
	jq jobs.Queue,
) error {
	admin := e.Group("/admin")

	repo := jobs.NewPostgresJobsRepository(models.New(pg.PGx))

	container := application.JobsCommandContainer{
		ListAllQueues: application.Traced(
			traceProvider, application.Metric(
				meterProvider, application.Logged(
					logger, application.ListAllQueues(repo),
				),
			),
		),
		GetQueue: application.Traced(
			traceProvider, application.Metric(
				meterProvider, application.Logged(
					logger, application.GetQueue(repo),
				),
			),
		),
		GetWorkers: application.Traced(
			traceProvider, application.Metric(
				meterProvider, application.Logged(
					logger, application.GetWorkers(repo),
				),
			),
		),
		ScheduleJobs: application.TracedU(
			traceProvider, application.MetricU(
				meterProvider, application.LoggedU(
					logger, application.ScheduleJobs(jq),
				),
			),
		),
		DeleteJob: application.TracedU(
			traceProvider, application.MetricU(
				meterProvider, application.LoggedU(
					logger, application.DeleteJob(repo),
				),
			),
		),
		RescheduleJob: application.TracedU(
			traceProvider, application.MetricU(
				meterProvider, application.LoggedU(
					logger, application.RescheduleJob(repo),
				),
			),
		),
	}

	_ = jq.RegisterJobFunc(
		application.TracedU(
			traceProvider,
			application.MetricU(
				meterProvider,
				application.LoggedU(
					logger,
					application.ProcessSomeJob(),
				),
			),
		),
	)
	_ = jq.RegisterJobFunc(
		application.TracedU(
			traceProvider,
			application.MetricU(
				meterProvider,
				application.LoggedU(
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
		jobs := admin.Group("/jobs")
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
