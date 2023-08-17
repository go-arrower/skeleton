package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/language"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

func GetUserByLogin(ctx context.Context, queries *models.Queries, loginEmail string) (user.User, error) {
	dbUser, err := queries.FindUserByLogin(ctx, loginEmail)
	if err != nil {
		return user.User{}, fmt.Errorf("%w", err)
	}

	return userFromModelWithSession(dbUser, nil), nil
}

func GetUserByID(ctx context.Context, queries *models.Queries, userID user.ID) (user.User, error) {
	dbUser, err := queries.FindUserByID(ctx, uuid.MustParse(string(userID)))
	if err != nil {
		return user.User{}, fmt.Errorf("%w", err)
	}

	sess, err := queries.FindSessionsByUserID(ctx, uuid.NullUUID{UUID: uuid.MustParse(string(userID)), Valid: true})
	if err != nil {
		return user.User{}, fmt.Errorf("%w", err)
	}

	return userFromModelWithSession(dbUser, sess), nil
}

func usersFromModel(ctx context.Context, queries *models.Queries, dbUsers []models.AuthUser) ([]user.User, error) {
	users := make([]user.User, len(dbUsers))

	for i, u := range dbUsers {
		user, err := userFromModel(ctx, queries, u)
		if err != nil {
			return nil, err
		}

		users[i] = user
	}

	return users, nil
}

func userFromModel(ctx context.Context, queries *models.Queries, dbUser models.AuthUser) (user.User, error) {
	sess, err := queries.FindSessionsByUserID(ctx, uuid.NullUUID{UUID: dbUser.ID, Valid: true})
	if err != nil {
		return user.User{}, fmt.Errorf("could not get sessions for user: %s: %v", dbUser.ID.String(), err)
	}

	return userFromModelWithSession(dbUser, sess), nil
}

func userFromModelWithSession(dbUser models.AuthUser, sessions []models.AuthSession) user.User {
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
		Name:              user.NewName(dbUser.NameFirstname, dbUser.NameLastname, dbUser.NameDisplayname),
		Birthday:          user.Birthday{}, // todo
		Locale:            user.Locale{},   // todo
		TimeZone:          user.TimeZone(dbUser.TimeZone),
		ProfilePictureURL: user.URL(dbUser.PictureUrl),
		Profile:           user.Profile{},
		Profile2:          prof, // todo
		Verified:          user.BoolFlag(dbUser.VerifiedAtUtc.Time),
		Blocked:           user.BoolFlag(dbUser.BlockedAtUtc.Time),
		SuperUser:         user.BoolFlag(dbUser.SuperuserAtUtc.Time),
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
			ExpiresAt: sess[i].ExpiresAtUtc.Time,
			Device:    user.NewDevice(sess[i].UserAgent),
		}
	}

	return sessions
}

func SaveUser(ctx context.Context, queries *models.Queries, user user.User) error {
	_, err := queries.UpsertUser(ctx, userToModel(user))
	if err != nil {
		return fmt.Errorf("could not upsert user: %v", err)
	}

	return nil
}

func userToModel(user user.User) models.UpsertUserParams {
	verifiedAt := pgtype.Timestamptz{Time: user.Verified.At(), Valid: true, InfinityModifier: pgtype.Finite}
	if user.Verified.At() == (time.Time{}) {
		verifiedAt = pgtype.Timestamptz{}
	}

	blockedAt := pgtype.Timestamptz{Time: user.Blocked.At(), Valid: true, InfinityModifier: pgtype.Finite}
	if user.Blocked.At() == (time.Time{}) {
		blockedAt = pgtype.Timestamptz{}
	}

	superUserAt := pgtype.Timestamptz{Time: user.SuperUser.At(), Valid: true, InfinityModifier: pgtype.Finite}
	if user.SuperUser.At() == (time.Time{}) {
		superUserAt = pgtype.Timestamptz{}
	}

	return models.UpsertUserParams{
		ID: uuid.MustParse(string(user.ID)),
		// only required for insert, otherwise the time will not be updated.
		CreatedAt:       pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true, InfinityModifier: pgtype.Finite},
		Login:           string(user.Login),
		PasswordHash:    string(user.PasswordHash),
		NameFirstname:   user.Name.FirstName(),
		NameLastname:    user.Name.LastName(),
		NameDisplayname: user.Name.DisplayName(),
		Birthday:        pgtype.Date{}, // todo
		Locale:          language.Tag(user.Locale).String(),
		TimeZone:        string(user.TimeZone),
		PictureUrl:      string(user.ProfilePictureURL),
		Profile:         map[string]*string{}, // todo
		VerifiedAtUtc:   verifiedAt,
		BlockedAtUtc:    blockedAt,
		SuperuserAtUtc:  superUserAt,
	}
}
