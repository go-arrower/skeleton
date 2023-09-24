package infrastructure

import (
	"errors"
	"fmt"

	"github.com/go-arrower/skeleton/contexts/admin"

	"github.com/go-arrower/arrower/alog"
	"github.com/go-arrower/arrower/jobs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	ErrMissingDependency = errors.New("missing dependency")
)

// Container holds global dependencies that can be used within each Context, to make initialisation easier,
// if the Context can operate with the shared resources. Otherwise, the Context is advised to initialise its
// own dependencies from its own configuration.
type Container struct {
	Logger        alog.Logger
	MeterProvider *metric.MeterProvider
	TraceProvider *trace.TracerProvider

	Config       *Config
	DB           *pgxpool.Pool
	DefaultQueue jobs.Queue
	WebRouter    *echo.Echo
	APIRouter    *echo.Group
	AdminRouter  *echo.Group

	ArrowerQueue jobs.Queue

	SettingsService admin.SettingsAPI
}

func (c *Container) EnsureAllDependenciesPresent() error {
	if c.Config == nil {
		return fmt.Errorf("%w: global config not found", ErrMissingDependency)
	}

	return nil
}
