package application

import (
	"context"
	"fmt"

	"github.com/go-arrower/arrower/alog"
	"github.com/go-arrower/arrower/app"

	"github.com/go-arrower/skeleton/shared/domain"
)

func NewSayHelloRequestHandler(logger alog.Logger) app.Request[SayHelloRequest, domain.TeamMember] {
	return app.NewValidatedRequest[SayHelloRequest, domain.TeamMember](
		nil,
		&sayHelloRequestHandler{logger: logger},
	)
}

type sayHelloRequestHandler struct {
	logger alog.Logger
}

type (
	SayHelloRequest struct {
		Name string `validate:"required,max=25"`
	}
)

func (h *sayHelloRequestHandler) H(ctx context.Context, req SayHelloRequest) (domain.TeamMember, error) {
	tm, err := domain.IsTeamMember(domain.MemberName(req.Name))
	if err != nil {
		return domain.TeamMember{}, fmt.Errorf("%w", err)
	}

	h.logger.InfoContext(ctx, "say hello", "team_member", tm)

	return tm, nil
}
