package web

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/shared/application"
	"github.com/go-arrower/skeleton/shared/views/pages"
)

func NewHelloController(app application.App) *HelloController {
	return &HelloController{app: app}
}

type HelloController struct {
	app application.App
}

func (hc *HelloController) SayHello() func(c echo.Context) error {
	return func(c echo.Context) error {
		name := c.Param("name")

		res, err := hc.app.SayHello.H(
			c.Request().Context(),
			application.SayHelloRequest{Name: name},
		)
		if err != nil {
			return c.String(http.StatusBadRequest, "ERROR")
		}

		return c.Render(http.StatusOK, "hello", echo.Map{
			"Member": pages.PresentHello(res),
		})
	}
}
