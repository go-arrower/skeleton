package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	prometheus2 "github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-arrower/arrower"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"golang.org/x/exp/slog"
)

func setupTelemetry(ctx context.Context) (*slog.Logger, *metric.MeterProvider, *trace.TracerProvider) {
	// labels/tags/resources that are common to all traces.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("arrower.skeleton"),
		attribute.String("arrower", "skeleton"), // NEEDS TO MATCH WITH THE LOGS LABEL (why? for the "Logs for this span" button in tempo?)
		// like kubernetes pod name
	)

	// CAN RESOURCES BE ADDED TO LOGGER SO ALL THREE HAVE THE SAME VALUES?

	h := arrower.NewFilteredLogger(os.Stderr)
	// h.SetLogLevel(arrower.LevelTrace)
	h.SetLogLevel(slog.LevelDebug)
	logger := h.Logger

	logger = arrower.NewDevelopmentLogger()

	exporter, err := prometheus.New()
	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(exporter),
	)

	//otel.SetMeterProvider(meterProvider)

	// example trace
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),                   // dev only
		otlptracegrpc.WithDialOption(grpc.WithBlock()), // dev only // useful for testing
	)
	if err != nil {
		panic(err)
	}

	traceProvider := trace.NewTracerProvider(
		// trace.WithBatcher(traceExporter), // prod
		trace.WithSyncer(traceExporter), // dev
		trace.WithResource(resource),
		// set the sampling rate based on the parent span to 60%
		//trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.6))), // prod
		trace.WithSampler(trace.AlwaysSample()), // dev
	)

	//otel.SetTracerProvider(traceProvider)

	// Start the prometheus HTTP server and pass the exporter Collector to it
	go serveMetrics()

	return logger, meterProvider, traceProvider
}

func serveMetrics() {
	log.Printf("serving metrics at localhost:2223/metrics")

	//http.Handle("/metrics", promhttp.Handler())
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus2.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true, // to enable Examplars in the export format
		},
	))

	err := http.ListenAndServe(":2223", nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
