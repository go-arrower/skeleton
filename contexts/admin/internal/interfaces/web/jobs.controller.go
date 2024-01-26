package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-arrower/arrower/alog"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
	"github.com/go-arrower/skeleton/shared/interfaces/web"
	"github.com/go-arrower/skeleton/shared/interfaces/web/views/pages"
)

const (
	defaultQueueName = "Default"

	historyTableSizeChanged = "arrower:admin.jobs.history.deleted"
)

func NewJobsController(logger alog.Logger, repo domain.Repository, presenter *web.DefaultPresenter, app *application.JobsApplication) *JobsController {
	return &JobsController{
		logger: logger,
		repo:   repo,
		p:      presenter,
		app:    app,
	}
}

type JobsController struct {
	logger alog.Logger
	repo   domain.Repository
	p      *web.DefaultPresenter

	Cmds    application.JobsCommandContainer
	Queries *models.Queries
	app     *application.JobsApplication
}

func (jc *JobsController) ListQueues() func(c echo.Context) error {
	return func(c echo.Context) error {
		res, err := jc.Cmds.ListAllQueues(c.Request().Context(), application.ListAllQueuesRequest{})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return c.Render(http.StatusOK, "=>jobs.home", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Queues", echo.Map{
			"Queues": res.QueueStats,
		}))
	}
}

func (jc *JobsController) PendingJobsPieChartData() func(echo.Context) error {
	type pieData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	return func(c echo.Context) error {
		res, err := jc.Cmds.ListAllQueues(c.Request().Context(), application.ListAllQueuesRequest{})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		var json []pieData
		for _, q := range res.QueueStats {
			json = append(json, pieData{Name: string(q.QueueName), Value: q.PendingJobs})
		}

		return c.JSON(http.StatusOK, json)
	}
}

func (jc *JobsController) ProcessedJobsLineChartData() func(echo.Context) error {
	type lineData struct {
		XAxis  []string `json:"xAxis"`
		Series []int    `json:"series"`
	}

	return func(c echo.Context) error {
		interval := c.Param("interval")

		var data []models.PendingJobsRow
		var err error

		if interval == "hour" { // show last 60 minutes
			data, err = jc.Queries.PendingJobs(c.Request().Context(), models.PendingJobsParams{
				DateBin:    pgtype.Interval{Valid: true, Microseconds: int64(time.Minute * 5 / time.Microsecond)},
				FinishedAt: pgtype.Timestamptz{Valid: true, Time: time.Now().UTC().Add(-time.Hour)},
				Limit:      12,
			})
		} else {
			data, err = jc.Queries.PendingJobs(c.Request().Context(), models.PendingJobsParams{ // show whole week
				DateBin:    pgtype.Interval{Valid: true, Days: 1},
				FinishedAt: pgtype.Timestamptz{Valid: true, Time: time.Now().UTC().Add(-time.Hour * 24 * 7)},
				Limit:      7,
			})
		}
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		var xaxis []string
		var series []int

		for _, d := range data {
			if interval == "hour" {
				xaxis = append([]string{d.T.Time.Format("15:04")}, xaxis...)
			} else {
				xaxis = append([]string{d.T.Time.Format("01.02")}, xaxis...)
			}

			series = append([]int{int(d.Count)}, series...)
		}

		return c.JSON(http.StatusOK, lineData{
			XAxis:  xaxis,
			Series: series,
		})
	}
}

func (jc *JobsController) ShowQueue() func(c echo.Context) error {
	return func(c echo.Context) error {
		queue := c.Param("queue")

		res, err := jc.Cmds.GetQueue(c.Request().Context(), application.GetQueueRequest{
			QueueName: queue,
		})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		page := buildQueuePage(queue, res.Jobs, res.Kpis)

		return c.Render(http.StatusOK, "=>jobs.queue", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Queue "+page.QueueName, echo.Map{
			"QueueName": page.QueueName,
			"Jobs":      page.Jobs,
			"Stats":     page.Stats,
		}))
	}
}

func (jc *JobsController) ListWorkers() func(c echo.Context) error {
	return func(c echo.Context) error {
		res, err := jc.Cmds.GetWorkers(c.Request().Context(), application.GetWorkersRequest{})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return c.Render(http.StatusOK, "=>jobs.workers", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Worker", echo.Map{
			"workers": presentWorkers(res.Pool),
		}))
	}
}

func (jc *JobsController) DeleteJob() func(c echo.Context) error {
	return func(c echo.Context) error {
		q := c.Param("queue")
		jobID := c.Param("job_id")

		_ = jc.Cmds.DeleteJob(c.Request().Context(), application.DeleteJobRequest{JobID: jobID})

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/jobs/%s", q))
	}
}

