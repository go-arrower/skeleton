package init

import (
	"github.com/labstack/echo/v4"

	"github.com/go-arrower/skeleton/contexts/auth"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/web"
)

// registerAdminRoutes initialises all admin routes of this Context. To access the user has to have admin permissions.
// The admin routes work best in combination with the Admin Context initialised.
func (c *AuthContext) registerAdminRoutes(router *echo.Group, di localDI) {
	sCont := web.SuperUserController{Queries: di.queries}

	router.Use(auth.EnsureUserIsSuperuserMiddleware)

	router.GET("/as_user/:userID", sCont.AdminLoginAsUser())
	router.GET("/leave_user", sCont.AdminLeaveUser())

	router.GET("/settings", c.settingsController.List())

	router.GET("/users", c.userController.List())
	router.GET("/users/create", c.userController.Create())
	router.POST("/users", c.userController.Register())
	router.GET("/users/:userID", c.userController.Show())
	router.GET("/users/:userID/sessions/:sessionKey", c.userController.DestroySession(di.queries))
}
