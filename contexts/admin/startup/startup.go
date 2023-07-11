package startup

import (
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/jobs/models"
	"github.com/go-arrower/arrower/postgres"
	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
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
				meterProvider,
				application.Logged(
					logger,
					application.ListAllQueues(repo),
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
		ScheduleJobs: application.ScheduleJobs(jq),
	}

	err := jq.RegisterJobFunc(
		application.LoggedU(logger,
			application.ProcessSomeJob())) // todo add meter & tracer
	logger.Info("REGISTER SOME JOB", slog.Any("err", err))

	_ = jq.RegisterJobFunc(
		application.LoggedU(logger,
			application.ProcessLongRunningJob())) // todo add meter & tracer

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
		jobs.GET("/schedule", cont.JobsSchedule())
		jobs.POST("/schedule", cont.JobsScheduleNew())
	}

	return nil
}
