package domain

import "time"

type (
	QueueName  string
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
