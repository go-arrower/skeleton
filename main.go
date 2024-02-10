package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/go-arrower/arrower/mw"

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

	arrower, shutdown, err := infrastructure.InitialiseDefaultArrowerDependencies(ctx,
		&infrastructure.Config{
			OrganisationName: "arrower",
			ApplicationName:  "skeleton",
			InstanceName:     "",
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
				Hostname:           "www.servername.tld",
				StatusEndpoint:     true,
				StatusEndpointPort: 2223,
			},
		})
	if err != nil {
		panic(err)
	}

	//err = arrower.Settings.Save(ctx, alog.SettingLogLevel, setting.NewValue(int(slog.LevelDebug)))
	//alog.Unwrap(arrower.Logger).SetLevel(slog.LevelDebug)
	//alog.Unwrap(arrower.Logger).SetLevel(alog.LevelDebug)

	//
	// load and initialise optional contexts provided by arrower
	adminContext, _ := admin_init.NewAdminContext(arrower)
	authContext, _ := auth_init.NewAuthContext(arrower)

	//
	// example route for a simple one-file setup
	arrower.WebRouter.GET("/", func(c echo.Context) error {
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

		presenter := web.NewDefaultPresenter(arrower.Settings)
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
	// initRegularExampleQueueLoad(ctx, arrower)
	arrower.WebRouter.Logger.Fatal(arrower.WebRouter.Start(fmt.Sprintf(":%d", arrower.Config.Web.Port)))

	//
	// shutdown app
	// todo implement graceful shutdown ect
	_ = shutdown(ctx)
	_ = authContext.Shutdown(ctx)
	_ = adminContext.Shutdown(ctx)
}

func initRegularExampleQueueLoad(ctx context.Context, di *infrastructure.Container) {
	type (
		SomeJob        struct{}
		NamedJob       struct{ Name string }
		LongRunningJob struct{}
	)

	_ = di.DefaultQueue.RegisterJobFunc(
		mw.TracedU(di.TraceProvider, mw.MetricU(di.MeterProvider, mw.LoggedU(di.Logger.(*slog.Logger),
			func(ctx context.Context, job SomeJob) error {
				di.Logger.InfoContext(ctx, "LOG ASYNC SIMPLE JOB")
				//panic("SOME JOB PANICS")

				time.Sleep(time.Duration(rand.Intn(10)) * time.Second) //nolint:gosec,gomnd,lll // weak numbers are ok, it is wait time

				if rand.Intn(100) > 70 { //nolint:gosec,gomndworkers,gomnd
					return errors.New("some error") //nolint:goerr113
				}

				return nil
			},
		))),
	)

	_ = di.DefaultQueue.RegisterJobFunc(
		mw.TracedU(di.TraceProvider, mw.MetricU(di.MeterProvider, mw.LoggedU(di.Logger.(*slog.Logger),
			func(ctx context.Context, job NamedJob) error {
				di.Logger.InfoContext(ctx, "named job", slog.String("name", job.Name))

				time.Sleep(time.Duration(rand.Intn(4)) * time.Second)

				return nil
			},
		))),
	)

	_ = di.DefaultQueue.RegisterJobFunc(
		mw.TracedU(di.TraceProvider, mw.MetricU(di.MeterProvider, mw.LoggedU(di.Logger.(*slog.Logger),
			func(ctx context.Context, job LongRunningJob) error {
				time.Sleep(time.Duration(rand.Intn(5)) * time.Minute) //nolint:gosec,gomnd // weak numbers are ok, it is wait time

				if rand.Intn(100) > 95 { //nolint:gosec,gomnd
					return errors.New("some error") //nolint:goerr113
				}

				return nil
			},
		))),
	)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				r := rand.Intn(100)

				if r%5 == 0 {
					_ = di.DefaultQueue.Enqueue(ctx, SomeJob{})
				}

				if r%12 == 0 {
					for i := 0; i/2 < r; i++ {
						// for i := range r { // fixme use new go1.22 style
						_ = di.DefaultQueue.Enqueue(ctx, NamedJob{Name: gofakeit.Name()})
					}
				}

				if r == 0 {
					_ = di.DefaultQueue.Enqueue(ctx, LongRunningJob{})
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
