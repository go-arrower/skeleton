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
	jq := jobs.NewInMemoryJobs()
	assert := jq.Assert(t)

	scheduleJobs := application.ScheduleJobs(jq)

	_ = scheduleJobs(ctx, "", "SomeJob", 1)
	assert.NotEmpty()
	assert.Queued(application.SomeJob{}, 1)

	_ = scheduleJobs(ctx, "", "LongRunningJob", 1)
	assert.NotEmpty()
	assert.Queued(application.SomeJob{}, 1)
	assert.QueuedTotal(2)
}
