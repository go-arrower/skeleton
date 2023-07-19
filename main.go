package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-arrower/arrower/alog"
	"github.com/go-arrower/arrower/jobs"
	"github.com/go-arrower/arrower/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	trace2 "go.opentelemetry.io/otel/trace"

	"github.com/go-arrower/skeleton/contexts/admin/startup"
	"github.com/go-arrower/skeleton/shared/infrastructure/template"
)

func main() {
	ctx := context.Background()

	logger, meterProvider, traceProvider := setupTelemetry(ctx)
	alog.Unwrap(logger).SetLevel(alog.LevelDebug)

	// dependencies
	pg, err := postgres.ConnectAndMigrate(ctx, postgres.Config{
		User:       "arrower",
		Password:   "secret",
		Database:   "arrower",
		Host:       "localhost",
		Port:       5432, //nolint:gomnd
		MaxConns:   100,  //nolint:gomnd
		Migrations: postgres.ArrowerDefaultMigrations,
	}, traceProvider)
	if err != nil {
		panic(err)
	}

	router := echo.New()
	router.Debug = true // todo only in dev mode
	router.Logger.SetOutput(io.Discard)
	router.Use(otelecho.Middleware("www.servername.tld", otelecho.WithTracerProvider(traceProvider)))
	router.Use(middleware.Static("public"))
	router.Use(injectMW)

	queue, _ := jobs.NewGueJobs(logger, meterProvider, traceProvider, pg.PGx)

	{ // example metrics
		opt := api.WithAttributes( // todo check if metrics and trae attributes can be shared
			attribute.Key("A").String("B"),
			attribute.Key("C").String("D"),
		)

		tracer := traceProvider.Tracer("myTracer", // namespace per library that is instrumented
			trace2.WithInstrumentationVersion("0.1337"),
		)

		// This is the equivalent of prometheus.NewCounterVec
		meter := meterProvider.Meter("github.com/open-telemetry/opentelemetry-go/example/prometheus",
			api.WithInstrumentationVersion("0.1337"),
		)
		counter, err := meter.Float64Counter("foo", api.WithDescription("a simple counter"))
		if err != nil {
			log.Fatal(err)
		}
		counter.Add(ctx, 5, opt)

		router.GET("/add", func(c echo.Context) error {
			newCtx, span := tracer.Start(c.Request().Context(), "add",
				trace2.WithAttributes(attribute.String("component", "addition")),
				//trace2.WithAttributes(attribute.String("job", "somejob")), // NEEDS TO MATCH WITH THE LOGS LABEL
				// todo what is the difference between a tempo resource and attribute?
			)
			defer span.End()

			{ // example metric to test Examplar
				h, err := meterProvider.Meter("some_hist").Int64Histogram("Das_hist")
				if err != nil {
					panic(err)
				}

				examplar := attribute.NewSet(attribute.KeyValue{
					//Key:   "traceID",
					//Value: attribute.StringValue(span.SpanContext().TraceID().String()),
					Key:   "someKey",
					Value: attribute.StringValue("someVal"),
				})
				e := &examplar

				go func() {
					for {
						t := time.NewTicker(1 * time.Second)
						select {
						case <-t.C:
							h.Record(newCtx, int64(rand.Intn(10)), api.WithAttributes(append(e.ToSlice(), attribute.String("component", "hist"))...))
							h.Record(newCtx, int64(rand.Intn(10)), api.WithAttributes(append(e.ToSlice(), attribute.String("component", "hist"))...))
							h.Record(newCtx, int64(rand.Intn(100)), api.WithAttributes(append(e.ToSlice(), attribute.String("component", "hist"))...))

							//h.(prometheus.ExemplarObserver).ObserveWithExemplar(
							//	time.Since(time.Now().Add(-5*time.Second)).Seconds(), prometheus.Labels{"traceID": span.SpanContext().TraceID().String()},
							//)
						}
					}
				}()
			}

			time.Sleep(5 * time.Second)

			return c.HTML(http.StatusOK, "Counter incremented")
		})
	}

	router.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "global=>home", "World") //nolint:wrapcheck
	})

	r, _ := template.NewRenderer(logger, traceProvider, os.DirFS("shared/interfaces/web/views"), true)
	router.Renderer = r

	_ = startup.Init(logger, traceProvider, meterProvider, router, pg, queue)

	router.Logger.Fatal(router.Start(":8080"))

	_ = queue.Shutdown(ctx)
	_ = traceProvider.Shutdown(ctx)
	_ = meterProvider.Shutdown(ctx)
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789 ")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
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
