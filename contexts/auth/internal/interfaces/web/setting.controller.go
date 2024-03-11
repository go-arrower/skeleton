package web

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
	"github.com/go-arrower/skeleton/shared/interfaces/web"
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

func NewSettingsController(presenter *web.DefaultPresenter, queries *models.Queries) *SettingsController {
	return &SettingsController{p: presenter, queries: queries}
}

type SettingsController struct {
	p *web.DefaultPresenter

	queries *models.Queries
}

func (sc SettingsController) List() func(echo.Context) error {
	return func(c echo.Context) error {
		page, _ := sc.p.MapDefaultBasePage(c.Request().Context(), "")
		return c.Render(http.StatusOK, "=>auth.settings", page)
	}
}