func (jc *JobsController) RescheduleJob() func(c echo.Context) error {
	return func(c echo.Context) error {
		q := c.Param("queue")
		jobID := c.Param("job_id")

		_ = jc.Cmds.RescheduleJob(c.Request().Context(), application.RescheduleJobRequest{JobID: jobID})

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/jobs/%s", q))
	}
}

func (jc *JobsController) ShowSettings() func(c echo.Context) error {
	return func(c echo.Context) error {
		size, _ := jc.Queries.JobTableSize(c.Request().Context())

		res, _ := jc.Cmds.ListAllQueues(c.Request().Context(), application.ListAllQueuesRequest{})

		var queues []string
		for q, _ := range res.QueueStats {
			queue := string(q)
			if queue == "" {
				queue = "Default"
			}

			queues = append(queues, queue)
		}

		return c.Render(http.StatusOK, "jobs.settings", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Settings", echo.Map{
			"Jobs":    size.Jobs,
			"History": size.History,
			"Queues":  queues,
		}))
	}
}

func (jc *JobsController) VacuumJobTables() func(echo.Context) error {
	return func(c echo.Context) error {
		table := c.Param("table")

		_ = jc.app.VacuumJobsTable(c.Request().Context(), table)

		// reload the dashboard badges with the size, by using htmx's oob technique
		size, _ := jc.Queries.JobTableSize(c.Request().Context())

		return c.Render(http.StatusOK, "jobs.settings#table-size", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Settings", echo.Map{
			"Jobs":    size.Jobs,
			"History": size.History,
		}))
	}
}

func (jc *JobsController) DeleteHistory() func(echo.Context) error {
	return func(c echo.Context) error {
		days, _ := strconv.Atoi(c.FormValue("days"))

		if days == 0 && c.FormValue("days") != "all" {
			return c.NoContent(http.StatusBadRequest)
		}

		_ = jc.app.PruneHistory(c.Request().Context(), days)

		// reload the dashboard badges with the size, by using htmx's oob technique
		size, _ := jc.Queries.JobTableSize(c.Request().Context())

		c.Response().Header().Set("HX-Trigger", historyTableSizeChanged)

		return c.Render(http.StatusOK, "jobs.settings#table-size", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Settings", echo.Map{
			"Jobs":    size.Jobs,
			"History": size.History,
		}))
	}
}

func (jc *JobsController) PruneHistory() func(echo.Context) error {
	return func(c echo.Context) error {
		days, _ := strconv.Atoi(c.FormValue("days"))
		estimateBefore := time.Now().Add(-1 * time.Duration(days) * time.Hour * 24)

		queue := c.FormValue("queue")
		if queue == "Default" {
			queue = ""
		}

		_ = jc.Queries.PruneHistoryPayload(c.Request().Context(), models.PruneHistoryPayloadParams{
			Queue:     queue,
			CreatedAt: pgtype.Timestamptz{Time: estimateBefore, Valid: true},
		})

		c.Response().Header().Set("HX-Trigger", historyTableSizeChanged)

		return c.NoContent(http.StatusOK)
	}
}

func (jc *JobsController) EstimateHistorySize() func(echo.Context) error {
	return func(c echo.Context) error {
		days, _ := strconv.Atoi(c.QueryParam("days"))

		estimateBefore := time.Now().Add(-1 * time.Duration(days) * time.Hour * 24)

		size, _ := jc.Queries.JobHistorySize(c.Request().Context(), pgtype.Timestamptz{Time: estimateBefore, Valid: true})

		var fmtSize string
		if size != "" {
			fmtSize = fmt.Sprintf("~ %s", size)
		}

		return c.String(http.StatusOK, fmtSize)
	}
}

func (jc *JobsController) EstimateHistoryPayloadSize() func(echo.Context) error {
	return func(c echo.Context) error {
		queue := c.QueryParam("queue")
		if queue == "Default" {
			queue = ""
		}

		days, _ := strconv.Atoi(c.QueryParam("days"))
		estimateBefore := time.Now().Add(-1 * time.Duration(days) * time.Hour * 24)

		size, _ := jc.Queries.JobHistoryPayloadSize(c.Request().Context(), models.JobHistoryPayloadSizeParams{
			Queue:     queue,
			CreatedAt: pgtype.Timestamptz{Time: estimateBefore, Valid: true},
		})

		var fmtSize string
		if size != "" {
			fmtSize = fmt.Sprintf("~ %s", size)
		}

		return c.String(http.StatusOK, fmtSize)
	}
}

func (jc *JobsController) CreateJobs() func(c echo.Context) error {
	return func(c echo.Context) error {
		res, _ := jc.Cmds.ListAllQueues(c.Request().Context(), application.ListAllQueuesRequest{})

		var queues []string
		for q, _ := range res.QueueStats {
			queue := string(q)
			if queue == "" {
				queue = "Default"
			}

			queues = append(queues, queue)
		}

		jobType, _ := jc.Queries.JobTypes(c.Request().Context(), "")
		_ = jobType

		return c.Render(http.StatusOK, "jobs.schedule", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Schedule a Job", echo.Map{
			"Queues":   queues,
			"JobTypes": jobType,
		}))
	}
}

