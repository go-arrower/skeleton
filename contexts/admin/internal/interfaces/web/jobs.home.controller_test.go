package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
)

func TestJobsController_JobsHome(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.JobsController{ //nolint:exhaustruct
			Cmds: application.JobsCommandContainer{
				ListAllQueues: func(context.Context, application.ListAllQueuesRequest) (application.ListAllQueuesResponse, error) {
					return application.ListAllQueuesResponse{}, nil
				},
			},
		}

		if assert.NoError(t, handler.JobsHome()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.JobsController{ //nolint:exhaustruct
			Cmds: application.JobsCommandContainer{
				ListAllQueues: func(context.Context, application.ListAllQueuesRequest) (application.ListAllQueuesResponse, error) {
					return application.ListAllQueuesResponse{}, errUCFailed
				},
			},
		}

		assert.Error(t, handler.JobsHome()(c))
	})
}

func TestJobsController_JobsQueue(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.JobsController{ //nolint:exhaustruct
			Cmds: application.JobsCommandContainer{
				GetQueue: func(context.Context, application.GetQueueRequest) (application.GetQueueResponse, error) {
					return application.GetQueueResponse{}, nil
				},
			},
		}

		if assert.NoError(t, handler.JobsQueue()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.JobsController{ //nolint:exhaustruct
			Cmds: application.JobsCommandContainer{
				GetQueue: func(context.Context, application.GetQueueRequest) (application.GetQueueResponse, error) {
					return application.GetQueueResponse{}, errUCFailed
				},
			},
		}

		assert.Error(t, handler.JobsQueue()(c))
	})
}

func TestJobsController_JobsWorkers(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.JobsController{
			Cmds: application.JobsCommandContainer{
				GetWorkers: func(context.Context, application.GetWorkersRequest) (application.GetWorkersResponse, error) {
					return application.GetWorkersResponse{}, nil
				},
			},
		}

		if assert.NoError(t, handler.JobsWorkers()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.JobsController{
			Cmds: application.JobsCommandContainer{
				GetWorkers: func(context.Context, application.GetWorkersRequest) (application.GetWorkersResponse, error) {
					return application.GetWorkersResponse{}, errUCFailed
				},
			},
		}

		assert.Error(t, handler.JobsWorkers()(c))
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

		handler := web.JobsController{
			Cmds: application.JobsCommandContainer{
				DeleteJob: func(ctx context.Context, in application.DeleteJobRequest) (application.DeleteJobResponse, error) {
					assert.Equal(t, "1337", in.JobID)

					return application.DeleteJobResponse{}, nil
				},
			},
		}

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

		handler := web.JobsController{
			Cmds: application.JobsCommandContainer{
				DeleteJob: func(ctx context.Context, in application.DeleteJobRequest) (application.DeleteJobResponse, error) {
					return application.DeleteJobResponse{}, errUCFailed
				},
			},
		}

		if assert.NoError(t, handler.DeleteJob()(c)) {
			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Equal(t, "/admin/jobs/Default", rec.Header().Get(echo.HeaderLocation))
		}
	})
}
