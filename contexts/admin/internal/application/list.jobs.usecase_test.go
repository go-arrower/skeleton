package application_test

import (
	"context"
	"testing"

	"github.com/go-arrower/arrower/jobs"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
)

func TestScheduleJobs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	jq := jobs.NewTestingJobs()
	assert := jq.Assert(t)

	scheduleJobs := application.ScheduleJobs(nil)

	_ = scheduleJobs(ctx, application.ScheduleJobsRequest{
		Queue:   "",
		JobType: "SomeJob",
		Count:   1,
	})

	assert.NotEmpty()
	assert.Queued(application.SomeJob{}, 1)

	_ = scheduleJobs(ctx, application.ScheduleJobsRequest{
		Queue:   "",
		JobType: "LongRunningJob",
		Count:   1,
	})

	assert.NotEmpty()
	assert.Queued(application.SomeJob{}, 1)
	assert.QueuedTotal(2)
}
