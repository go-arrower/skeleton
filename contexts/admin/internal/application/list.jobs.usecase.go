package application

import (
	crand "crypto/rand"
	"encoding/json"
	"strings"
	"time"

	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"

	"go.opentelemetry.io/otel/propagation"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid/v2"

	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

func queueKpiToStats(queue string, kpis jobs.QueueKPIs) jobs.QueueStats {
	var errorRate float64

	if kpis.FailedJobs != 0 {
		errorRate = float64(kpis.FailedJobs * 100 / kpis.PendingJobs)
	}

	var duration time.Duration
	if kpis.AvailableWorkers != 0 {
		duration = time.Duration(kpis.PendingJobs/kpis.AvailableWorkers) * kpis.AverageTimePerJob
	}

	return jobs.QueueStats{
		QueueName:            jobs.QueueName(queue),
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

type JobPayload struct { // todo reuse the one in the jobs package
	// Carrier contains the otel tracing information.
	Carrier propagation.MapCarrier `json:"carrier"`
	// JobData is the actual data as string instead of []byte,
	// so that it is readable more easily when assessing it via psql directly.
	JobData interface{} `json:"jobData"`
}

func buildJobs(in ScheduleJobsRequest, carrier propagation.MapCarrier) []models.ScheduleJobsParams {
	jobs := make([]models.ScheduleJobsParams, in.Count)

	entropy := &ulid.LockedMonotonicReader{
		MonotonicReader: ulid.Monotonic(crand.Reader, 0),
	}

	buf := map[string]interface{}{}
	_ = json.Unmarshal([]byte(strings.TrimSpace(in.Payload)), &buf)

	args, _ := json.Marshal(JobPayload{JobData: buf, Carrier: carrier})

	for i := 0; i < in.Count; i++ {
		jobID, _ := ulid.New(ulid.Now(), entropy)

		jobs[i] = models.ScheduleJobsParams{
			JobID:     jobID.String(),
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			Queue:     in.Queue,
			JobType:   in.JobType,
			Priority:  in.Priority,
			RunAt:     pgtype.Timestamptz{Time: in.RunAt, Valid: true},
			Args:      args,
		}

	}

	return jobs
}
