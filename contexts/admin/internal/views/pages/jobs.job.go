package pages

import (
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

type historicJob struct {
	models.ArrowerGueJobsHistory
	PrettyPayload string
	CreatedAt     string
	EnqueuedAgo   string
	FinishedAgo   string
}

func ConvertFinishedJobsForShow(jobs []models.ArrowerGueJobsHistory) []historicJob {
	fjobs := make([]historicJob, len(jobs))

	for i, j := range jobs {
		fjobs[i] = historicJob{
			ArrowerGueJobsHistory: j,
			PrettyPayload:         prettyJobPayloadAsFormattedJSON(j.Args),
			CreatedAt:             formatAsDateOrTimeToday(j.CreatedAt.Time),
			EnqueuedAgo:           timeAgo(j.CreatedAt.Time),
			FinishedAgo:           timeAgo(j.FinishedAt.Time),
		}
	}

	return fjobs
}
