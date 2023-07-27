package web

import (
	"net/http"

	"github.com/labstack/echo/v4"

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

type TenantController struct {
	Queries *models.Queries
}

func (tc TenantController) List() func(echo.Context) error {
	return func(c echo.Context) error {
		t, _ := tc.Queries.AllTenants(c.Request().Context())

		return c.Render(http.StatusOK, "=>auth.tenants", t) //nolint:wrapcheck
	}
}

func (tc TenantController) Create() func(echo.Context) error {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "auth=>auth.register.tenant", nil) //nolint:wrapcheck
	}
}

func (tc TenantController) Store() func(echo.Context) error {
	return func(c echo.Context) error {
		name := c.FormValue("name")

		_ = tc.Queries.CreateTenant(c.Request().Context(), name)

		return c.Redirect(http.StatusSeeOther, "/") //nolint:wrapcheck
	}
}
