package pages

import (
	"bytes"
	"encoding/json"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

type historicJob struct {
	models.ArrowerGueJobsHistory
	PrettyPayload string
}

func ConvertFinishedJobs(jobs []models.ArrowerGueJobsHistory) []historicJob {
	fjobs := make([]historicJob, len(jobs))

	for i, j := range jobs {
		fjobs[i] = historicJob{
			ArrowerGueJobsHistory: j,
			PrettyPayload:         prettyJobPayloadAsFormattedJSON(j.Args),
		}
	}

	return fjobs
}

func prettyJobPayloadAsFormattedJSON(p []byte) string {
	return prettyString(string(p))
}

func prettyJobPayloadDataAsFormattedJSON(payload application.JobPayload) string {
	return prettyString(payload.JobData)
}

func prettyString(str string) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "  "); err != nil {
		return ""
	}

	return prettyJSON.String()
}
