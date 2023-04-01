package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (cont JobsController) JobsHome() func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "jobs.home", nil) //nolint:wrapcheck
	}
}

func (cont JobsController) JobsQueue() func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.Render(http.StatusOK, "jobs.queue", nil) //nolint:wrapcheck
	}
}
