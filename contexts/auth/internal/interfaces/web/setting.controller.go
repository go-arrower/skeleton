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

type SettingsController struct {
	Queries *models.Queries
}

func (sc SettingsController) List() func(echo.Context) error {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "=>auth.settings", nil)
	}
}
