package main

import (
	"context"
	"net/http"

	"github.com/go-arrower/arrower/alog"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
)

func setupTelemetry(ctx context.Context) (*slog.Logger, *metric.MeterProvider, *trace.TracerProvider) {
	// labels/tags/resources that are common to all traces.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("arrower.skeleton"),

		// NEEDS TO MATCH WITH THE LOGS LABEL (why? for the "Logs for this span" button in tempo?)
		attribute.String("arrower", "skeleton"),

		// like kubernetes pod name
	)

	// CAN RESOURCES BE ADDED TO LOGGER SO ALL THREE HAVE THE SAME VALUES?

	logger := alog.NewDevelopment()

	exporter, err := prometheus.New()
	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(exporter),
	)

	// otel.SetMeterProvider(meterProvider)

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
		// trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.6))), // prod
		trace.WithSampler(trace.AlwaysSample()), // dev
	)

	// otel.SetTracerProvider(traceProvider)

	// Start the prometheus HTTP server and pass the exporter Collector to it
	go serveMetrics(ctx, logger)

	return logger, meterProvider, traceProvider
}

func serveMetrics(ctx context.Context, logger alog.Logger) {
	const (
		port = ":2223"
		path = "/metrics"
	)

	logger.DebugCtx(ctx, "serving metrics",
		slog.String("port", port),
		slog.String("path", path),
	)

	// http.Handle("/metrics", promhttp.Handler())
	http.Handle(path, promhttp.HandlerFor(
		prometheus2.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true, // to enable Examplars in the export format
		},
	))

	err := http.ListenAndServe(port, nil)
	if err != nil {
		logger.DebugCtx(ctx, "error serving http", slog.String("err", err.Error()))

		return
	}
}
