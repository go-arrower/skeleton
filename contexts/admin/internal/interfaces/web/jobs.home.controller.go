package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-arrower/skeleton/shared/interfaces/web/views/pages"

	"github.com/go-arrower/skeleton/contexts/admin/internal/domain"

	"github.com/go-arrower/arrower/jobs"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

func (cont JobsController) JobsHome() func(c echo.Context) error {
	return func(c echo.Context) error {
		res, err := cont.Cmds.ListAllQueues(c.Request().Context(), application.ListAllQueuesRequest{})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return c.Render(http.StatusOK, "=>jobs.home", ListQueuesPage{Queues: res.QueueStats}) //nolint:wrapcheck
	}
}

func (cont JobsController) JobsQueue() func(c echo.Context) error {
	return func(c echo.Context) error {
		queue := c.Param("queue")

		res, err := cont.Cmds.GetQueue(c.Request().Context(), application.GetQueueRequest{
			QueueName: queue,
		})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		page := buildQueuePage(queue, res.Jobs, res.Kpis)

		return c.Render(http.StatusOK, "=>jobs.queue", page) //nolint:wrapcheck
	}
}

func (cont JobsController) JobsWorkers() func(c echo.Context) error {
	return func(c echo.Context) error {
		res, err := cont.Cmds.GetWorkers(c.Request().Context(), application.GetWorkersRequest{})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return c.Render(http.StatusOK, "=>jobs.workers", presentWorkers(res.Pool)) //nolint:wrapcheck
	}
}

func (cont JobsController) JobsSchedule() func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "=>jobs.schedule", nil) //nolint:wrapcheck
	}
}

func (cont JobsController) JobsScheduleNew() func(c echo.Context) error {
	return func(c echo.Context) error {
		q := c.FormValue("queue")
		jt := c.FormValue("job_type")
		num := c.FormValue("count")

		count, err := strconv.Atoi(num)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = cont.Cmds.ScheduleJobs(c.Request().Context(), q, jt, count)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/jobs/%s", q))
	}
}

func presentWorkers(pool []jobs.WorkerPool) []pages.JobWorker {
	p := make([]pages.JobWorker, len(pool))

	for i, _ := range pool {
		p[i].ID = pool[i].ID
		p[i].Queue = pool[i].Queue
		p[i].Workers = pool[i].Workers
		p[i].LastSeenAt = pool[i].LastSeen.Format(time.TimeOnly)

		p[i].LastSeenAtColour = "text-green-600"
		if time.Now().Sub(pool[i].LastSeen)/time.Second > 30 {
			p[i].LastSeenAtColour = "text-orange-600"
		}

		p[i].NotSeenSince = notSeenSinceTimeString(pool[i].LastSeen)
	}

	return p
}

func notSeenSinceTimeString(t time.Time) string {
	seconds := time.Now().Sub(t).Seconds()

	if seconds > 60 {
		return fmt.Sprintf("%d m %d sec", int(seconds/60), int(seconds)%60)
	}

	return fmt.Sprintf("%d sec", int(seconds))
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
		Queues map[domain.QueueName]domain.QueueStats
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

		if jobs[i].Queue == "" {
			jobs[i].Queue = "Default"
		}
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

func (cont JobsController) DeleteJob() func(c echo.Context) error {
	return func(c echo.Context) error {
		q := c.Param("queue")
		jobID := c.Param("job_id")

		_, _ = cont.Cmds.DeleteJob(c.Request().Context(), application.DeleteJobRequest{JobID: jobID})

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/jobs/%s", q))
	}
}
