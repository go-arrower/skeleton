package application

import "github.com/go-arrower/arrower/app"

// App is a dependency injection container.
type App struct {
	PruneJobHistory app.Request[PruneJobHistoryRequest, PruneJobHistoryResponse]
	VacuumJobTable  app.Request[VacuumJobTableRequest, VacuumJobTableResponse]
	DeleteJob       app.Command[DeleteJobCommand]
	GetQueue        app.Query[GetQueueQuery, GetQueueResponse]
	GetWorkers      app.Query[GetWorkersQuery, GetWorkersResponse]
}
