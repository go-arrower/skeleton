package init

import (
	"context"
	"fmt"

	"github.com/go-arrower/arrower/mw"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application"

	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"

	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/web"

	"github.com/go-arrower/skeleton/shared/infrastructure"
)

const contextName = "auth"

type AuthContext struct {
	userController     web.UserController
	settingsController web.SettingsController
}

func NewAuthContext(di *infrastructure.Container) (*AuthContext, error) {
	// todo if di == nil => load and initialise all dependencies from config

	if err := di.EnsureAllDependenciesPresent(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	logger := di.Logger.WithGroup(contextName)
	meter := di.MeterProvider.Meter(fmt.Sprintf("%s/%s", di.Config.ApplicationName, contextName))
	tracer := di.TraceProvider.Tracer(fmt.Sprintf("%s/%s", di.Config.ApplicationName, contextName))
	_ = meter
	_ = tracer

	queries := models.New(di.DB)
	authContext := AuthContext{
		userController: web.UserController{
			Queries: queries,
			CmdLoginUser: mw.Traced(di.TraceProvider,
				mw.Metric(di.MeterProvider,
					mw.Logged(logger,
						mw.Validate(nil,
							application.LoginUser(di.Logger, queries, di.ArrowerQueue),
						),
					),
				),
			),
			CmdRegisterUser: mw.Traced(di.TraceProvider,
				mw.Metric(di.MeterProvider,
					mw.Logged(logger,
						mw.Validate(nil,
							application.RegisterUser(di.Logger, queries, di.ArrowerQueue),
						),
					),
				),
			),
			CmdShowUserUser: mw.Traced(di.TraceProvider,
				mw.Metric(di.MeterProvider,
					mw.Logged(logger,
						mw.Validate(nil,
							application.ShowUser(queries),
						),
					),
				),
			),
		},
	}

	_ = authContext.registerWebRoutes(di.WebRouter.Group(fmt.Sprintf("/%s", contextName)))
	_ = authContext.registerAPIRoutes(di.APIRouter)
	_ = authContext.registerAdminRoutes(di.AdminRouter.Group(fmt.Sprintf("/%s", contextName)), localDI{queries: queries}) // todo only, if admin context is present

	_ = authContext.registerJobs(di.ArrowerQueue)

	return &authContext, nil
}

func (c *AuthContext) Shutdown(ctx context.Context) error {
	return nil
}

type localDI struct {
	queries *models.Queries
}
