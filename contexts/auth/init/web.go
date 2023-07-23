package init

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// registerWebRoutes initialises all routes of this Context.
func (c *AuthContext) registerWebRoutes(router *echo.Group) error {
	router.GET("/", func(c echo.Context) error {
		c.String(http.StatusOK, "HELLO FROM AUTH WEB")

		return nil
	})

	router.GET("", func(c echo.Context) error {
		c.String(http.StatusOK, "HELLO FROM AUTH WEB")

		return nil
	})

	router.GET("/register/tenant", c.tenantController.Create())
	router.POST("/register/tenant", c.tenantController.Store())

	return nil
}
