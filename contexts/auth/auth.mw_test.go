package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth"
)

func TestJobsController_JobsHome(t *testing.T) {
	t.Parallel()

	t.Run("no session => no id", func(t *testing.T) {
		t.Parallel()

		echoRouter := newTestRouterToAssertOnHandler(func(c echo.Context) error {
			ctx := c.Request().Context()
			assert.False(t, auth.IsLoggedIn(ctx))
			assert.Empty(t, auth.CurrentUserID(ctx))

			return c.NoContent(http.StatusOK) //nolint:wrapcheck
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		echoRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("logged in user", func(t *testing.T) {
		t.Parallel()

		echoRouter := newTestRouterToAssertOnHandler(func(c echo.Context) error {
			ctx := c.Request().Context()
			assert.True(t, auth.IsLoggedIn(ctx))
			assert.Equal(t, auth.UserID("1337"), auth.CurrentUserID(ctx))

			return c.NoContent(http.StatusOK) //nolint:wrapcheck
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(getSessionCookie(echoRouter))
		rec := httptest.NewRecorder()

		echoRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// newTestRouterToAssertOnHandler is a helper for unit tests, by returning a valid web router.
func newTestRouterToAssertOnHandler(handler func(c echo.Context) error) *echo.Echo {
	echoRouter := echo.New()

	echoRouter.Use(session.Middleware(sessions.NewFilesystemStore("", []byte("secret"))))
	echoRouter.Use(auth.EnrichCtxWithUserInfoMiddleware)
	echoRouter.GET("/", handler)

	// endpoint to set an example cookie, that the middleware under test can work with.
	echoRouter.GET("/createSession", func(c echo.Context) error {
		sess, _ := session.Get("session", c)

		sess.Values["auth.user_logged_in"] = true
		sess.Values["auth.user_id"] = "1337"

		_ = sess.Save(c.Request(), c.Response())

		return c.NoContent(http.StatusOK) //nolint:wrapcheck
	})

	return echoRouter
}

// getSessionCookie calls the /createSession route setup in newTestRouterToAssertOnHandler.
func getSessionCookie(e *echo.Echo) *http.Cookie {
	req := httptest.NewRequest(http.MethodGet, "/createSession", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	response := rec.Result()
	defer response.Body.Close()

	return response.Cookies()[0]
}