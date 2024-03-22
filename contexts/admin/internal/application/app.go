package application

import "github.com/go-arrower/arrower/app"

// App is a dependency injection container.
type App struct {
	PruneJobHistory app.Request[PruneJobHistoryRequest, PruneJobHistoryResponse]
}
