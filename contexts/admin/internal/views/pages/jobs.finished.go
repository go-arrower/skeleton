package pages

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
	"github.com/labstack/echo/v4"
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

		fjobs[i].Payload = m.JobData
		//fjobs[i].Payload = prettyJobPayloadDataAsFormattedJSON(m)

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

func timeAgo(t time.Time) string { // todo move to shared views as general helper
	seconds := time.Since(t).Nanoseconds()

	switch seconds := time.Duration(seconds); {
	case seconds < time.Minute:
		return "now"
	case seconds < time.Hour:
		minutes := int(math.Round(float64(seconds / time.Minute)))
		if minutes == 1 {
			return fmt.Sprintf("%d minute ago", minutes)
		}

		return fmt.Sprintf("%d minutes ago", minutes)
	case seconds > time.Hour*24:
		days := int(math.Round(float64(seconds / (time.Hour * 24))))
		if days == 1 {
			return fmt.Sprintf("%d day ago", days)
		}

		return fmt.Sprintf("%d days ago", days)
	}

	return "unclear"
}
