package init

import (
	"context"

	alogmodels "github.com/go-arrower/arrower/alog/models"
	"github.com/go-arrower/skeleton/contexts/admin/internal/application"
	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/repository/models"
	"github.com/go-arrower/skeleton/contexts/admin/internal/interfaces/web"
	"github.com/go-arrower/skeleton/shared/infrastructure"
	web2 "github.com/go-arrower/skeleton/shared/interfaces/web"
)

func NewAdminContext(di *infrastructure.Container) (*AdminContext, error) {
	adminContext := &AdminContext{
		Container: di,

		jobRepository: repository.NewTracedJobsRepository(repository.NewPostgresJobsRepository(di.PGx)),

		settingsController: web.NewSettingsController(di.AdminRouter),
		logsController:     web.NewLogsController(di.Logger, di.Settings, alogmodels.New(di.PGx), di.AdminRouter.Group("/logs"), web2.NewDefaultPresenter(di.Settings)),
	}

	jobsController := web.NewJobsController(di.Logger, adminContext.jobRepository, web2.NewDefaultPresenter(di.Settings), application.NewJobsApplication(di.PGx))
	jobsController.Queries = models.New(di.PGx)
	adminContext.jobsController = jobsController

	registerAdminRoutes(adminContext)

	return adminContext, nil
}

type AdminContext struct {
	*infrastructure.Container

	jobRepository jobs.Repository

	settingsController *web.SettingsController
	jobsController     *web.JobsController
	logsController     *web.LogsController
}

func (c *AdminContext) Shutdown(_ context.Context) error {
	return nil
}
