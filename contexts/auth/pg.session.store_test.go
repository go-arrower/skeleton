//go:build integration

package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/google/uuid"

	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-arrower/arrower/postgres"
	"github.com/go-arrower/arrower/tests"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/go-arrower/skeleton/contexts/auth"
)

var (
	ctx       = context.Background()
	pgHandler *postgres.Handler
)

func TestMain(m *testing.M) {
	handler, cleanup := tests.GetDBConnectionForIntegrationTesting(ctx)
	pgHandler = handler

	//
	// Run tests
	code := m.Run()

	//
	// Cleanup
	_ = handler.Shutdown(ctx)
	_ = cleanup()

	os.Exit(code)
}

func TestNewPGSessionStore(t *testing.T) {
	t.Parallel()

	t.Run("create fails", func(t *testing.T) {
		t.Parallel()

		ss, err := auth.NewPGSessionStore(nil, keyPairs)
		assert.Error(t, err)
		assert.Empty(t, ss)
	})

	t.Run("create succeeds", func(t *testing.T) {
		t.Parallel()

		pg := tests.PrepareTestDatabase(pgHandler).PGx

		ss, err := auth.NewPGSessionStore(pg, keyPairs)
		assert.NoError(t, err)
		assert.NotEmpty(t, ss)

		assert.NotEmpty(t, ss.Codecs)
		assert.NotEmpty(t, ss.Options)
	})
}

func TestNewPGSessionStore_HTTPRequest(t *testing.T) {
	t.Parallel()

	pg := tests.PrepareTestDatabase(pgHandler).PGx
	echoRouter := newTestRouter(pg)

	var cookie *http.Cookie // the cookie to use over all requests

	t.Run("set initial cookie when surfing the site", func(t *testing.T) { //nolint:parallel // the tests depend on each other and the order is important
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		echoRouter.ServeHTTP(rec, req)

		result := rec.Result()
		defer result.Body.Close()
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Body.String())

		cookie = rec.Result().Cookies()[0] // safe cookie for reuse later on

		// assert cookie
		assert.Len(t, result.Cookies(), 1)
		assert.Equal(t, "/", result.Cookies()[0].Path)
		assert.Equal(t, "session", result.Cookies()[0].Name)
		assert.Equal(t, http.SameSiteStrictMode, result.Cookies()[0].SameSite)
		assert.Equal(t, 86400*30, result.Cookies()[0].MaxAge)

		// assert db entry
		queries := models.New(pg)
		sessions, _ := queries.AllSessions(ctx)

		assert.Len(t, sessions, 1)
		// cookie and session expire at the same time, allow 1 second of diff to make sure different granulates
		// in the representation like nanoseconds in pg are not an issue.
		assert.True(t, result.Cookies()[0].Expires.Sub(sessions[0].ExpiresAt.Time) < 1)
		assert.Empty(t, sessions[0].UserID)
	})

	t.Run("session already exists => user logs in", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		req.AddCookie(cookie) // use the cookie / session from the call before
		rec := httptest.NewRecorder()
		echoRouter.ServeHTTP(rec, req)

		result := rec.Result()
		defer result.Body.Close()
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Body.String())
		assert.Len(t, result.Cookies(), 1)

		// assert db entry
		queries := models.New(pg)
		sessions, _ := queries.AllSessions(ctx)

		assert.Len(t, sessions, 1)
		assert.NotEmpty(t, sessions[0].UserID)
		assert.Equal(t, userID, sessions[0].UserID.UUID)
	})

	t.Run("destroy session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/destroy", nil)
		req.AddCookie(cookie) // use the cookie / session from the call before
		rec := httptest.NewRecorder()
		echoRouter.ServeHTTP(rec, req)

		// assert cookie
		result := rec.Result()
		defer result.Body.Close()
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Body.String())
		assert.Len(t, result.Cookies(), 1)
		assert.Equal(t, "/", result.Cookies()[0].Path)
		assert.Equal(t, "session", result.Cookies()[0].Name)
		assert.Equal(t, -1, result.Cookies()[0].MaxAge)

		// assert db entry
		queries := models.New(pg)
		sessions, _ := queries.AllSessions(ctx)

		assert.Len(t, sessions, 0)
	})
}

// --- --- --- TEST DATA --- --- ---

var (
	keyPairs = []byte("secret")
	userID   = uuid.New()
)

func newTestRouter(pg *pgxpool.Pool) *echo.Echo {
	ss, _ := auth.NewPGSessionStore(pg, keyPairs)
	echoRouter := echo.New()
	echoRouter.Use(session.Middleware(ss))

	echoRouter.GET("/", func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		sess.Values["some-session"] = "some-value"

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.NoContent(http.StatusOK)
	})

	echoRouter.GET("/login", func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		sess.Values[auth.SessKeyUserID] = userID.String()

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.NoContent(http.StatusOK)
	})

	echoRouter.GET("/destroy", func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		delete(sess.Values, auth.SessKeyUserID)

		sess.Options = &sessions.Options{
			Path:   "/",
			MaxAge: -1, // delete cookie immediately
		}

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.NoContent(http.StatusOK)
	})

	// seed db with example user
	_, _ = pg.Exec(ctx, `INSERT INTO auth.user (id, login) VALUES ($1, $2);`, userID, "login")

	return echoRouter
}
