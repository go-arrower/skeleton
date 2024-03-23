package application

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-arrower/arrower/app"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oklog/ulid/v2"
	"go.opentelemetry.io/otel/propagation"

	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

func NewScheduleJobsCommandHandler(queries *models.Queries) app.Command[ScheduleJobsCommand] {
	return &scheduleJobsCommandHandler{queries: queries}
}

type ScheduleJobsCommand struct {
	Queue    string
	JobType  string
	Priority int16
	Payload  string
	Count    int
	RunAt    time.Time
}

type scheduleJobsCommandHandler struct {
	queries *models.Queries
}

func (h *scheduleJobsCommandHandler) H(ctx context.Context, cmd ScheduleJobsCommand) error {
	carrier := propagation.MapCarrier{}
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})

	propagator.Inject(ctx, carrier)

	_, err := h.queries.ScheduleJobs(ctx, buildJobs(cmd, carrier))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func buildJobs(in ScheduleJobsCommand, carrier propagation.MapCarrier) []models.ScheduleJobsParams {
	jobs := make([]models.ScheduleJobsParams, in.Count)

	entropy := &ulid.LockedMonotonicReader{
		MonotonicReader: ulid.Monotonic(rand.Reader, 0),
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

type JobPayload struct { // todo reuse the one in the jobs package
	// Carrier contains the otel tracing information.
	Carrier propagation.MapCarrier `json:"carrier"`
	// JobData is the actual data as string instead of []byte,
	// so that it is readable more easily when assessing it via psql directly.
	JobData interface{} `json:"jobData"`
}
