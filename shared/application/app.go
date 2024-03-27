package application

import (
	"github.com/go-arrower/arrower/app"
	"github.com/go-arrower/skeleton/shared/domain"
)

type App struct {
	SayHello app.Request[SayHelloRequest, domain.TeamMember]
}
