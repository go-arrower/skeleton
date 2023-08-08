package web

import (
	"context"
	"net/http"

	"github.com/gorilla/securecookie"

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

var knownDeviceKeyPairs = securecookie.CodecsFromPairs([]byte("secret"))

func (uc UserController) Login() func(echo.Context) error {
	type loginCredentials struct {
		application.LoginUserRequest
		RememberMe bool `form:"remember_me"`
	}

	return func(c echo.Context) error {
		if auth.IsLoggedIn(c.Request().Context()) {
			return c.Redirect(http.StatusSeeOther, "/") //nolint:wrapcheck
		}

		if c.Request().Method == http.MethodGet {
			return c.Render(http.StatusOK, "auth=>auth.login", nil) //nolint:wrapcheck
		}

		// POST: Login

		sess, err := session.Get("session", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		loginUser := loginCredentials{
			LoginUserRequest: application.LoginUserRequest{
				IP:          c.RealIP(), // see: https://echo.labstack.com/docs/ip-address
				UserAgent:   c.Request().UserAgent(),
				SessionKey:  sess.ID,
				IsNewDevice: isUnknownDevice(c),
			},
		} //nolint:exhaustruct
		if err := c.Bind(&loginUser); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		response, err := uc.CmdLoginUser(c.Request().Context(), loginUser.LoginUserRequest)
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
				"Errors":     valErrs,
				"LoginEmail": loginUser.LoginEmail,
			})
		}

		maxAge := 0 // session cookie => browser should delete the cookie when it closes
		if loginUser.RememberMe {
			maxAge = 60 * 60 * 24 * 30 //  60 sec * 60 min * 24 hours * 30 day
		}

		sess.Options = &sessions.Options{
			Path:     "/",
			Domain:   "",
			MaxAge:   maxAge,
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

		// set known device cookie
		encoded, err := securecookie.EncodeMulti("arrower.auth.known_device", map[string]bool{"known_device": true}, knownDeviceKeyPairs...)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		http.SetCookie(c.Response(), sessions.NewCookie("arrower.auth.known_device", encoded, &sessions.Options{
			Path:     "/auth",
			Domain:   "",
			MaxAge:   60 * 60 * 24 * 365 * 20, // 20 years, chromium has a max of 400 days, see: https://developer.chrome.com/blog/cookie-max-age-expires/
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}))

		return c.Redirect(http.StatusSeeOther, "/") //nolint:wrapcheck
	}
}

// isUnknownDevice checks if this device is already known, as in has successfully logged in, and is unknown otherwise.
func isUnknownDevice(c echo.Context) bool {
	for _, cookie := range c.Request().Cookies() {
		if cookie.Name == "arrower.auth.known_device" {
			val := map[string]bool{}
			err := securecookie.DecodeMulti("arrower.auth.known_device", cookie.Value, &val, knownDeviceKeyPairs...)
			if err == nil && val["known_device"] == true {
				return false
			}
		}
	}

	return true
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

func (uc UserController) Register() func(echo.Context) error {
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
