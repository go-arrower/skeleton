package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"

	"github.com/go-arrower/arrower/alog"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
	"github.com/go-arrower/skeleton/shared/interfaces/web"
	"github.com/go-arrower/skeleton/shared/interfaces/web/views/pages"
)

const (
	defaultQueueName = "Default"

	historyTableSizeChanged = "arrower:admin.jobs.history.deleted"

	htmlDatetimeLayout = "2006-01-02T15:04" // format used by the HTML datetime-local input element
)

func NewJobsController(logger alog.Logger, repo jobs.Repository, presenter *web.DefaultPresenter, app *application.JobsApplication) *JobsController {
	return &JobsController{
		logger: logger,
		repo:   repo,
		p:      presenter,
		app:    app,
	}
}

type JobsController struct {
	logger alog.Logger
	repo   jobs.Repository
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

		return c.Render(http.StatusOK, "jobs.home", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Queues",
			echo.Map{
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

		return c.Render(http.StatusOK, "jobs.workers", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Worker", echo.Map{
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

func (jc *JobsController) ShowMaintenance() func(c echo.Context) error {
	return func(c echo.Context) error {
		size, _ := jc.Queries.JobTableSize(c.Request().Context())

		res, _ := jc.Cmds.ListAllQueues(c.Request().Context(), application.ListAllQueuesRequest{}) // fixme: don't call existing use case, create own or call domain model

		var queues []string
		for q, _ := range res.QueueStats {
			queue := string(q)
			if queue == "" {
				queue = "Default"
			}

			queues = append(queues, queue)
		}

		return c.Render(http.StatusOK, "jobs.maintenance", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Maintenance", echo.Map{
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

		return c.Render(http.StatusOK, "jobs.maintenance#table-size", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Settings", echo.Map{
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

		return c.Render(http.StatusOK, "jobs.maintenance#table-size", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Settings", echo.Map{
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
		queues, _ := jc.app.Queues(c.Request().Context())

		jobType, _ := jc.app.JobTypesForQueue(c.Request().Context(), jobs.DefaultQueueName)

		y, m, d := time.Now().Date()

		return c.Render(http.StatusOK, "jobs.schedule", jc.p.MustMapDefaultBasePage(c.Request().Context(), "Schedule a Job", echo.Map{
			"Queues":   queues,
			"JobTypes": jobType,
			"RunAt":    time.Now().Format(htmlDatetimeLayout),
			"RunAtMin": fmt.Sprintf("%d-%02d-%02dT00:00", y, m, d),
		}))
	}
}

func (jc *JobsController) ShowJobTypes() func(_ echo.Context) error {
	return func(c echo.Context) error {
		queue := c.QueryParam("queue")

		jobType, _ := jc.app.JobTypesForQueue(c.Request().Context(), jobs.QueueName(queue))

		return c.Render(http.StatusOK, "jobs.schedule#known-job-types", echo.Map{
			"JobTypes": jobType,
		})
	}
}

func (jc *JobsController) PayloadExamples() func(_ echo.Context) error {
	return func(c echo.Context) error {
		queue := c.QueryParam("queue")
		jobType := c.QueryParam("job-type")

		if jobs.QueueName(queue) == jobs.DefaultQueueName {
			queue = ""
		}

		payloads, _ := jc.Queries.LastHistoryPayloads(c.Request().Context(), models.LastHistoryPayloadsParams{
			Queue:   queue,
			JobType: jobType,
		})

		pJson := make([]string, len(payloads))
		for i, p := range payloads {
			var jobPayload application.JobPayload
			_ = json.Unmarshal(p, &jobPayload)

			pJson[i] = prettyString(jobPayload.JobData)
		}

		if queue == "" {
			queue = defaultQueueName
		}

		return c.Render(http.StatusOK, "jobs.schedule#payload-examples", echo.Map{
			"Payloads": pJson,
			"Queue":    queue,
			"JobType":  jobType,
		})
	}
}

func prettyString(str string) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "  "); err != nil {
		return ""
	}

	return prettyJSON.String()
}

func (jc *JobsController) ScheduleJobs() func(c echo.Context) error {
	return func(c echo.Context) error {
		queue := c.FormValue("queue")
		jt := c.FormValue("job-type")
		prio := c.FormValue("priority")
		payload := c.FormValue("payload")
		num := c.FormValue("count")
		t := c.FormValue("runAt-time")

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

		runAt, _ := time.Parse(htmlDatetimeLayout, t)

		err = jc.Cmds.ScheduleJobs(c.Request().Context(), application.ScheduleJobsRequest{
			Queue:    jq,
			JobType:  jt,
			Priority: int16(priority),
			Payload:  payload,
			Count:    count,
			RunAt:    runAt.Add(-1 * time.Hour), // todo needs to apply read tz, to prevent dirty hack, to overcome client and server times
		})
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/jobs/%s", queue))
	}
}

func presentWorkers(pool []jobs.WorkerPool) []pages.JobWorker {
	jobWorkers := make([]pages.JobWorker, len(pool))

	for i, _ := range pool {
		jobWorkers[i].ID = pool[i].ID
		jobWorkers[i].Queue = pool[i].Queue
		jobWorkers[i].Workers = pool[i].Workers
		jobWorkers[i].Version = pool[i].Version
		jobWorkers[i].JobTypes = pool[i].JobTypes

		sort.Slice(jobWorkers[i].JobTypes, func(ii, ij int) bool {
			return jobWorkers[i].JobTypes[ii] <= jobWorkers[i].JobTypes[ij]
		})

		var warningSecondsWorkerPoolNotSeenSince time.Duration = 30

		jobWorkers[i].LastSeenAtColourSuccess = true
		if time.Since(pool[i].LastSeen)/time.Second >= warningSecondsWorkerPoolNotSeenSince {
			jobWorkers[i].LastSeenAtColourSuccess = false
		}

		jobWorkers[i].NotSeenSince = notSeenSinceTimeString(pool[i].LastSeen, warningSecondsWorkerPoolNotSeenSince)
	}

	sort.Slice(jobWorkers, func(i, j int) bool {
		return jobWorkers[i].ID <= jobWorkers[j].ID
	})

	return jobWorkers
}

func notSeenSinceTimeString(t time.Time, warningSecondsWorkerPoolNotSeenSince time.Duration) string {
	seconds := time.Since(t).Seconds()

	if time.Duration(seconds) >= warningSecondsWorkerPoolNotSeenSince && seconds < 60 {
		return "recently"
	}

	secondsPerMinute := 60.0
	if seconds > secondsPerMinute {
		minutes := int(math.Round(seconds / secondsPerMinute))
		if minutes == 1 {
			return fmt.Sprintf("%d minute ago", minutes)
		}

		return fmt.Sprintf("%d minutes ago", minutes)
	}

	return "now"
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
		Queues map[jobs.QueueName]jobs.QueueStats
	}

	QueuePage struct {
		Jobs      []jobs.PendingJob
		QueueName string
		Stats     QueueStats
	}
)

func buildQueuePage(queue string, jobs []jobs.PendingJob, kpis jobs.QueueKPIs) QueuePage {
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

func prettyFormatPayload(jobs []jobs.PendingJob) []jobs.PendingJob {
	for i := 0; i < len(jobs); i++ { //nolint:varnamelen
		var m application.JobPayload
		_ = json.Unmarshal([]byte(jobs[i].Payload), &m)

		data, _ := json.MarshalIndent(m.JobData, "", "  ") //nolint:errchkjson // trust the type checks to work for simplicity
		jobs[i].Payload = string(data)
		jobs[i].Payload = m.JobData

		if jobs[i].Queue == "" {
			jobs[i].Queue = defaultQueueName
			jobs[i].RunAtFmt = fmtRunAtTime(jobs[i].RunAt)
		}
	}

	return jobs
}

func fmtRunAtTime(t time.Time) string {
	now := time.Now()

	isToday := t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
	if isToday {
		return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
	}

	return t.Format("2006.01.02 15:04")
}

func queueKpiToStats(queue string, kpis jobs.QueueKPIs) QueueStats {
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
		AverageTimePerJob:    kpis.AverageTimePerJob.Truncate(time.Millisecond),
		EstimateUntilEmpty:   duration.Truncate(time.Millisecond),
	}
}
