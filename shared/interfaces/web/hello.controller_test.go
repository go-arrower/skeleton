package web_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-arrower/arrower/app"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/shared/application"
	"github.com/go-arrower/skeleton/shared/domain"
	"github.com/go-arrower/skeleton/shared/interfaces/web"
)

func TestHelloController_SayHello(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.NewHelloController(application.App{
			SayHello: app.TestSuccessRequestHandler[application.SayHelloRequest, domain.TeamMember](),
		})

		assert.NoError(t, handler.SayHello()(c))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("usecase failure", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := echoRouter.NewContext(req, rec)

		handler := web.NewHelloController(application.App{
			SayHello: app.TestFailureRequestHandler[application.SayHelloRequest, domain.TeamMember](),
		})

		assert.NoError(t, handler.SayHello()(c))
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

type emptyRenderer struct{}

func (t *emptyRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return nil
}

// newTestRouter is a helper for unit tests, by returning a valid web router.
func newTestRouter() *echo.Echo {
	e := echo.New()
	e.Renderer = &emptyRenderer{}

	return e
}
