package init

import (
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/auth"
)

// registerWebRoutes initialises all routes of this Context.
func (c *AuthContext) registerWebRoutes(router *echo.Group) {
	router.GET("/login", c.userController.Login()).Name = auth.RouteLogin
	router.POST("/login", c.userController.Login())
	router.GET("/logout", c.userController.Logout()).Name = auth.RouteLogout // todo make POST to prevent CSRF
	router.GET("/register", c.userController.Create())
	router.POST("/register", c.userController.Register())

	router.GET("/profile", nil, auth.EnsureUserIsLoggedInMiddleware)
}
