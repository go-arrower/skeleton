package web_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/web"
)

func TestUserController_Login(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()

	t.Run("already logged in", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		c := echoRouter.NewContext(req, rec)
		c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), auth.CtxAuthLoggedIn, true)))

		if assert.NoError(t, web.UserController{}.Login()(c)) {
			assert.Equal(t, http.StatusSeeOther, rec.Code)
		}
	})

	t.Run("show login form", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		c := echoRouter.NewContext(req, rec)

		if assert.NoError(t, web.UserController{}.Login()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), "login")
		}
	})

	t.Run("login succeeds", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=1337&password=12345678"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		controller := web.UserController{
			CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
				assert.Equal(t, "1337", in.LoginEmail)
				assert.Equal(t, "12345678", in.Password)

				return application.LoginUserResponse{}, nil
			},
		}

		echoRouter.POST("/login", controller.Login())
		echoRouter.ServeHTTP(rec, req)

		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Len(t, result.Cookies(), 1)
		assert.Equal(t, "/", result.Cookies()[0].Path)
		assert.Equal(t, "session", result.Cookies()[0].Name)
	})

	t.Run("login fails", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("login=1337&password=12345678"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		c := echoRouter.NewContext(req, rec)
		controller := web.UserController{
			CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
				assert.Equal(t, "1337", in.LoginEmail)
				assert.Equal(t, "12345678", in.Password)

				return application.LoginUserResponse{}, errUCFailed
			},
		}

		if assert.NoError(t, controller.Login()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), "login")
		}
	})
}

func TestUserController_Logout(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()

	t.Run("not logged in", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		c := echoRouter.NewContext(req, rec)
		// c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), auth.CtxAuthLoggedIn, false)))

		if assert.NoError(t, web.UserController{}.Logout()(c)) {
			assert.Equal(t, http.StatusSeeOther, rec.Code)
		}
	})

	t.Run("logout succeeds", func(t *testing.T) {
		t.Parallel()

		// log in first
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=1337&password=12345678"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		controller := web.UserController{
			CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
				return application.LoginUserResponse{}, nil
			},
		}

		echoRouter.POST("/login", controller.Login())
		echoRouter.ServeHTTP(rec, req)
		assert.Len(t, rec.Result().Cookies(), 1)

		// log out
		req = httptest.NewRequest(http.MethodGet, "/logout", nil)
		req.AddCookie(rec.Result().Cookies()[0])
		rec = httptest.NewRecorder()

		echoRouter.GET("/logout", controller.Logout())
		echoRouter.ServeHTTP(rec, req)

		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Len(t, result.Cookies(), 1)
		assert.Equal(t, "/", result.Cookies()[0].Path)
		assert.Equal(t, -1, result.Cookies()[0].MaxAge)
	})
}

var errUCFailed = errors.New("use case error")

type emptyRenderer struct{}

func (t *emptyRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	_, _ = w.Write([]byte(name))

	return nil
}

// newTestRouter is a helper for unit tests, by returning a valid web router.
func newTestRouter() *echo.Echo {
	e := echo.New()
	e.Renderer = &emptyRenderer{}
	e.Use(session.Middleware(sessions.NewFilesystemStore("", []byte("secret"))))
	e.Use(auth.EnrichCtxWithUserInfoMiddleware)

	return e
}
