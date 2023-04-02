package web

import (
	"net/http"
	"time"

	"github.com/go-arrower/arrower/jobs"

	"github.com/labstack/echo/v4"
)

func (cont JobsController) JobsHome() func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "jobs.home", nil) //nolint:wrapcheck
	}
}

func (cont JobsController) JobsQueue() func(c echo.Context) error {
	return func(c echo.Context) error {
		queue := c.Param("queue")
		if queue == "Default" {
			queue = ""
		}

		jobs, _ := cont.Repo.PendingJobs(c.Request().Context(), queue)
		kpis, _ := cont.Repo.QueueKPIs(c.Request().Context(), queue)

		page := buildQueuePage(queue, jobs, kpis)

		return c.Render(http.StatusOK, "jobs.queue", page) //nolint:wrapcheck
	}
}

type (
	QueueStats struct {
		QueueName          string
		PendingJobs        int
		PendingJobsPerType map[string]int
		FailedJobs         int
		AvailableWorkers   int
		ErrorRate          float64 // can be calculated: FailedJobs * 100 / PendingJobs
		AverageTimePerJob  time.Duration
		EstimateUntilEmpty time.Duration // can be calculated
	}

	QueuePage struct {
		QueueName string
		Stats     QueueStats
		Jobs      []jobs.PendingJob
	}
)

func buildQueuePage(queue string, jobs []jobs.PendingJob, kpis jobs.QueueKPIs) QueuePage {
	var errorRate float64

	if kpis.FailedJobs != 0 {
		errorRate = float64(kpis.FailedJobs * 100 / kpis.PendingJobs)
	}

	return QueuePage{
		QueueName: queue, // if "" => Default
		Stats: QueueStats{
			QueueName:          queue,
			PendingJobs:        kpis.PendingJobs,
			PendingJobsPerType: kpis.PendingJobsPerType,
			FailedJobs:         kpis.FailedJobs,
			AvailableWorkers:   kpis.AvailableWorkers,
			ErrorRate:          errorRate,
			AverageTimePerJob:  kpis.AverageTimePerJob,
			EstimateUntilEmpty: time.Duration(kpis.PendingJobs) * kpis.AverageTimePerJob,
		},

		Jobs: jobs,
	}
}
