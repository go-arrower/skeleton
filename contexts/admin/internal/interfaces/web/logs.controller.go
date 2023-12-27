package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/go-arrower/arrower/alog"
	"github.com/go-arrower/arrower/alog/models"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/shared/interfaces/web"
)

func NewLogsController(logger alog.Logger, queries *models.Queries, routes *echo.Group, presenter *web.DefaultPresenter) *LogsController {
	return &LogsController{
		logger:  logger,
		queries: queries,
		r:       routes,
		p:       presenter,
	}
}

type LogsController struct {
	logger  alog.Logger
	queries *models.Queries
	r       *echo.Group
	p       *web.DefaultPresenter
}

func (lc *LogsController) ShowLogs() {
	// FIXME: how to add route with and without trailing slash

	type filter struct {
		Range int    `query:"range"`
		Level string `query:"level"`
		K0    string `query:"k0"`
		F0    string `query:"f0"`
		K1    string `query:"k1"`
		F1    string `query:"f1"`
		K2    string `query:"k2"`
		F2    string `query:"f2"`
	}
	type log struct {
		Time   time.Time
		UserID string
		Log    map[string]any
	}

	lc.r.GET("/", func(c echo.Context) error {
		timeParam := c.QueryParam("time")
		searchMsgParam := c.QueryParam("msg")

		var filter filter
		err := c.Bind(&filter)
		if err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}

		if filter.Range == 0 {
			filter.Range = 15
		}

		filterTime := time.Now().UTC().Add(-time.Duration(filter.Range) * time.Minute)
		if t, err := time.Parse("2006-01-02T15:04:05.999999999", timeParam); err == nil {
			filterTime = t
		}

		filterTime = filterTime.Add(-1 * time.Hour)

		level := []string{"INFO"}
		if filter.Level == "DEBUG" {
			level = []string{"INFO", "DEBUG"}
		}

		queryParams := models.GetRecentLogsParams{
			Time:  pgtype.Timestamptz{Time: filterTime, Valid: true},
			Msg:   "%" + searchMsgParam + "%",
			Level: level,
			Limit: 1000,
		}

		if filter.K0 != "" {
			queryParams.F0 = fmt.Sprintf(`$.**.%s ? (@ like_regex "^.*%s.*" flag "i")`, filter.K0, filter.F0)
		}
		if filter.K1 != "" {
			queryParams.F1 = fmt.Sprintf(`$.**.%s ? (@ like_regex "^.*%s.*" flag "i")`, filter.K1, filter.F1)
		}
		if filter.K2 != "" {
			queryParams.F2 = fmt.Sprintf(`$.**.%s ? (@ like_regex "^.*%s.*" flag "i")`, filter.K2, filter.F2)
		}

		rawLogs, _ := lc.queries.GetRecentLogs(c.Request().Context(), queryParams)

		var logs []log
		for _, l := range rawLogs {
			log := log{}

			json.Unmarshal(l.Log, &log.Log)
			log.Time = l.Time.Time
			log.UserID = l.UserID.UUID.String()
			if (uuid.NullUUID{}) == l.UserID {
				log.UserID = ""
			}

			logs = append(logs, log)
		}

		vals := echo.Map{
			"SearchMsg": searchMsgParam,
			"Logs":      logs,
			"Filter":    filter,
		}

		if len(logs) == 0 {
			vals["LastLogTime"] = time.Now()
		}

		return c.Render(http.StatusOK, "=>logs.show", lc.p.MustMapDefaultBasePage(c.Request().Context(), "Logs", vals))
	}).Name = "admin.logs"
}