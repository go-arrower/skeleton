package web

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application"

	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"

	"github.com/labstack/echo/v4"
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
	Queries *models.Queries
}

func (tc UserController) List() func(echo.Context) error {
	return func(c echo.Context) error {
		u, _ := tc.Queries.AllUsers(c.Request().Context())

		return c.Render(http.StatusOK, "=>auth.users", u) //nolint:wrapcheck
	}
}

func (tc UserController) Create() func(echo.Context) error {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "auth=>auth.user.create", nil) //nolint:wrapcheck
	}
}

func (tc UserController) Store() func(echo.Context) error {
	type registerCredentials struct {
		Login                string `form:"login" validate:"required,email"`
		Password             string `form:"password" validate:"min=8,alphanumunicode"`
		PasswordConfirmation string `form:"password_confirmation" validate:"eqfield=Password"`
	}

	registerUser := application.Validate(nil, application.RegisterUser(tc.Queries))

	return func(c echo.Context) error {
		newUser := registerCredentials{}

		if err := c.Bind(&newUser); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		_, err := registerUser(c.Request().Context(), application.RegisterUserRequest{
			RegisterEmail:        newUser.Login,
			Password:             newUser.Password,
			PasswordConfirmation: newUser.PasswordConfirmation,
		})
		if err != nil {
			var valErrs = make(map[string]string)
			var validationErrors validator.ValidationErrors

			if _, ok := err.(validator.ValidationErrors); !ok {
				valErrs["Login"] = "Invalid user name"
			} else {
				validationErrors = err.(validator.ValidationErrors)
			}

			for _, e := range validationErrors {
				valErrs[e.StructField()] = e.Translate(nil)
			}

			return c.Render(http.StatusOK, "auth=>auth.user.create", map[string]any{
				"Errors": valErrs,
				"Login":  newUser.Login,
			})
		}

		return c.Redirect(http.StatusSeeOther, "/admin/auth/users")
	}
}
