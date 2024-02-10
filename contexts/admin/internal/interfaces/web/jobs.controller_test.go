package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

		//handler := &web.JobsController{
		//	Cmds: application.JobsCommandContainer{
		//		ListAllQueues: func(context.Context, application.ListAllQueuesRequest) (application.ListAllQueuesResponse, error) {
		//			return application.ListAllQueuesResponse{}, nil
		//		},
		//	},
		//}

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsSuccessApplication())

		if assert.NoError(t, handler.ListQueues()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		//handler := web.JobsController{
		//	Cmds: application.JobsCommandContainer{
		//		ListAllQueues: func(context.Context, application.ListAllQueuesRequest) (application.ListAllQueuesResponse, error) {
		//			return application.ListAllQueuesResponse{}, errUCFailed
		//		},
		//	},
		//}
		handler := web.NewJobsController(nil, nil, nil, application.NewJobsFailureApplication())

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

		//handler := web.JobsController{
		//	Cmds: application.JobsCommandContainer{
		//		GetQueue: func(context.Context, application.GetQueueRequest) (application.GetQueueResponse, error) {
		//			return application.GetQueueResponse{}, nil
		//		},
		//	},
		//}

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsSuccessApplication())

		if assert.NoError(t, handler.ShowQueue()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		//handler := web.JobsController{
		//	Cmds: application.JobsCommandContainer{
		//		GetQueue: func(context.Context, application.GetQueueRequest) (application.GetQueueResponse, error) {
		//			return application.GetQueueResponse{}, errUCFailed
		//		},
		//	},
		//}

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsFailureApplication())

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

		//handler := web.JobsController{
		//	Cmds: application.JobsCommandContainer{
		//		GetWorkers: func(context.Context, application.GetWorkersRequest) (application.GetWorkersResponse, error) {
		//			return application.GetWorkersResponse{}, nil
		//		},
		//	},
		//}

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsSuccessApplication())

		if assert.NoError(t, handler.ListWorkers()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		//handler := web.JobsController{
		//	Cmds: application.JobsCommandContainer{
		//		GetWorkers: func(context.Context, application.GetWorkersRequest) (application.GetWorkersResponse, error) {
		//			return application.GetWorkersResponse{}, errUCFailed
		//		},
		//	},
		//}

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsFailureApplication())

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

		//handler := web.JobsController{
		//	Cmds: application.JobsCommandContainer{
		//		DeleteJob: func(ctx context.Context, in application.DeleteJobRequest) error {
		//			assert.Equal(t, "1337", in.JobID)
		//
		//			return nil
		//		},
		//	},
		//}

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsSuccessApplication())

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

		//handler := web.JobsController{
		//	Cmds: application.JobsCommandContainer{
		//		DeleteJob: func(ctx context.Context, in application.DeleteJobRequest) error {
		//			return errUCFailed
		//		},
		//	},
		//}

		handler := web.NewJobsController(nil, nil, nil, application.NewJobsFailureApplication())

		if assert.NoError(t, handler.DeleteJob()(c)) {
			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Equal(t, "/admin/jobs/Default", rec.Header().Get(echo.HeaderLocation))
		}
	})
}
