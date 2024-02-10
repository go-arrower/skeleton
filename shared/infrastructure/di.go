package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4/middleware"

	"github.com/go-arrower/arrower/setting"

	"github.com/go-arrower/arrower/alog"
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/postgres"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"

	"github.com/go-arrower/skeleton/contexts/auth"
	"github.com/go-arrower/skeleton/shared/infrastructure/template"
)

var (
	ErrMissingDependency = errors.New("missing dependency")
)

// Container holds global dependencies that can be used within each Context, to make initialisation easier.
// If the Context can operate with the shared resources.
// Otherwise, the Context is advised to initialise its own dependencies from its own configuration.
type Container struct {
	Logger        alog.Logger
	MeterProvider *metric.MeterProvider
	TraceProvider *trace.TracerProvider

	Config *Config
	PGx    *pgxpool.Pool
	db     *postgres.Handler

	WebRouter   *echo.Echo
	APIRouter   *echo.Group
	AdminRouter *echo.Group

	ArrowerQueue jobs.Queue
	DefaultQueue jobs.Queue

	Settings setting.Settings
}

func (c *Container) EnsureAllDependenciesPresent() error {
	if c.Config == nil {
		return fmt.Errorf("%w: global config not found", ErrMissingDependency)
	}

	return nil
}

func InitialiseDefaultArrowerDependencies(ctx context.Context, conf *Config) (*Container, func(ctx context.Context) error, error) {
	container := &Container{
		Config: conf,
	}

	{ // observability
		//labels/tags/resources that are common to all traces and metrics.
		resource := resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(fmt.Sprintf("%s.%s", conf.OrganisationName, conf.ApplicationName)),

			// NEEDS TO MATCH WITH THE LOGS LABEL (why? for the "Logs for this span" button in tempo?)
			attribute.String(conf.OrganisationName, conf.ApplicationName),

			// more attributes like e.g. kubernetes pod name
		)
		// FIND OUT: CAN RESOURCES BE ADDED TO LOGGER SO ALL THREE HAVE THE SAME VALUES?

		{
			opts := []otlptracegrpc.Option{
				otlptracegrpc.WithEndpoint("localhost:4317"), // todo configure
			}
			if conf.Debug {
				opts = append(opts,
					otlptracegrpc.WithInsecure(),
					otlptracegrpc.WithDialOption(grpc.WithBlock()),
				)
			}

			traceExporter, err := otlptracegrpc.New(ctx, opts...)
			if err != nil {
				return nil, nil, err
			}

			traceProvider := trace.NewTracerProvider(
				trace.WithBatcher(traceExporter), // prod
				trace.WithResource(resource),
				// set the sampling rate based on the parent span to 60%
				trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.6))),
			)
			if conf.Debug {
				traceProvider = trace.NewTracerProvider(
					trace.WithSyncer(traceExporter),
					trace.WithResource(resource),
					trace.WithSampler(trace.AlwaysSample()),
				)
			}

			container.TraceProvider = traceProvider
			// otel.SetTracerProvider(traceProvider)
		}

		{
			exporter, err := prometheus.New()
			if err != nil {
				return nil, nil, err
			}

			meterProvider := metric.NewMeterProvider(
				metric.WithResource(resource),
				metric.WithReader(exporter),
			)

			container.MeterProvider = meterProvider
			// otel.SetMeterProvider(meterProvider)
		}
	}

	{ // postgres
		pg, err := postgres.ConnectAndMigrate(ctx, postgres.Config{
			User:       conf.Postgres.User,
			Password:   conf.Postgres.Password,
			Database:   conf.Postgres.Database,
			Host:       conf.Postgres.Host,
			Port:       conf.Postgres.Port,
			MaxConns:   conf.Postgres.MaxConns,
			Migrations: postgres.ArrowerDefaultMigrations,
		}, container.TraceProvider)
		if err != nil {
			return nil, nil, err
		}

		container.PGx = pg.PGx
		container.db = pg
	}

	container.Settings = setting.NewPostgresSettings(container.PGx)

	container.Logger = alog.New()
	if conf.Debug {
		container.Logger = alog.NewDevelopment(container.PGx, container.Settings)
	}
	// slog.SetDefault(container.Logger.(*slog.Logger)) // todo test if this works even if the cast works

	{ // echo router
		// todo extract echo setup to main arrower repo, ones it is "ready" and can be abstracted for easier use
		router := echo.New()
		router.HideBanner = true
		router.Logger.SetOutput(io.Discard)
		router.Validator = &CustomValidator{validator: validator.New()}
		router.IPExtractor = echo.ExtractIPFromXFFHeader() // see: https://echo.labstack.com/docs/ip-address
		router.Use(otelecho.Middleware(conf.Web.Hostname, otelecho.WithTracerProvider(container.TraceProvider)))
		router.Use(echoprometheus.NewMiddleware(conf.ApplicationName))
		router.Use(middleware.Static("public")) // todo use fs instead

		hotReload := false
		if conf.Debug {
			router.Debug = true
			router.Use(injectMW)
			hotReload = true
		}

		r, _ := template.NewRenderer(container.Logger, container.TraceProvider, os.DirFS("shared/views"), hotReload)
		router.Renderer = r

		// router.Use(session.Middleware())
		ss, _ := auth.NewPGSessionStore(container.PGx, conf.Web.Secret)
		container.WebRouter = router
		container.WebRouter.Use(session.Middleware(ss))
		// di.WebRouter.Use(middleware.CSRF())
		container.WebRouter.Use(auth.EnrichCtxWithUserInfoMiddleware)

		container.AdminRouter = container.WebRouter.Group("/admin")
		container.AdminRouter.Use(auth.EnsureUserIsSuperuserMiddleware)

		container.APIRouter = router.Group("/api") // todo add api middleware
	}

	{ // jobs
		name := conf.InstanceName
		if name == "" {
			name = getOutboundIP()
		}

		queue, err := jobs.NewPostgresJobs(container.Logger, container.MeterProvider, container.TraceProvider, container.PGx,
			jobs.WithPoolName(name),
		)
		if err != nil {
			return nil, nil, err
		}

		arrowerQueue, err := jobs.NewPostgresJobs(container.Logger, container.MeterProvider, container.TraceProvider, container.PGx,
			jobs.WithQueue("Arrower"),
			jobs.WithPoolName(name),
		)
		if err != nil {
			return nil, nil, err
		}

		container.DefaultQueue = queue
		container.ArrowerQueue = arrowerQueue
	}

	//
	// Start the prometheus HTTP server and pass the exporter Collector to it
	if conf.Web.StatusEndpoint {
		go serveMetrics(ctx, container.Logger, conf.Web.StatusEndpointPort)
	}

	return container, shutdown(container), nil
}

func shutdown(di *Container) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		di.Logger.InfoContext(ctx, "shutdown...")

		_ = di.WebRouter.Shutdown(ctx)
		_ = di.DefaultQueue.Shutdown(ctx)
		_ = di.ArrowerQueue.Shutdown(ctx)
		_ = di.TraceProvider.Shutdown(ctx)
		_ = di.MeterProvider.Shutdown(ctx)
		_ = di.db.Shutdown(ctx)
		// todo shutdown meter/status endpoint

		return nil // todo error handling
	}
}

func serveMetrics(ctx context.Context, logger alog.Logger, port int) {
	const path = "/metrics"

	addr := fmt.Sprintf(":%d", port)

	logger.InfoContext(ctx, "serving metrics",
		slog.String("addr", addr),
		slog.String("path", path),
	)

	// http.Handle("/metrics", promhttp.Handler())
	http.Handle(path, promhttp.HandlerFor(
		prometheus2.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true, // to enable Examplars in the export format
		},
	))

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		logger.DebugContext(ctx, "error serving http", slog.String("err", err.Error()))

		return
	}
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

// Get preferred outbound ip of this machine.
//
// Actually, it does not establish any connection and the destination does not need to be existed at all :)
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
