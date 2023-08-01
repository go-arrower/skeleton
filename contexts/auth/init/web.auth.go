package init

import (
	"net/http"

	"github.com/go-arrower/skeleton/contexts/auth"

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

	router.GET("/login", c.userController.Login()).Name = auth.RouteLogin
	router.POST("/login", c.userController.Login())
	router.GET("/logout", c.userController.Logout()).Name = auth.RouteLogout // todo make POST to prevent CSRF
	router.GET("/register", c.userController.Create())
	router.POST("/register", c.userController.Store())
	router.GET("/profile", nil) // todo protect route, so only logged in user can view

	return nil
}
