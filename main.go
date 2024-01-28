package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-arrower/skeleton/shared/interfaces/web"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	admin_init "github.com/go-arrower/skeleton/contexts/admin/init"
	"github.com/go-arrower/skeleton/contexts/auth"
	auth_init "github.com/go-arrower/skeleton/contexts/auth/init"
	"github.com/go-arrower/skeleton/shared/infrastructure"
)

func main() {
	ctx := context.Background()

	di, shutdown, err := infrastructure.InitialiseDefaultArrowerDependencies(ctx,
		&infrastructure.Config{
			OrganisationName: "arrower",
			ApplicationName:  "skeleton",
			Debug:            true,
			Postgres: infrastructure.Postgres{
				User:     "arrower",
				Password: "secret",
				Database: "arrower",
				Host:     "localhost",
				Port:     5432, //nolint:gomnd
				MaxConns: 100,  //nolint:gomnd
			},
			Web: infrastructure.Web{
				Secret:             []byte("secret"),
				Port:               8080,
				StatusEndpoint:     true,
				StatusEndpointPort: 2223,
			},
		})
	if err != nil {
		panic(err)
	}

	//alog.Unwrap(di.Logger).SetLevel(slog.LevelDebug)
	//alog.Unwrap(di.Logger).SetLevel(alog.LevelDebug)

	//
	// load and initialise optional contexts provided by arrower
	adminContext, _ := admin_init.NewAdminContext(di)
	sAPI, _ := adminContext.SettingsAPI(ctx)
	di.SettingsService = sAPI
	authContext, _ := auth_init.NewAuthContext(di)

	//
	// example route for a simple one-file setup
	di.WebRouter.GET("/", func(c echo.Context) error {
		sess, err := session.Get(auth.SessionName, c)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		userID := "World"
		if id, ok := sess.Values[auth.SessKeyUserID].(string); ok {
			userID = id
		}

		flashes := sess.Flashes()

		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		presenter := web.NewDefaultPresenter(sAPI)
		p, _ := presenter.MapDefaultBasePage(c.Request().Context(), "", map[string]interface{}{
			"userID": userID,
		})
		p["Flashes"] = flashes
		p["UserID"] = userID

		return c.Render(http.StatusOK, "=>home", p)
		//return c.Render(http.StatusOK, "=>home", echo.Map{
		//	"Flashes": flashes,
		//	"userID":  userID,
		//})
	})

	//
	// start app
	di.WebRouter.Logger.Fatal(di.WebRouter.Start(fmt.Sprintf(":%d", di.Config.Web.Port)))

	//
	// shutdown app
	// todo implement graceful shutdown ect
	_ = shutdown(ctx)
	_ = authContext.Shutdown(ctx)
	_ = adminContext.Shutdown(ctx)
}
