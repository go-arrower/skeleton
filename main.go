package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	trace2 "go.opentelemetry.io/otel/trace"

	"github.com/go-arrower/arrower"
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/postgres"
	"github.com/go-arrower/skeleton/contexts/admin/startup"
	"github.com/go-arrower/skeleton/shared/infrastructure/template"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"golang.org/x/exp/slog"
)

func main() {
	ctx := context.Background()

	// observability
	h := arrower.NewFilteredLogger(os.Stderr)
	// h.SetLogLevel(arrower.LevelTrace)
	//h.SetLogLevel(arrower.LevelDebug)
	h.SetLogLevel(slog.LevelDebug)
	logger := h.Logger

	exporter, err := prometheus.New()
	if err != nil {
		panic(err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter("github.com/open-telemetry/opentelemetry-go/example/prometheus")

	// Start the prometheus HTTP server and pass the exporter Collector to it
	go serveMetrics()

	// dependencies
	pg, err := postgres.ConnectAndMigrate(ctx, postgres.Config{
		User:       "arrower",
		Password:   "secret",
		Database:   "arrower",
		Host:       "localhost",
		Port:       5432, //nolint:gomnd
		MaxConns:   100,  //nolint:gomnd
		Migrations: postgres.ArrowerDefaultMigrations,
	})
	if err != nil {
		panic(err)
	}

	router := echo.New()
	router.Debug = true // todo only in dev mode
	router.Logger.SetOutput(io.Discard)
	router.Use(middleware.Static("public"))
	router.Use(injectMW)

	queue, _ := jobs.NewGueJobs(pg.PGx)

	{ // example queue workers
		_ = queue.RegisterWorker(func(ctx context.Context, j someJob) error {
			time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

			if rand.Intn(100) > 60 { //nolint:gosec,gomnd
				return errors.New("some error") //nolint:goerr113
			}

			return nil
		})

		_ = queue.RegisterWorker(func(ctx context.Context, j longRunningJob) error {
			time.Sleep(time.Duration(rand.Intn(5)) * time.Minute)

			if rand.Intn(100) > 95 { //nolint:gosec,gomnd
				return errors.New("some error") //nolint:goerr113
			}

			return nil
		})

		_ = queue.RegisterWorker(func(ctx context.Context, j otherJob) error { return nil })
		_ = queue.StartWorkers()
	}

	// example trace
	exporterT, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}
	// labels/tags/resources that are common to all traces.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("my-tempo-service-name"),
		attribute.String("some-attribute", "some-value"),
	)

	providerT := trace.NewTracerProvider(
		//trace.WithBatcher(exporterT), // prod
		trace.WithSyncer(exporterT), // dev
		trace.WithResource(resource),
		// set the sampling rate based on the parent span to 60%
		//trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.6))), // prod
		trace.WithSampler(trace.AlwaysSample()), // dev
	)

	{ // example metrics
		opt := api.WithAttributes(
			attribute.Key("A").String("B"),
			attribute.Key("C").String("D"),
		)

		// This is the equivalent of prometheus.NewCounterVec
		counter, err := meter.Float64Counter("foo", api.WithDescription("a simple counter"))
		if err != nil {
			log.Fatal(err)
		}
		counter.Add(ctx, 5, opt)

		router.GET("/add", func(c echo.Context) error {

			newCtx, span := providerT.Tracer("myTracer").Start(c.Request().Context(), "add",
				trace2.WithAttributes(attribute.String("component", "addition")),
				trace2.WithAttributes(attribute.String("someKey", "someValue")),
			)
			defer span.End()

			counter.Add(newCtx, 1, opt)

			return c.HTML(http.StatusOK, "Counter incremented")
		})
	}

	router.GET("/", func(c echo.Context) error {
		for i := 0; i < 256; i++ {
			_ = queue.Enqueue(ctx, someJob{
				Val: randomString(rand.Intn(1000)),
				Field: Field{
					F0: randomString(8),
					F1: rand.Intn(32),
				},
			}, jobs.WithRunAt(time.Now().Add(time.Second*20)))
			_ = queue.Enqueue(ctx, otherJob{}, jobs.WithRunAt(time.Now().Add(time.Second*20)))
		}
		for i := 0; i < 8; i++ {
			_ = queue.Enqueue(ctx, longRunningJob{"Hallo long running job!"}, jobs.WithRunAt(time.Now().Add(time.Second*10)))
		}

		return c.Render(http.StatusOK, "global=>home", "World") //nolint:wrapcheck
	})

	r, _ := template.NewRenderer(logger, os.DirFS("shared/interfaces/web/views"), true)
	router.Renderer = r

	_ = startup.Init(router, pg, logger)

	router.Logger.Fatal(router.Start(":8080"))

	_ = queue.Shutdown(ctx)
}

type someJob struct {
	Val   string
	Field Field
}
type longRunningJob struct {
	Val string
}

type otherJob struct{}

type Field struct {
	F0 string
	F1 int
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789 ")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func serveMetrics() {
	log.Printf("serving metrics at localhost:2223/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2223", nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}

func injectMW(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := next(c); err != nil {
			c.Error(err)
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
