package application

import (
	"context"
	"fmt"
	"strings"

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
	ListUsersQuery struct {
		Query string
	}
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

	total := uint(len(users))

	users = searchUsersEXPENSIVE(users, query.Query)

	return ListUsersResponse{
		Users:    users,
		Filtered: uint(len(users)),
		Total:    total,
	}, nil
}

// searchUsersEXPENSIVE should be done by the database instead of here
// if the list of users grows beyond the current testing size.
func searchUsersEXPENSIVE(usrs []user.User, query string) []user.User {
	users := []user.User{}

	query = strings.TrimSpace(strings.ToLower(query))

	for _, u := range usrs {
		searchNameConcat := strings.ToLower(u.Name.FirstName()) +
			strings.ToLower(u.Name.LastName()) +
			strings.ToLower(u.Name.DisplayName())

		matchesSearch := strings.Contains(string(u.Login), query) || strings.Contains(searchNameConcat, query)
		if matchesSearch {
			users = append(users, u)
		}
	}

	return users
}
