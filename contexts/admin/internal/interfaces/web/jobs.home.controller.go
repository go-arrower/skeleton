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

		page := buildQueuePage(queue, jobs)

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
		ErrorRate          float32 // can be calculated: FailedJobs * 100 / PendingJobs
		AverageTimePerJob  time.Duration
		EstimateUntilEmpty time.Duration // can be calculated
	}

	QueuePage struct {
		QueueName string
		Stats     QueueStats
		Jobs      []jobs.PendingJob
	}
)

func buildQueuePage(queue string, jobs []jobs.PendingJob) QueuePage {
	return QueuePage{
		QueueName: queue,
		Stats: QueueStats{
			QueueName:   queue,
			PendingJobs: len(jobs),
			PendingJobsPerType: map[string]int{
				"some_type":       1,
				"register":        2,
				"clean_something": 3,
				"domain_job":      4,
			},
			FailedJobs:         3,
			AvailableWorkers:   10,
			ErrorRate:          0.16,
			AverageTimePerJob:  1500 * time.Millisecond,
			EstimateUntilEmpty: 2754000 * time.Millisecond,
		},

		Jobs: jobs,
	}
}

var exampleJobs = []map[string]string{
	{
		"ID":         "gaht6e",
		"Type":       "register_email",
		"Priority":   "0",
		"Payload":    "{data:[1,2,3]}",
		"RunAt":      "13:37",
		"LastError":  "no error message",
		"ErrorCount": "0",
	},
	{
		"ID":         "lz8abg",
		"Type":       "register_email",
		"Priority":   "0",
		"Payload":    "{data:[1,2,3]}",
		"RunAt":      "13:37",
		"LastError":  "no error message",
		"ErrorCount": "0",
	},
	{
		"ID":         "0jgzabg",
		"Type":       "welcome_job",
		"Priority":   "0",
		"Payload":    "{data:[1,2,3]}",
		"RunAt":      "13:37",
		"LastError":  "no error message",
		"ErrorCount": "0",
	},
}
