package application

import (
	"context"
	"fmt"

	"github.com/go-arrower/arrower/app"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
)

func NewListUsersQueryHandler(repo user.Repository) app.Query[ListUsersQuery, ListUsersResponse] {
	return &listUsersQueryHandler{repo: repo}
}

type listUsersQueryHandler struct {
	repo user.Repository
}

type (
	ListUsersQuery    struct{}
	ListUsersResponse struct {
		Users    []user.User
		Filtered uint
		Total    uint
	}
)

func (h *listUsersQueryHandler) H(ctx context.Context, query ListUsersQuery) (ListUsersResponse, error) {
	users, err := h.repo.All(ctx)
	if err != nil {
		return ListUsersResponse{}, fmt.Errorf("could not get users: %w", err)
	}

	return ListUsersResponse{
		Users:    users,
		Filtered: uint(len(users)),
		Total:    uint(len(users)),
	}, nil
}