func (jc *JobsController) ShowJobTypes() func(_ echo.Context) error {
	return func(c echo.Context) error {
		queue := c.QueryParam("queue")

		if queue == "Default" {
			queue = ""
		}

		jobType, _ := jc.Queries.JobTypes(c.Request().Context(), queue)

		return c.Render(http.StatusOK, "jobs.schedule#known-job-types", echo.Map{
			"JobTypes": jobType,
		})
	}
}

func (jc *JobsController) ScheduleJobs() func(c echo.Context) error {
	return func(c echo.Context) error {
		queue := c.FormValue("queue")
		jt := c.FormValue("job-type")
		prio := c.FormValue("priority")
		payload := c.FormValue("payload")
		num := c.FormValue("count")

		jq := queue
		if queue == "Default" {
			jq = ""
		}

		priority, err := strconv.Atoi(prio)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		priority = priority * -1

		count, err := strconv.Atoi(num)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		err = jc.Cmds.ScheduleJobs(c.Request().Context(), application.ScheduleJobsRequest{
			Queue:    jq,
			JobType:  jt,
			Priority: int16(priority),
			Payload:  payload,
			Count:    count,
		})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/jobs/%s", queue))
	}
}

func presentWorkers(pool []domain.WorkerPool) []pages.JobWorker {
	jobWorkers := make([]pages.JobWorker, len(pool))

	for i, _ := range pool {
		jobWorkers[i].ID = pool[i].ID
		jobWorkers[i].Queue = pool[i].Queue
		jobWorkers[i].Workers = pool[i].Workers
		jobWorkers[i].LastSeenAt = pool[i].LastSeen.Format(time.TimeOnly)

		var warningTimeWorkerPoolNotSeenSince time.Duration = 30

		jobWorkers[i].LastSeenAtColour = "text-green-600"
		if time.Since(pool[i].LastSeen)/time.Second > warningTimeWorkerPoolNotSeenSince {
			jobWorkers[i].LastSeenAtColour = "text-orange-600"
		}

		jobWorkers[i].NotSeenSince = notSeenSinceTimeString(pool[i].LastSeen)
	}

	return jobWorkers
}

func notSeenSinceTimeString(t time.Time) string {
	seconds := time.Since(t).Seconds()

	secondsPerMinute := 60
	if seconds > float64(secondsPerMinute) {
		return fmt.Sprintf("%d m %d sec", int(seconds/float64(secondsPerMinute)), int(seconds)%secondsPerMinute)
	}

	return fmt.Sprintf("%d sec", int(seconds))
}

type (
	QueueStats struct {
		PendingJobsPerType   map[string]int
		QueueName            string
		PendingJobs          int
		FailedJobs           int
		ProcessedJobs        int
		AvailableWorkers     int
		PendingJobsErrorRate float64 // can be calculated: FailedJobs * 100 / PendingJobs
		AverageTimePerJob    time.Duration
		EstimateUntilEmpty   time.Duration // can be calculated
	}

	ListQueuesPage struct {
		Queues map[domain.QueueName]domain.QueueStats
	}

	QueuePage struct {
		Jobs      []domain.PendingJob
		QueueName string
		Stats     QueueStats
	}
)

func buildQueuePage(queue string, jobs []domain.PendingJob, kpis domain.QueueKPIs) QueuePage {
	if queue == "" {
		queue = defaultQueueName
	}

	jobs = prettyFormatPayload(jobs)

	return QueuePage{
		QueueName: queue,
		Stats:     queueKpiToStats(queue, kpis),

		Jobs: jobs,
	}
}

func prettyFormatPayload(jobs []domain.PendingJob) []domain.PendingJob {
	for i := 0; i < len(jobs); i++ { //nolint:varnamelen
		var m map[string]interface{}
		_ = json.Unmarshal([]byte(jobs[i].Payload), &m)

		data, _ := json.MarshalIndent(m, "", "  ") //nolint:errchkjson // trust the type checks to work for simplicity
		jobs[i].Payload = string(data)

		if jobs[i].Queue == "" {
			jobs[i].Queue = defaultQueueName
		}
	}

	return jobs
}

func queueKpiToStats(queue string, kpis domain.QueueKPIs) QueueStats {
	if queue == "" {
		queue = defaultQueueName
	}

	var errorRate float64

	if kpis.FailedJobs != 0 {
		errorRate = float64(kpis.FailedJobs * 100 / kpis.PendingJobs)
	}

	var duration time.Duration
	if kpis.AvailableWorkers != 0 {
		duration = time.Duration(kpis.PendingJobs/kpis.AvailableWorkers) * kpis.AverageTimePerJob
	}

	return QueueStats{
		QueueName:            queue,
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
