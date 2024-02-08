package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/go-arrower/arrower/postgres"
)

var ErrJobLockedAlready = fmt.Errorf("%w: job might be processing already", postgres.ErrQueryFailed)

type (
	QueueStats struct { // todo return this from repo to prevent any mapping for trivial models like this
		QueueName            QueueName
		PendingJobs          int
		PendingJobsPerType   map[string]int
		FailedJobs           int
		ProcessedJobs        int
		AvailableWorkers     int
		PendingJobsErrorRate float64 // can be calculated: FailedJobs * 100 / PendingJobs
		AverageTimePerJob    time.Duration
		EstimateUntilEmpty   time.Duration // can be calculated
	}
)

type (
	PendingJob struct {
		CreatedAt  time.Time
		UpdatedAt  time.Time
		RunAt      time.Time
		RunAtFmt   string
		ID         string
		Type       string
		Queue      string
		Payload    string
		LastError  string
		ErrorCount int32
		Priority   int16
	}

	QueueKPIs struct {
		PendingJobsPerType map[string]int
		PendingJobs        int
		FailedJobs         int
		ProcessedJobs      int
		AvailableWorkers   int
		AverageTimePerJob  time.Duration
	}

	WorkerPool struct {
		LastSeen time.Time
		ID       string
		Queue    string
		Workers  int
		Version  string
		JobTypes []string
	}

	// Repository manages the data access to the underlying Jobs implementation.
	Repository interface {
		Queues(ctx context.Context) (QueueNames, error)
		PendingJobs(ctx context.Context, queue string) ([]PendingJob, error) // TODO use the QueueName type
		QueueKPIs(ctx context.Context, queue QueueName) (QueueKPIs, error)
		Delete(ctx context.Context, jobID string) error
		RunJobAt(ctx context.Context, jobID string, runAt time.Time) error
		WorkerPools(ctx context.Context) ([]WorkerPool, error)
	}
)
