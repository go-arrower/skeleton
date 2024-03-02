package pages

import (
	"encoding/json"

	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
)

func NewFinishedJobs(jobs []jobs.PendingJob) echo.Map {
	type finishedJob struct {
		EnqueuedAtFmt string
		FinishedAtFmt string
		ID            string
		Type          string
		Queue         string
		Payload       string
	}

	fjobs := make([]finishedJob, len(jobs))
	for i := 0; i < len(jobs); i++ {
		var m application.JobPayload
		_ = json.Unmarshal([]byte(jobs[i].Payload), &m)

		//fjobs[i].Payload = m.JobData.(string)
		fjobs[i].Payload = prettyJobPayloadDataAsFormattedJSON(m)

		fjobs[i].EnqueuedAtFmt = timeAgo(jobs[i].CreatedAt)
		fjobs[i].FinishedAtFmt = timeAgo(jobs[i].UpdatedAt) // todo use finished at
		fjobs[i].ID = jobs[i].ID
		fjobs[i].Type = jobs[i].Type
		fjobs[i].Queue = jobs[i].Queue
	}

	return echo.Map{
		"Jobs": fjobs,
	}
}
