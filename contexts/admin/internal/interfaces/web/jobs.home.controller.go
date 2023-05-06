package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"

	"github.com/go-arrower/arrower/jobs"

	"github.com/labstack/echo/v4"
)

func (cont JobsController) JobsHome() func(c echo.Context) error {
	return func(c echo.Context) error {
		res, _ := cont.Cmds.ListAllQueues(c.Request().Context(), application.ListAllQueuesRequest{})

		return c.Render(http.StatusOK, "=>jobs.home", ListQueuesPage{Queues: res.QueueStats}) //nolint:wrapcheck
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

		return c.Render(http.StatusOK, "=>jobs.queue", page) //nolint:wrapcheck
	}
}

func (cont JobsController) JobsWorkers() func(c echo.Context) error {
	return func(c echo.Context) error {
		wp, _ := cont.Repo.WorkerPools(c.Request().Context())

		return c.Render(http.StatusOK, "=>jobs.workers", wp) //nolint:wrapcheck
	}
}

type (
	QueueStats struct {
		QueueName            string
		PendingJobs          int
		PendingJobsPerType   map[string]int
		FailedJobs           int
		ProcessedJobs        int
		AvailableWorkers     int
		PendingJobsErrorRate float64 // can be calculated: FailedJobs * 100 / PendingJobs
		AverageTimePerJob    time.Duration
		EstimateUntilEmpty   time.Duration // can be calculated
	}

	ListQueuesPage struct {
		Queues map[string]application.QueueStats
	}

	QueuePage struct {
		QueueName string
		Stats     QueueStats
		Jobs      []jobs.PendingJob
	}
)

func buildQueuePage(queue string, jobs []jobs.PendingJob, kpis jobs.QueueKPIs) QueuePage {
	if queue == "" {
		queue = "Default"
	}

	jobs = prettyFormatPayload(jobs)

	return QueuePage{
		QueueName: queue,
		Stats:     queueKpiToStats(queue, kpis),

		Jobs: jobs,
	}
}

func prettyFormatPayload(jobs []jobs.PendingJob) []jobs.PendingJob {
	for i := 0; i < len(jobs); i++ {
		var m map[string]interface{}
		_ = json.Unmarshal([]byte(jobs[i].Payload), &m)

		data, _ := json.MarshalIndent(m, "", "  ")
		jobs[i].Payload = string(data)
	}

	return jobs
}

func queueKpiToStats(queue string, kpis jobs.QueueKPIs) QueueStats {
	if queue == "" {
		queue = "Default"
	}

	var errorRate float64

	if kpis.FailedJobs != 0 {
		errorRate = float64(kpis.FailedJobs * 100 / kpis.PendingJobs)
	}

	var duration time.Duration = 0
	if kpis.AvailableWorkers != 0 {
		duration = time.Duration(kpis.PendingJobs/kpis.AvailableWorkers) * kpis.AverageTimePerJob
	}

	return QueueStats{
		QueueName:            queue,
		PendingJobs:          kpis.PendingJobs,
		PendingJobsPerType:   kpis.PendingJobsPerType,
		FailedJobs:           kpis.FailedJobs,
		ProcessedJobs:        kpis.ProcessedJobs,
		AvailableWorkers:     kpis.AvailableWorkers,
		PendingJobsErrorRate: errorRate,
		AverageTimePerJob:    kpis.AverageTimePerJob,
		EstimateUntilEmpty:   duration,
	}
}
