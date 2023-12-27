package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/go-arrower/arrower/alog"

	"github.com/labstack/echo-contrib/echoprometheus"

	"github.com/go-arrower/skeleton/shared/interfaces/web"

	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/postgres"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	admin_init "github.com/go-arrower/skeleton/contexts/admin/init"
	"github.com/go-arrower/skeleton/contexts/auth"
	auth_init "github.com/go-arrower/skeleton/contexts/auth/init"
	"github.com/go-arrower/skeleton/shared/infrastructure"
	"github.com/go-arrower/skeleton/shared/infrastructure/template"
)

func main() {
	ctx := context.Background()

	// dependencies
	di := &infrastructure.Container{Config: &infrastructure.Config{ApplicationName: "arrower skeleton"}}

	di.Logger, di.MeterProvider, di.TraceProvider = setupTelemetry(ctx)
	alog.Unwrap(di.Logger).SetLevel(slog.LevelDebug)
	alog.Unwrap(di.Logger).SetLevel(alog.LevelDebug)

	pg, err := postgres.ConnectAndMigrate(ctx, postgres.Config{
		User:       "arrower",
		Password:   "secret",
		Database:   "arrower",
		Host:       "localhost",
		Port:       5432, //nolint:gomnd
		MaxConns:   100,  //nolint:gomnd
		Migrations: postgres.ArrowerDefaultMigrations,
	}, di.TraceProvider)
	if err != nil {
		panic(err)
	}

	di.DB = pg.PGx

	di.Logger = alog.New(
		alog.WithLevel(slog.LevelDebug),
		alog.WithHandler(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource:   true,
			Level:       nil, // this level is ignored, ArrowerLogger's level is used for all handlers.
			ReplaceAttr: alog.MapLogLevelsToName,
		})),
		alog.WithHandler(alog.NewLokiHandler(nil)),
		alog.WithHandler(alog.NewPostgresHandler(di.DB, nil)),
	)
	alog.Unwrap(di.Logger).SetLevel(alog.LevelDebug)

	router := echo.New()
	router.Debug = true // todo only in dev mode
	router.HideBanner = true
	router.Logger.SetOutput(io.Discard)
	router.Validator = &CustomValidator{validator: validator.New()}
	router.IPExtractor = echo.ExtractIPFromXFFHeader()                                                   // see: https://echo.labstack.com/docs/ip-address
	router.Use(otelecho.Middleware("www.servername.tld", otelecho.WithTracerProvider(di.TraceProvider))) // todo set servername
	router.Use(echoprometheus.NewMiddleware("skeleton"))                                                 //todo set name
	router.Use(middleware.Static("public"))
	router.Use(injectMW)

	di.APIRouter = router.Group("/api") // todo add api middleware

	// router.Use(session.Middleware())
	ss, _ := auth.NewPGSessionStore(pg.PGx, []byte("secret")) // todo use secure key
	di.WebRouter = router
	di.WebRouter.Use(session.Middleware(ss))
	// di.WebRouter.Use(middleware.CSRF())
	di.WebRouter.Use(auth.EnrichCtxWithUserInfoMiddleware)

	di.AdminRouter = di.WebRouter.Group("/admin")
	di.AdminRouter.Use(auth.EnsureUserIsSuperuserMiddleware)

	ip := getOutboundIP()

	queue, _ := jobs.NewPostgresJobs(di.Logger, di.MeterProvider, di.TraceProvider, pg.PGx,
		jobs.WithPoolName(ip),
	)
	arrowerQueue, _ := jobs.NewPostgresJobs(di.Logger, di.MeterProvider, di.TraceProvider, pg.PGx,
		jobs.WithQueue("arrower"),
		jobs.WithPoolName(ip),
	)
	di.DefaultQueue = queue
	di.ArrowerQueue = arrowerQueue

	r, _ := template.NewRenderer(di.Logger, di.TraceProvider, os.DirFS("shared/interfaces/web/views"), true)
	router.Renderer = r

	adminContext, _ := admin_init.NewAdminContext(di)
	sAPI, _ := adminContext.SettingsAPI(ctx)
	di.SettingsService = sAPI
	authContext, _ := auth_init.NewAuthContext(di)

	presenter := web.NewDefaultPresenter(sAPI)

	router.GET("/", func(c echo.Context) error {
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

	router.Logger.Fatal(router.Start(":8080"))

	_ = authContext.Shutdown(ctx)
	_ = queue.Shutdown(ctx)
	_ = arrowerQueue.Shutdown(ctx)
	_ = di.TraceProvider.Shutdown(ctx)
	_ = di.MeterProvider.Shutdown(ctx)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err //nolint:wrapcheck // return the original validate error to not break the API for the caller.

		// Optionally, you could return the error to give each route more control over the status code
		// return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}

func injectMW(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := next(c); err != nil {
			c.Error(err)
		}

		// skip htmx requests, as the code is already present on the page from a previous load
		if c.Request().Header.Get("HX-Request") == "true" {
			return nil
		}

		if strings.Contains(c.Response().Header().Get("Content-Type"), "text/html") {
			_, _ = c.Response().Write([]byte(hotReloadJSCode))
		}

		return nil
	}
}

// Get preferred outbound ip of this machine.
//
// Actually it does not establish any connection and the destination does not need to be existed at all :)
// So, what the code does actually, is to get the local up address if it would connect to that target,
// you can change to any other IP address you want. conn.LocalAddr().String() is the local ip and port.
// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func getOutboundIP() string {
	conn, err := net.Dial("udp", "5.1.66.255:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

//nolint:lll
const hotReloadJSCode = `<!-- Code injected by hot-reload middleware -->
<div id="arrower-status" style="position:absolute; bottom:0; right:0; display:flex; flex-direction:column; align-items:flex-end; margin:10px;">
	<svg style="width:75px;" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#fa4901" class="w-6 h-6">
  		<path fill-rule="evenodd" d="M12.963 2.286a.75.75 0 00-1.071-.136 9.742 9.742 0 00-3.539 6.177A7.547 7.547 0 016.648 6.61a.75.75 0 00-1.152-.082A9 9 0 1015.68 4.534a7.46 7.46 0 01-2.717-2.248zM15.75 14.25a3.75 3.75 0 11-7.313-1.172c.628.465 1.35.81 2.133 1a5.99 5.99 0 011.925-3.545 3.75 3.75 0 013.255 3.717z" clip-rule="evenodd" />
	</svg>
	<span>Arrower not active!<span>
</div>

<script>
	function refreshCSS() {
		var sheets = [].slice.call(document.getElementsByTagName("link"));
		var head = document.getElementsByTagName("head")[0];
		for (var i = 0; i < sheets.length; ++i) {
			var elem = sheets[i];
			head.removeChild(elem);
			var rel = elem.rel;
			if (elem.href && typeof rel != "string" || rel.length == 0 || rel.toLowerCase() == "stylesheet") {
				var url = elem.href.replace(/(&|\?)_cacheOverride=\d+/, '');
				elem.href = url + (url.indexOf('?') >= 0 ? '&' : '?') + '_cacheOverride=' + (new Date().valueOf());
			}
			head.appendChild(elem);
		}
	}

    var loc = window.location;
    var uri = 'ws:';

    if (loc.protocol === 'https:') {
        uri = 'wss:';
    }
    //uri += '//' + loc.host;
    uri += loc.hostname +':3030/' + 'ws';

    console.log("connect to:", uri) // => todo build uri that connects to arrower cli instead of developer's app
    ws = new WebSocket(uri)

    ws.onopen = function() {
        console.log('Connected')
		setTimeout(function() {document.getElementById('arrower-status').style.visibility='hidden'}, 200);
    }

    ws.onmessage = function(msg) {
		console.log("RECEIVED RELOAD", msg.data);
        if (msg.data === 'reload') {
			window.location.reload();
		}
		else if (msg.data == 'refreshCSS') {
			refreshCSS();
		}
    }

    ws.onclose = msg => {
        console.log("Client closed connection")

		document.getElementById('arrower-status').style.visibility='visible'
        // setTimeout(window.location.reload.bind(window.location), 400);
    }

    ws.onerror = error => {
        console.log("Socket error: ", error)
    }
</script>
`
