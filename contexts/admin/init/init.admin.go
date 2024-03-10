package init

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

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

func NewAdminContext(di *infrastructure.Container) (*AdminContext, error) {
	logger := di.Logger.With(slog.String("context", contextName))

	jobRepository := repository.NewTracedJobsRepository(repository.NewPostgresJobsRepository(di.PGx))

	adminContext := &AdminContext{
		globalContainer: di,

		logger: logger,

		jobRepository: jobRepository,

		settingsController: web.NewSettingsController(di.AdminRouter),
		jobsController: web.NewJobsController(
			logger,
			models.New(di.PGx),
			jobRepository,
			sweb.NewDefaultPresenter(di.Settings),
			application.NewLoggedJobsApplication(application.NewJobsApplication(di.PGx), logger),
		),
		logsController: web.NewLogsController(
			logger,
			di.Settings,
			alogmodels.New(di.PGx),
			di.AdminRouter.Group("/logs"),
			sweb.NewDefaultPresenter(di.Settings),
		),
	}

	var views fs.FS = views.AdminViews
	if di.Config.Debug {
		views = os.DirFS("contexts/admin/internal/views")
	}

	err := di.WebRenderer.AddContext(contextName, views)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	registerAdminRoutes(adminContext)

	logger.Debug("context admin initialised")

	return adminContext, nil
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
