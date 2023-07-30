package web

import (
	"context"
	"net/http"

	"github.com/go-arrower/arrower/mw"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/auth"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

/*
Proposal for naming conventions:
	- index (list)
	- create (new)
	- store (new)
	- show
	- edit
	- update
	- delete
*/

type UserController struct {
	Queries      *models.Queries
	CmdLoginUser func(context.Context, application.LoginUserRequest) (application.LoginUserResponse, error)
}

func (uc UserController) Login() func(echo.Context) error {
	return func(c echo.Context) error {
		if auth.IsLoggedIn(c.Request().Context()) {
			return c.Redirect(http.StatusSeeOther, "/") //nolint:wrapcheck
		}

		if c.Request().Method == http.MethodGet {
			return c.Render(http.StatusOK, "auth=>auth.login", nil) //nolint:wrapcheck
		}

		loginUser := application.LoginUserRequest{} //nolint:exhaustruct
		if err := c.Bind(&loginUser); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		response, err := uc.CmdLoginUser(c.Request().Context(), application.LoginUserRequest{
			LoginEmail: loginUser.LoginEmail,
			Password:   loginUser.Password,
		})
		if err != nil {
			valErrs := make(map[string]string)

			var validationErrors validator.ValidationErrors

			if _, ok := err.(validator.ValidationErrors); !ok {
				valErrs["Login"] = "Invalid user name"
			} else {
				validationErrors = err.(validator.ValidationErrors)
			}

			for _, e := range validationErrors {
				valErrs[e.StructField()] = e.Translate(nil)
			}

			return c.Render(http.StatusOK, "auth=>auth.login", map[string]any{ //nolint:wrapcheck
				"Errors": valErrs,
				"Login":  loginUser.LoginEmail,
			})
		}

		sess, err := session.Get("session", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		sess.Options = &sessions.Options{
			Path:     "/",
			Domain:   "",
			MaxAge:   7 * 24 * 60 * 60, // 7 days * 24 hours * 60 min * 60 sec
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode, // cookies will not be sent, if the request originates from a third party, to prevent CSRF
		}
		sess.Values[auth.SessKeyLoggedIn] = true
		sess.Values[auth.SessKeyUserID] = string(response.User.ID)

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.Redirect(http.StatusSeeOther, "/") //nolint:wrapcheck
	}
}

func (uc UserController) Logout() func(echo.Context) error {
	return func(c echo.Context) error {
		if !auth.IsLoggedIn(c.Request().Context()) {
			return c.Redirect(http.StatusSeeOther, "/") //nolint:wrapcheck
		}

		sess, err := session.Get("session", c) // todo extract "session" as variable & rename arrower.auth
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		delete(sess.Values, auth.SessKeyLoggedIn)
		delete(sess.Values, auth.SessKeyUserID)

		sess.Options = &sessions.Options{
			Path:   "/",
			MaxAge: -1, // delete cookie immediately
		}

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}
}

func (uc UserController) List() func(echo.Context) error {
	return func(c echo.Context) error {
		u, _ := uc.Queries.AllUsers(c.Request().Context())

		return c.Render(http.StatusOK, "=>auth.users", u) //nolint:wrapcheck
	}
}

func (uc UserController) Create() func(echo.Context) error {
	return func(c echo.Context) error {
		if auth.IsLoggedIn(c.Request().Context()) {
			return c.Redirect(http.StatusSeeOther, "/")
		}

		return c.Render(http.StatusOK, "auth=>auth.user.create", nil) //nolint:wrapcheck
	}
}

func (uc UserController) Store() func(echo.Context) error {
	registerUser := mw.Validate(nil, application.RegisterUser(uc.Queries))

	return func(c echo.Context) error {
		newUser := application.RegisterUserRequest{} //nolint:exhaustruct

		if err := c.Bind(&newUser); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		_, err := registerUser(c.Request().Context(), newUser)
		if err != nil {
			valErrs := make(map[string]string)

			var validationErrors validator.ValidationErrors

			if _, ok := err.(validator.ValidationErrors); !ok {
				valErrs["RegisterEmail"] = "Invalid user name"
			} else {
				validationErrors = err.(validator.ValidationErrors)
			}

			for _, e := range validationErrors {
				valErrs[e.StructField()] = e.Translate(nil)
			}

			return c.Render(http.StatusOK, "auth=>auth.user.create", map[string]any{ //nolint:wrapcheck
				"Errors":        valErrs,
				"RegisterEmail": newUser.RegisterEmail,
			})
		}

		return c.Redirect(http.StatusSeeOther, "/admin/auth/users") //nolint:wrapcheck
	}
}
