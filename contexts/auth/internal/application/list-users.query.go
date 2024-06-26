package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-arrower/skeleton/contexts/auth/internal/domain"

	"github.com/go-arrower/arrower/app"
)

func NewListUsersQueryHandler(repo domain.Repository) app.Query[ListUsersQuery, ListUsersResponse] {
	return &listUsersQueryHandler{repo: repo}
}

type listUsersQueryHandler struct {
	repo domain.Repository
}

type (
	ListUsersQuery struct {
		Query  string
		Filter domain.Filter
	}
	ListUsersResponse struct {
		Users    []domain.User
		Filtered uint
		Total    uint
	}
)

func (h *listUsersQueryHandler) H(ctx context.Context, query ListUsersQuery) (ListUsersResponse, error) {
	users, err := h.repo.All(ctx, query.Filter)
	if err != nil {
		return ListUsersResponse{}, fmt.Errorf("could not get users: %w", err)
	}

	total, err := h.repo.Count(ctx)
	if err != nil {
		return ListUsersResponse{}, fmt.Errorf("could not count users: %w", err)
	}

	users = searchUsersEXPENSIVE(users, query.Query)

	filtered := uint(total)
	if query.Query != "" {
		filtered = uint(len(users))
	}

	return ListUsersResponse{
		Users:    users,
		Filtered: filtered,
		Total:    uint(total),
	}, nil
}

// searchUsersEXPENSIVE should be done by the database instead of here
// if the list of users grows beyond the current testing size.
// / !!! this approach combined with pagination can lead to not all results showing !!!
func searchUsersEXPENSIVE(usrs []domain.User, query string) []domain.User {
	users := []domain.User{}

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
