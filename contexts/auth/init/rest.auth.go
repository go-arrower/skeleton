package init

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// registerAPIRoutes initialises all api routes of this Context. API routes require a valid auth.APIKey.
// It is best practise to version your API.
func (c *AuthContext) registerAPIRoutes(router *echo.Group) error {
	router = router.Group(fmt.Sprintf("/v1/%s", contextName))

	router.GET("/", func(c echo.Context) error {
		c.String(http.StatusOK, "HELLO FROM AUTH API")

		return nil
	})

	return nil
}
