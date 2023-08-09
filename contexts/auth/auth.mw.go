package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-arrower/arrower"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var ErrInvalidSessionValue = errors.New("invalid session value")

const (
	CtxAuthLoggedIn    arrower.CTXKey = "auth.pass"
	CtxAuthUserID      arrower.CTXKey = "auth.user_id"
	CtxAuthIsSuperuser arrower.CTXKey = "auth.superuser"
)

const (
	// FIXME: is redundant and can disappear from the session, use the existance of user_id to set the flag in the ctx middleware
	SessKeyLoggedIn    = "auth.user_is_logged_in" // FIXME don't export from the context => move internally
	SessKeyUserID      = "auth.user_id"
	SessKeyIsSuperuser = "auth.user_is_superuser"
)

// EnsureUserIsLoggedInMiddleware makes sure the routes can only be accessed by a logged-in user.
// It does set the User in the same way EnrichCtxWithUserInfoMiddleware does.
func EnsureUserIsLoggedInMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	type passed struct {
		loggedIn bool
		userId   bool
	}

	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		passed := passed{}

		if sess.Values[SessKeyLoggedIn] != nil {
			lin, ok := sess.Values[SessKeyLoggedIn].(bool)
			if !ok {
				return fmt.Errorf("could not access user_logged_in: %w", ErrInvalidSessionValue)
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), CtxAuthLoggedIn, lin)))
			passed.loggedIn = lin
		}

		if sess.Values[SessKeyUserID] != nil {
			uID, ok := sess.Values[SessKeyUserID].(string)
			if !ok {
				return fmt.Errorf("could not access user_id: %w", ErrInvalidSessionValue)
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), CtxAuthUserID, uID)))
			passed.userId = true
		}

		if passed.loggedIn && passed.userId {
			return next(c)
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}
}

func EnsureUserIsSuperuserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	type passed struct {
		loggedIn    bool
		userId      bool
		isSuperuser bool
	}

	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		passed := passed{}

		if sess.Values[SessKeyLoggedIn] != nil {
			lin, ok := sess.Values[SessKeyLoggedIn].(bool)
			if !ok {
				return fmt.Errorf("could not access user_logged_in: %w", ErrInvalidSessionValue)
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), CtxAuthLoggedIn, lin)))
			passed.loggedIn = lin
		}

		if sess.Values[SessKeyUserID] != nil {
			uID, ok := sess.Values[SessKeyUserID].(string)
			if !ok {
				return fmt.Errorf("could not access user_id: %w", ErrInvalidSessionValue)
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), CtxAuthUserID, uID)))
			passed.userId = true
		}

		if sess.Values[SessKeyIsSuperuser] != nil {
			su, ok := sess.Values[SessKeyIsSuperuser].(bool)
			if !ok {
				return fmt.Errorf("could not access user_is_superuser: %w", ErrInvalidSessionValue)
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), CtxAuthIsSuperuser, su)))
			passed.isSuperuser = su
		}

		if passed.loggedIn && passed.userId && passed.isSuperuser {
			return next(c)
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}
}

// EnrichCtxWithUserInfoMiddleware checks if a User is logged in and puts those values into the http request's context,
// so they are available in other parts of the app. For convenience use the helpers like: IsLoggedIn.
// If you want to ensure only logged-in users can access a URL use EnsureUserIsLoggedInMiddleware instead.
func EnrichCtxWithUserInfoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if sess.Values[SessKeyLoggedIn] != nil {
			lin, ok := sess.Values[SessKeyLoggedIn].(bool)
			if !ok {
				return fmt.Errorf("could not access user_logged_in: %w", ErrInvalidSessionValue)
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), CtxAuthLoggedIn, lin)))
		}

		if sess.Values[SessKeyUserID] != nil {
			uID, ok := sess.Values[SessKeyUserID].(string)
			if !ok {
				return fmt.Errorf("could not access user_id: %w", ErrInvalidSessionValue)
			}

			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), CtxAuthUserID, uID)))
		}

		if sess.Values[SessKeyIsSuperuser] != nil {
			su, ok := sess.Values[SessKeyIsSuperuser].(bool)
			if ok {
				c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), CtxAuthIsSuperuser, su)))
			}
		}

		return next(c)
	}
}

// ensure only authed user can access
func LoginRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}

// ensure only admins can access
func AuthAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}

func IsLoggedIn(ctx context.Context) bool {
	if v, ok := ctx.Value(CtxAuthLoggedIn).(bool); ok {
		return v
	}

	return false
}

func CurrentUserID(ctx context.Context) string {
	if v, ok := ctx.Value(CtxAuthUserID).(string); ok {
		return v
	}

	return ""
}

func IsSuperUser(ctx context.Context) bool {
	if v, ok := ctx.Value(CtxAuthIsSuperuser).(bool); ok {
		return v
	}

	return false
}
