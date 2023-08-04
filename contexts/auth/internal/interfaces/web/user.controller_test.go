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

	t.Run("redirect if already logged in", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		echoRouter := newTestRouter()
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

		echoRouter := newTestRouter()
		c := echoRouter.NewContext(req, rec)

		if assert.NoError(t, web.UserController{}.Login()(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), "login")
		}
	})

	t.Run("login succeeds", func(t *testing.T) {
		t.Parallel()

		t.Skip() // THE TEST IS PROPER, it fails because of the filesystemStore, see: https://github.com/gorilla/sessions/issues/267

		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("login=1337&password=12345678"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "arrower/0")
		rec := httptest.NewRecorder()

		controller := web.UserController{
			CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
				assert.Equal(t, "1337", in.LoginEmail)
				assert.Equal(t, "12345678", in.Password)
				assert.NotEmpty(t, in.IP)
				assert.NotEmpty(t, in.UserAgent)
				assert.NotEmpty(t, in.SessionKey)

				return application.LoginUserResponse{}, nil
			},
		}

		echoRouter := newTestRouter()
		echoRouter.POST("/login", controller.Login())
		echoRouter.ServeHTTP(rec, req)

		result := rec.Result()
		defer result.Body.Close()

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Empty(t, rec.Body.String())
		assert.Len(t, rec.Result().Cookies(), 2, "login session and known_device cookie expected")
		assert.Equal(t, "/", result.Cookies()[0].Path)
		assert.Equal(t, "session", result.Cookies()[0].Name)
		assert.Equal(t, 0, rec.Result().Cookies()[0].MaxAge, "cookie should expire when browser closes")
		assert.Equal(t, http.SameSiteStrictMode, result.Cookies()[0].SameSite)
	})

	t.Run("login fails", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", loginPostPayload())
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		controller := web.UserController{
			CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
				assert.Equal(t, "1337", in.LoginEmail)
				assert.Equal(t, "12345678", in.Password)

				return application.LoginUserResponse{}, errUCFailed
			},
		}

		echoRouter := newTestRouter()
		echoRouter.POST("/", controller.Login())
		echoRouter.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "login")
		assert.Len(t, rec.Result().Cookies(), 0, "failed logins should not have a known_device cookie")
	})

	t.Run("unknown device succeeds login", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", loginPostPayload())
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		controller := web.UserController{
			CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
				assert.True(t, in.IsNewDevice)

				return application.LoginUserResponse{}, nil
			},
		}

		echoRouter := newTestRouter()
		echoRouter.POST("/", controller.Login())
		echoRouter.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Len(t, rec.Result().Cookies(), 2, "login session and known_device cookie expected")
		assert.Equal(t, "/auth", rec.Result().Cookies()[1].Path)
		assert.Equal(t, "arrower.auth.known_device", rec.Result().Cookies()[1].Name)
		assert.Equal(t, http.SameSiteStrictMode, rec.Result().Cookies()[1].SameSite)

		t.Run("known device succeeds login", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", loginPostPayload())
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(rec.Result().Cookies()[1])
			rec := httptest.NewRecorder()

			controller := web.UserController{
				CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
					assert.False(t, in.IsNewDevice)

					return application.LoginUserResponse{}, nil
				},
			}

			echoRouter.POST("/", controller.Login())
			echoRouter.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Empty(t, rec.Body.String())
			assert.Len(t, rec.Result().Cookies(), 2, "login session and known_device cookie expected")
			assert.Equal(t, "/auth", rec.Result().Cookies()[1].Path)
			assert.Equal(t, "arrower.auth.known_device", rec.Result().Cookies()[1].Name)
			assert.Equal(t, http.SameSiteStrictMode, rec.Result().Cookies()[1].SameSite)
		})
	})

	t.Run("remember me increases cookie lifetime", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("remember_me=true"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		controller := web.UserController{
			CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
				return application.LoginUserResponse{}, nil
			},
		}

		echoRouter := newTestRouter()
		echoRouter.POST("/", controller.Login())
		echoRouter.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Len(t, rec.Result().Cookies(), 2, "login session and known_device cookie expected")
		assert.Equal(t, "/", rec.Result().Cookies()[0].Path)
		assert.Equal(t, "session", rec.Result().Cookies()[0].Name)
		assert.Equal(t, 60*60*24*30, rec.Result().Cookies()[0].MaxAge)
		assert.Equal(t, http.SameSiteStrictMode, rec.Result().Cookies()[0].SameSite)
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

		req := httptest.NewRequest(http.MethodPost, "/login", loginPostPayload())
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		controller := web.UserController{
			CmdLoginUser: func(ctx context.Context, in application.LoginUserRequest) (application.LoginUserResponse, error) {
				return application.LoginUserResponse{}, nil
			},
		}

		echoRouter.POST("/login", controller.Login())
		echoRouter.ServeHTTP(rec, req)
		assert.Len(t, rec.Result().Cookies(), 2, "login session and known_device cookie expected")

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

func TestUserController_Create(t *testing.T) {
	t.Parallel()

	echoRouter := newTestRouter()

	t.Run("redirect if already logged in", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		c := echoRouter.NewContext(req, rec)
		c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), auth.CtxAuthLoggedIn, true)))

		if assert.NoError(t, web.UserController{}.Create()(c)) {
			assert.Equal(t, http.StatusSeeOther, rec.Code)
		}
	})
}

// --- --- --- test data --- --- ---

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

// FIXME the param &remember_me=true is only there because of the bug in https://github.com/gorilla/sessions/issues/267
func loginPostPayload() io.Reader {
	// is a function, so each caller is it's own reader, so that it does not get drained, if it was read already
	return strings.NewReader("login=1337&password=12345678&remember_me=true")
}
