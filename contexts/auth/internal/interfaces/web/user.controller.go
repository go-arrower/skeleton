package web

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-playground/validator/v10"

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

	return func(c echo.Context) error {
		newUser := registerCredentials{}

		if err := c.Bind(&newUser); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		if err := c.Validate(newUser); err != nil {
			validationErrors := err.(validator.ValidationErrors)

			var valErrs = make(map[string]string)
			for _, e := range validationErrors {
				valErrs[e.StructField()] = e.Translate(nil)
			}

			return c.Render(http.StatusOK, "auth=>auth.user.create", map[string]any{
				"Errors": valErrs,
				"Login":  newUser.Login,
			})
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = tc.Queries.CreateUser(c.Request().Context(), models.CreateUserParams{
			UserLogin:        newUser.Login,
			UserPasswordHash: string(hashed),
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.Redirect(http.StatusSeeOther, "/admin/auth/users")
	}
}
