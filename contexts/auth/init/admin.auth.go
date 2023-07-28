package init

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// registerAdminRoutes initialises all admin routes of this Context. To access the user has to have admin permissions.
// The admin routes work best in combination with the Admin Context initialised.
func (c *AuthContext) registerAdminRoutes(router *echo.Group) error {
	router.GET("/", func(c echo.Context) error {
		c.String(http.StatusOK, "HELLO FROM AUTH ADMIN")

		return nil
	})

	router.GET("/tenants", c.tenantController.List())

	router.GET("/users", c.userController.List())
	router.GET("/users/create", c.userController.Create())
	router.POST("/users", c.userController.Store())

	return nil
}
