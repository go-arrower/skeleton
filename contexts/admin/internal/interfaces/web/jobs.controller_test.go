package web_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-arrower/arrower/app"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
)

func TestJobsController_JobsHome(t *testing.T) { //nolint:dupl
	t.Parallel()

	echoRouter := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.NewJobsController(nil, nil, nil, nil, application.NewJobsSuccessApplication(), application.App{})

		if assert.NoError(t, handler.ListQueues()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.NewJobsController(nil, nil, nil, nil, application.NewJobsFailureApplication(), application.App{})

		assert.Error(t, handler.ListQueues()(c))
	})
}

func TestJobsController_JobsQueue(t *testing.T) { //nolint:dupl
	t.Parallel()

	echoRouter := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.NewJobsController(nil, nil, nil, nil, application.NewJobsSuccessApplication(), application.App{})

		if assert.NoError(t, handler.ShowQueue()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.NewJobsController(nil, nil, nil, nil, application.NewJobsFailureApplication(), application.App{})

		assert.Error(t, handler.ShowQueue()(c))
	})
}

func TestJobsController_JobsWorkers(t *testing.T) { //nolint:dupl
	t.Parallel()

	echoRouter := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.NewJobsController(nil, nil, nil, nil, application.NewJobsSuccessApplication(), application.App{})

		if assert.NoError(t, handler.ListWorkers()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.NewJobsController(nil, nil, nil, nil, application.NewJobsFailureApplication(), application.App{})

		assert.Error(t, handler.ListWorkers()(c))
	})
}

func TestJobsController_DeleteJob(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		c.SetPath("/:queue/delete/:job_id")
		c.SetParamNames("queue", "job_id")
		c.SetParamValues("Default", "1337")

		handler := web.NewJobsController(nil, nil, nil, nil, application.NewJobsSuccessApplication(), application.App{})

		if assert.NoError(t, handler.DeleteJob()(c)) {
			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Equal(t, "/admin/jobs/Default", rec.Header().Get(echo.HeaderLocation))
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		c.SetPath("/:queue/delete/:job_id")
		c.SetParamNames("queue", "job_id")
		c.SetParamValues("Default", "1337")

		handler := web.NewJobsController(nil, nil, nil, nil, application.NewJobsFailureApplication(), application.App{})

		if assert.NoError(t, handler.DeleteJob()(c)) {
			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Equal(t, "/admin/jobs/Default", rec.Header().Get(echo.HeaderLocation))
		}
	})
}

func TestJobsController_DeleteHistory(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()

	// set http POST payload
	f := make(url.Values)
	f.Set("days", "all")
	validRequest := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	validRequest.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(validRequest, rec)

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsSuccessApplication(),
			application.App{
				PruneJobHistory: app.TestSuccessRequestHandler[application.PruneJobHistoryRequest, application.PruneJobHistoryResponse](),
			},
		)

		assert.NoError(t, handler.DeleteHistory()(c))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "arrower:admin.jobs.history.deleted", rec.Header().Get("HX-Trigger"))
	})

	t.Run("usecase failure", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(validRequest, rec)

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsSuccessApplication(),
			application.App{
				PruneJobHistory: app.TestFailureRequestHandler[application.PruneJobHistoryRequest, application.PruneJobHistoryResponse](),
			},
		)

		assert.NoError(t, handler.DeleteHistory()(c))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "", rec.Header().Get("HX-Trigger"))
	})
}
