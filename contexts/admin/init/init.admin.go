// Package init is the context's startup API.
//
// Put all initialisations here.
// For example, load context-specific configuration, setup dependency injection,
// register routes, workers and more.
package init

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/go-arrower/arrower/app"

	"github.com/go-arrower/arrower/alog"
	alogmodels "github.com/go-arrower/arrower/alog/models"

	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
	"github.com/go-arrower/skeleton/contexts/admin/internal/views"
	"github.com/go-arrower/skeleton/shared/infrastructure"
	sweb "github.com/go-arrower/skeleton/shared/interfaces/web"
)

const contextName = "admin"

func NewAdminContext(ctx context.Context, di *infrastructure.Container) (*AdminContext, error) {
	err := ensureRequiredDependencies(di)
	if err != nil {
		return nil, fmt.Errorf("missing dependencies to initialise context admin: %w", err)
	}

	admin, err := setupAdminContext(di)
	if err != nil {
		return nil, fmt.Errorf("could not initialise context admin: %w", err)
	}

	admin.logger.DebugContext(ctx, "context admin initialised")

	return admin, nil
}

type AdminContext struct {
	globalContainer *infrastructure.Container

	logger alog.Logger

	jobRepository jobs.Repository

	settingsController *web.SettingsController
	jobsController     *web.JobsController
	logsController     *web.LogsController
}

func (c *AdminContext) Shutdown(_ context.Context) error {
	return nil
}

func ensureRequiredDependencies(di *infrastructure.Container) error {
	if di.Logger == nil {
		return fmt.Errorf("%w: logger", infrastructure.ErrMissingDependency)
	}

	if di.PGx == nil {
		return fmt.Errorf("%w: pgx", infrastructure.ErrMissingDependency)
	}

	if di.WebRouter == nil {
		return fmt.Errorf("%w: web router", infrastructure.ErrMissingDependency)
	}

	if di.AdminRouter == nil {
		return fmt.Errorf("%w: admin router", infrastructure.ErrMissingDependency)
	}

	if di.WebRenderer == nil {
		return fmt.Errorf("%w: renderer", infrastructure.ErrMissingDependency)
	}

	if di.Settings == nil {
		return fmt.Errorf("%w: settings", infrastructure.ErrMissingDependency)
	}

	return nil
}

func setupAdminContext(di *infrastructure.Container) (*AdminContext, error) {
	logger := di.Logger.With(slog.String("context", contextName))

	jobRepository := repository.NewTracedJobsRepository(repository.NewPostgresJobsRepository(di.PGx))

	admin := &AdminContext{
		globalContainer: di,

		logger: logger,

		jobRepository: jobRepository,

		settingsController: web.NewSettingsController(di.AdminRouter),
		jobsController: web.NewJobsController(logger, models.New(di.PGx), jobRepository, sweb.NewDefaultPresenter(di.Settings), application.NewLoggedJobsApplication(
			application.NewJobsApplication(
				di.PGx,
				models.New(di.PGx),
				repository.NewPostgresJobsRepository(di.PGx),
			),
			logger,
		),
			application.App{ // todo add instrumentation
				PruneJobHistory: app.NewInstrumentedRequest(di.TraceProvider, di.MeterProvider, di.Logger, application.NewPruneJobHistoryRequestHandler(models.New(di.PGx))),
			},
		),
		logsController: web.NewLogsController(
			logger,
			di.Settings,
			alogmodels.New(di.PGx),
			di.AdminRouter.Group("/logs"),
			sweb.NewDefaultPresenter(di.Settings),
		),
	}

	{ // add context-specific web views.
		var views fs.FS = views.AdminViews
		if di.Config.Debug {
			views = os.DirFS("contexts/admin/internal/views")
		}

		err := di.WebRenderer.AddContext(contextName, views)
		if err != nil {
			return nil, fmt.Errorf("could not add context views: %w", err)
		}
	}

	registerAdminRoutes(admin)

	return admin, nil
}
