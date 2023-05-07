package web_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
)

type emptyRenderer struct{}

func (t *emptyRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return nil
}

func TestJobsController_JobsHome(t *testing.T) {
	t.Parallel()

	echoRouter := echo.New()
	echoRouter.Renderer = &emptyRenderer{} // todo move to arrower & maybe add constructor for echo tests
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
					return application.ListAllQueuesResponse{}, errors.New("some-error") //nolint:goerr113
				},
			},
		}

		assert.Error(t, handler.JobsHome()(c))
	})
}
