package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/auth"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
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

func NewUserController(secret []byte) UserController {
	return UserController{knownDeviceKeyPairs: securecookie.CodecsFromPairs(secret)} //nolint:exhaustruct
}

type UserController struct {
	Queries *models.Queries

	CmdLoginUser    func(context.Context, application.LoginUserRequest) (application.LoginUserResponse, error)
	CmdRegisterUser func(context.Context, application.RegisterUserRequest) (application.RegisterUserResponse, error)
	CmdShowUserUser func(context.Context, application.ShowUserRequest) (application.ShowUserResponse, error)

	knownDeviceKeyPairs []securecookie.Codec
}

func (uc UserController) Login() func(echo.Context) error {
	type loginCredentials struct {
		application.LoginUserRequest
		RememberMe bool `form:"remember_me"`
	}

	return func(c echo.Context) error {
		if auth.IsLoggedIn(c.Request().Context()) {
			return c.Redirect(http.StatusSeeOther, "/")
		}

		if c.Request().Method == http.MethodGet {
			return c.Render(http.StatusOK, "auth=>auth.login", nil)
		}

		// POST: Login

		sess, err := session.Get(auth.SessionName, c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		loginUser := loginCredentials{ //nolint:exhaustruct // other values will be set with bind below
			LoginUserRequest: application.LoginUserRequest{
				IP:          c.RealIP(), // see: https://echo.labstack.com/docs/ip-address
				UserAgent:   c.Request().UserAgent(),
				SessionKey:  sess.ID,
				IsNewDevice: isUnknownDevice(uc.knownDeviceKeyPairs, c),
			},
		}
		if err = c.Bind(&loginUser); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		response, err := uc.CmdLoginUser(c.Request().Context(), loginUser.LoginUserRequest)
		if err != nil {
			valErrs := make(map[string]string)

			var validationErrors validator.ValidationErrors

			if !errors.As(err, &validationErrors) {
				valErrs["LoginEmail"] = "Invalid user name"
			}

			for _, e := range validationErrors {
				valErrs[e.StructField()] = e.Translate(nil)
			}

			return c.Render(http.StatusOK, "auth=>auth.login", map[string]any{
				"Errors":     valErrs,
				"LoginEmail": loginUser.LoginEmail,
			})
		}

		sess.AddFlash("Login successful")

		maxAge := 0 // session cookie => browser should delete the cookie when it closes

		if loginUser.RememberMe {
			const oneMonth = 60 * 60 * 24 * 30 //  60 sec * 60 min * 24 hours * 30 day
			maxAge = oneMonth
		}

		sess.Options = &sessions.Options{
			Path:     "/",
			Domain:   "",
			MaxAge:   maxAge,
			Secure:   false,
			HttpOnly: true,
			// cookies will not be sent, if the request originates from a third party, to prevent CSRF
			SameSite: http.SameSiteStrictMode,
		}
		sess.Values[auth.SessKeyLoggedIn] = true
		sess.Values[auth.SessKeyUserID] = string(response.User.ID)
		sess.Values[auth.SessKeyIsSuperuser] = response.User.IsSuperuser()

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = setKnownDeviceCookie(uc.knownDeviceKeyPairs, c) // set the Cookie always to renew the MaxAge
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}
}

func setKnownDeviceCookie(knownDeviceKeyPairs []securecookie.Codec, c echo.Context) error {
	encoded, err := securecookie.EncodeMulti(
		"arrower.auth.known_device",
		map[string]bool{"known_device": true},
		knownDeviceKeyPairs...,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	const twentyYears = 60 * 60 * 24 * 365 * 20

	http.SetCookie(c.Response(), sessions.NewCookie("arrower.auth.known_device", encoded, &sessions.Options{
		Path:     "/auth",
		Domain:   "",
		MaxAge:   twentyYears, // chromium has 400 days max: https://developer.chrome.com/blog/cookie-max-age-expires/
		Secure:   false,
		HttpOnly: true,
		// cookies will not be sent, if the request originates from a third party, to prevent CSRF
		SameSite: http.SameSiteStrictMode,
	}))

	return nil
}

// isUnknownDevice checks if this device is already known, as in has successfully logged in, and is unknown otherwise.
func isUnknownDevice(knownDeviceKeyPairs []securecookie.Codec, c echo.Context) bool {
	for _, cookie := range c.Request().Cookies() {
		if cookie.Name == "arrower.auth.known_device" {
			val := map[string]bool{}

			err := securecookie.DecodeMulti("arrower.auth.known_device", cookie.Value, &val, knownDeviceKeyPairs...)
			if err == nil && val["known_device"] {
				return false
			}
		}
	}

	return true
}

func (uc UserController) Logout() func(echo.Context) error {
	return func(c echo.Context) error {
		if !auth.IsLoggedIn(c.Request().Context()) {
			return c.Redirect(http.StatusSeeOther, "/")
		}

		sess, err := session.Get(auth.SessionName, c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		delete(sess.Values, auth.SessKeyLoggedIn)
		delete(sess.Values, auth.SessKeyUserID)
		delete(sess.Values, auth.SessKeyIsSuperuser)

		sess.Options = &sessions.Options{ //nolint:exhaustruct // not all options are required, as the cookie will be deleted.
			Path:   "/",
			MaxAge: -1, // delete cookie immediately
		}

		// sess.AddFlash("Logout successful")

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

		return c.Render(http.StatusOK, "=>auth.users", echo.Map{
			"users": u,
		})
	}
}

func (uc UserController) Create() func(echo.Context) error {
	return func(c echo.Context) error {
		if auth.IsLoggedIn(c.Request().Context()) {
			return c.Redirect(http.StatusSeeOther, "/")
		}

		return c.Render(http.StatusOK, "auth=>auth.user.create", nil)
	}
}

func (uc UserController) Register() func(echo.Context) error {
	return func(c echo.Context) error {
		if auth.IsLoggedIn(c.Request().Context()) {
			return c.Redirect(http.StatusSeeOther, "/")
		}

		sess, err := session.Get(auth.SessionName, c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		newUser := application.RegisterUserRequest{ //nolint:exhaustruct // other values will be set with bind below
			IP:         c.RealIP(), // see: https://echo.labstack.com/docs/ip-address
			UserAgent:  c.Request().UserAgent(),
			SessionKey: sess.ID,
		}

		if err = c.Bind(&newUser); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		response, err := uc.CmdRegisterUser(c.Request().Context(), newUser)
		if err != nil {
			valErrs := make(map[string]string)

			var validationErrors validator.ValidationErrors

			if !errors.As(err, &validationErrors) {
				valErrs["RegisterEmail"] = "Invalid user name"
			}

			for _, e := range validationErrors {
				valErrs[e.StructField()] = e.Translate(nil)
			}

			return c.Render(http.StatusOK, "auth=>auth.user.create", map[string]any{
				"Errors":        valErrs,
				"RegisterEmail": newUser.RegisterEmail,
			})
		}

		sess.Options = &sessions.Options{
			Path:     "/",
			Domain:   "",
			MaxAge:   0, // only until browser closes, as the account is not verified yet
			Secure:   false,
			HttpOnly: true,
			// cookies will not be sent, if the request originates from a third party, to prevent CSRF
			SameSite: http.SameSiteStrictMode,
		}
		sess.Values[auth.SessKeyLoggedIn] = true
		sess.Values[auth.SessKeyUserID] = string(response.User.ID)

		sess.AddFlash("Register successful")

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = setKnownDeviceCookie(uc.knownDeviceKeyPairs, c)
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, "/admin/auth/users")
	}
}

func (uc UserController) Show() func(echo.Context) error {
	return func(c echo.Context) error {
		userID := c.Param("userID")

		res, err := uc.CmdShowUserUser(c.Request().Context(), application.ShowUserRequest{UserID: user.ID(userID)})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.Render(http.StatusOK, "=>auth.user.show", echo.Map{
			"User": res.User,
		})
	}
}

func (uc UserController) DestroySession(queries *models.Queries) func(echo.Context) error {
	return func(c echo.Context) error {
		userID := c.Param("userID")
		sessionID := c.Param("sessionKey")

		err := queries.DeleteSessionByKey(c.Request().Context(), []byte(sessionID))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.Redirect(http.StatusSeeOther, "/admin/auth/users/"+userID)
	}
}
