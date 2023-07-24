package init

import (
	"context"
	"fmt"

	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"

	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/web"

	"github.com/go-arrower/skeleton/shared/infrastructure"
)

const contextName = "auth"

type AuthContext struct {
	tenantController web.TenantController
	userController   web.UserController
}

func NewAuthContext(di *infrastructure.Container) (*AuthContext, error) {
	if err := di.EnsureAllDependenciesPresent(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	logger := di.Logger.WithGroup(contextName)
	meter := di.MeterProvider.Meter(fmt.Sprintf("%s/%s", di.Config.ApplicationName, contextName))
	tracer := di.TraceProvider.Tracer(fmt.Sprintf("%s/%s", di.Config.ApplicationName, contextName))
	_ = logger
	_ = meter
	_ = tracer

	authContext := AuthContext{
		tenantController: web.TenantController{Queries: models.New(di.DB)},
		userController:   web.UserController{Queries: models.New(di.DB)},
	}

	_ = authContext.registerWebRoutes(di.WebRouter.Group(fmt.Sprintf("/%s", contextName)))
	_ = authContext.registerAPIRoutes(di.APIRouter)
	_ = authContext.registerAdminRoutes(di.AdminRouter.Group(fmt.Sprintf("/%s", contextName))) // todo only, if admin context is present

	_ = authContext.registerJobs(di.ArrowerQueue)

	return &AuthContext{}, nil
}

func (c *AuthContext) Shutdown(ctx context.Context) error {
	return nil
}
