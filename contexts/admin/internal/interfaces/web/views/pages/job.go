package pages

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
)

func NewJob(jobs []models.ArrowerGueJobsHistory) Job {
	showAction := false
	if !jobs[0].Success { // todo check if jobs is empty
		showAction = true
	}

	return Job{
		Jobs:        jobs,
		ShowActions: showAction,
	}
}

type (
	Jobs []models.ArrowerGueJobsHistory
	Job  struct {
		Jobs // either way works
		// Jobs []models.ArrowerGueJobsHistory

		ShowActions bool
	}
)

// TimelineTime could reuse a shared timeFMT function, so it is coherent across pages
func (j Job) TimelineTime(t pgtype.Timestamptz) string {
	isToday := t.Time.Format("2006.01.02") == time.Now().Format("2006.01.02")
	if isToday {
		return t.Time.Format("03:04")
	}

	return t.Time.Format("2006.01.02 03:04")
}
