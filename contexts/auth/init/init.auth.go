package init

import (
	"context"
	"fmt"
	"log/slog"

	web2 "github.com/go-arrower/skeleton/shared/interfaces/web"

	"github.com/go-arrower/skeleton/contexts/admin"

	"github.com/go-arrower/arrower/mw"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application"
	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/web"
	"github.com/go-arrower/skeleton/shared/infrastructure"
)

const contextName = "auth"

func NewAuthContext(di *infrastructure.Container) (*AuthContext, error) {
	// todo if di == nil => load and initialise all dependencies from config

	if err := di.EnsureAllDependenciesPresent(); err != nil {
		return nil, fmt.Errorf("could not initialise auth context: %w", err)
	}

	logger := di.Logger.WithGroup(contextName)
	meter := di.MeterProvider.Meter(fmt.Sprintf("%s/%s", di.Config.ApplicationName, contextName))
	tracer := di.TraceProvider.Tracer(fmt.Sprintf("%s/%s", di.Config.ApplicationName, contextName))
	_ = meter
	_ = tracer

	queries := models.New(di.DB)
	repo, _ := repository.NewPostgresRepository(di.DB)
	webRoutes := di.WebRouter.Group(fmt.Sprintf("/%s", contextName))
	adminRouter := di.AdminRouter.Group(fmt.Sprintf("/%s", contextName))

	userController := web.NewUserController(webRoutes, web2.NewDefaultPresenter(di.SettingsService), []byte("secret"))
	userController.Queries = queries
	userController.CmdLoginUser = mw.Traced(di.TraceProvider,
		mw.Metric(di.MeterProvider,
			mw.Logged(logger,
				mw.Validate(nil,
					application.LoginUser(di.Logger, repo, di.ArrowerQueue),
				),
			),
		),
	)
	userController.CmdRegisterUser = mw.Traced(di.TraceProvider,
		mw.Metric(di.MeterProvider,
			mw.Logged(logger,
				mw.Validate(nil,
					application.RegisterUser(di.Logger, di.SettingsService, repo, di.ArrowerQueue),
				),
			),
		),
	)
	userController.CmdShowUserUser = mw.Traced(di.TraceProvider,
		mw.Metric(di.MeterProvider,
			mw.Logged(logger,
				mw.Validate(nil,
					application.ShowUser(repo),
				),
			),
		),
	)
	userController.CmdVerifyUser = mw.TracedU(di.TraceProvider,
		mw.MetricU(di.MeterProvider,
			mw.LoggedU(logger,
				mw.ValidateU(nil,
					application.VerifyUser(repo),
				),
			),
		),
	)
	userController.CmdBlockUser = mw.Traced(di.TraceProvider,
		mw.Metric(di.MeterProvider,
			mw.Logged(logger,
				mw.Validate(nil,
					application.BlockUser(repo),
				),
			),
		),
	)
	userController.CmdUnBlockUser = mw.Traced(di.TraceProvider,
		mw.Metric(di.MeterProvider,
			mw.Logged(logger,
				mw.Validate(nil,
					application.UnblockUser(repo),
				),
			),
		),
	)

	{ // register default auth settings
		di.SettingsService.Add(context.Background(), admin.Setting{
			Key:   admin.SettingRegistration,
			Value: admin.NewSettingValue(true),
			UIOptions: admin.Options{
				Type:         admin.Checkbox,
				Label:        "Enable Registration",
				Info:         "Allows new Users to register themselves",
				DefaultValue: admin.NewSettingValue(true),
				ReadOnly:     false,
				Danger:       false,
			},
		})
		di.SettingsService.Add(context.Background(), admin.Setting{
			Key:   admin.SettingLogin,
			Value: admin.NewSettingValue(true),
			UIOptions: admin.Options{
				Type:         admin.Checkbox,
				Label:        "Enable Login",
				Info:         "Allows Users to login to the application",
				DefaultValue: admin.NewSettingValue(true),
				ReadOnly:     false,
				Danger:       false,
			},
		})
	}

	authContext := AuthContext{
		settingsController: web.NewSettingsController(web2.NewDefaultPresenter(di.SettingsService), queries),
		userController:     userController,
		logger:             logger,
		traceProvider:      di.TraceProvider,
		meterProvider:      di.MeterProvider,
		queries:            queries,
		repo:               repo,
	}

	authContext.registerWebRoutes(webRoutes)
	authContext.registerAPIRoutes(di.APIRouter)
	authContext.registerAdminRoutes(adminRouter, localDI{queries: queries}) // todo only, if admin context is present

	authContext.registerJobs(di.ArrowerQueue)

	return &authContext, nil
}

type AuthContext struct {
	settingsController *web.SettingsController
	userController     web.UserController

	logger        *slog.Logger
	traceProvider trace.TracerProvider
	meterProvider metric.MeterProvider
	queries       *models.Queries
	repo          user.Repository
}

func (c *AuthContext) Shutdown(ctx context.Context) error {
	return nil
}

type localDI struct {
	queries *models.Queries
}
