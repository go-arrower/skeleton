package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

func RepoGetUserByLogin(ctx context.Context, queries *models.Queries, loginEmail string) (user.User, error) {
	dbUser, err := queries.FindUserByLogin(ctx, loginEmail)
	if err != nil {
		return user.User{}, fmt.Errorf("%w", err)
	}

	return userFromModel(dbUser, nil), nil
}

func RepoGetUserByID(ctx context.Context, queries *models.Queries, userID user.ID) (user.User, error) {
	dbUser, err := queries.FindUserByID(ctx, uuid.MustParse(string(userID)))
	if err != nil {
		return user.User{}, fmt.Errorf("%w", err)
	}

	sess, err := queries.FindSessionsByUserID(ctx, uuid.NullUUID{UUID: uuid.MustParse(string(userID)), Valid: true})
	if err != nil {
		return user.User{}, fmt.Errorf("%w", err)
	}

	return userFromModel(dbUser, sess), nil
}

func userFromModel(dbUser models.AuthUser, sessions []models.AuthSession) user.User {
	prof := make(map[string]*string)

	profile := dbUser.Profile.Scan(&prof)
	_ = profile
	_ = prof
	_ = dbUser.Profile.Value

	return user.User{
		ID:                user.ID(dbUser.ID.String()),
		Login:             user.Login(dbUser.Login),
		PasswordHash:      user.PasswordHash(dbUser.PasswordHash),
		RegisteredAt:      dbUser.CreatedAt.Time,
		FirstName:         dbUser.FirstName,
		LastName:          dbUser.LastName,
		Name:              dbUser.Name,
		Birthday:          user.Birthday{}, // todo
		Locale:            user.Locale{},   // todo
		TimeZone:          user.TimeZone(dbUser.TimeZone),
		ProfilePictureURL: user.URL(dbUser.PictureUrl),
		Profile:           user.Profile{},
		Profile2:          prof, // todo
		Verified:          user.VerifiedFlag(dbUser.VerifiedAt.Time),
		Blocked:           user.BlockedFlag(dbUser.BlockedAt.Time),
		SuperUser:         user.SuperUserFlag(dbUser.SuperUserAt.Time),
		Sessions:          sessionsFromModel(sessions),
	}
}

func sessionsFromModel(sess []models.AuthSession) []user.Session {
	if sess == nil {
		return []user.Session{}
	}

	sessions := make([]user.Session, len(sess))

	for i := range sess {
		sessions[i] = user.Session{
			ID:        string(sess[i].Key),
			CreatedAt: sess[i].CreatedAt.Time,
			ExpiresAt: sess[i].ExpiresAt.Time,
			Device:    user.NewDevice(sess[i].UserAgent),
		}
	}

	return sessions
}
