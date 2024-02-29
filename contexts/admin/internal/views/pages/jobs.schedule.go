package pages

import (
	"bytes"
	"encoding/json"

	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
)

func PresentJobsExamplePayloads(queue, jobType string, payloads [][]byte) echo.Map {
	prettyPayloads := make([]string, len(payloads))

	for i, p := range payloads {
		var jobPayload application.JobPayload
		_ = json.Unmarshal(p, &jobPayload)

		prettyPayloads[i] = prettyJobPayloadDataAsFormattedJSON(jobPayload)
	}

	if queue == "" {
		queue = string(jobs.DefaultQueueName)
	}

	return echo.Map{
		"Queue":    queue,
		"JobType":  jobType,
		"Payloads": prettyPayloads,
	}
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
