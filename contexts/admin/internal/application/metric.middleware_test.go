package application_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	prometheusSDK "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

func TestMetric(t *testing.T) {
	t.Parallel()

	/*
		About the test cases and how assertions are set up:

		The testing of metrics is done against a prometheus, so that is as close to the original as possible and
		does not depend on mocks or fakes.
		Using the normal promhttp.HandlerFor registers the handler in the default mux and tests can not run in parallel.
		Using a custom mux does not work. Prometheus offers a solution with testutil,
		see https://github.com/open-telemetry/opentelemetry-go/blob/main/exporters/prometheus/exporter_test.go
	*/

	handler := http.HandlerFunc(promhttp.HandlerFor(
		prometheusSDK.DefaultGatherer,
		promhttp.HandlerOpts{EnableOpenMetrics: true}, //nolint:exhaustruct // to enable Examplars in the export format
	).ServeHTTP)

	t.Run("successful command", func(t *testing.T) {
		t.Parallel()

		// setup prometheus exporter for testing
		registry := prometheusSDK.NewRegistry()
		exporter, _ := prometheus.New(prometheus.WithRegisterer(registry))
		meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))

		cmd := application.Metric(meterProvider, func(context.Context, exampleCommand) (string, error) {
			return "", nil
		})

		_, _ = cmd(context.Background(), exampleCommand{})

		// call the prometheus endpoint to scrape all metrics
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "", nil)

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		err := testutil.GatherAndCompare(
			registry,
			metricsForSucceedingUseCase,
			// restrict to specific metrics prefix.
			// Prevents missing boilerplate metrics and varying values of usecases_duration_sum.
			"usecases_total", "usecases_duration_seconds_bucket")
		assert.NoError(t, err)
	})

	t.Run("failed command", func(t *testing.T) {
		t.Parallel()

		// setup prometheus exporter for testing
		registry := prometheusSDK.NewRegistry()
		exporter, _ := prometheus.New(prometheus.WithRegisterer(registry))
		meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))

		cmd := application.Metric(meterProvider, func(context.Context, exampleCommand) (string, error) {
			return "", errUseCaseFails
		})

		_, _ = cmd(context.Background(), exampleCommand{})

		// call the prometheus endpoint to scrape all metrics
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "", nil)

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		err := testutil.GatherAndCompare(
			registry,
			metricsForFailingUseCase,
			// restrict to specific metrics prefix.
			// Prevents missing boilerplate metrics and varying values of usecases_duration_sum.
			"usecases_total", "usecases_duration_seconds_bucket")
		assert.NoError(t, err)
	})
}
