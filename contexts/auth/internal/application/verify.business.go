package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

var ErrVerificationFailed = errors.New("verification failed")

type VerificationToken struct {
	validUntil time.Time
	userID     user.ID
	token      uuid.UUID
}

func (t VerificationToken) Token() uuid.UUID {
	return t.token
}

func (t VerificationToken) UserID() user.ID {
	return t.userID
}

type VerificationOpt func(vs *VerificationService)

// WithValidTime overwrites the time a VerificationToken is valid.
func WithValidTime(validTime time.Duration) VerificationOpt {
	return func(vs *VerificationService) {
		vs.validTime = validTime
	}
}

func NewVerificationService(queries *models.Queries, opts ...VerificationOpt) *VerificationService {
	const oneWeek = time.Hour * 24 * 7 // valid for one week

	verificationService := &VerificationService{
		queries:   queries,
		validTime: oneWeek,
	}

	for _, opt := range opts {
		opt(verificationService)
	}

	return verificationService
}

type VerificationService struct {
	queries   *models.Queries
	validTime time.Duration
}

func (s *VerificationService) NewVerificationToken(ctx context.Context, user user.User) (VerificationToken, error) {
	token := VerificationToken{
		token:      uuid.New(),
		validUntil: time.Now().UTC().Add(s.validTime),
		userID:     user.ID,
	}

	err := s.queries.CreateVerificationToken(ctx, models.CreateVerificationTokenParams{
		Token:         token.token,
		UserID:        uuid.MustParse(string(token.userID)),
		ValidUntilUtc: pgtype.Timestamptz{Time: token.validUntil, Valid: true, InfinityModifier: pgtype.Finite},
	})
	if err != nil {
		return VerificationToken{}, fmt.Errorf("could not save new verification token: %w", err)
	}

	return token, nil
}

func (s *VerificationService) Verify(ctx context.Context, usr *user.User, rawToken uuid.UUID) error {
	token, err := s.queries.VerificationTokenByToken(ctx, rawToken)
	if err != nil {
		return fmt.Errorf("%w: could not fetch verification token: %v", ErrVerificationFailed, err)
	}

	if user.ID(token.UserID.String()) != usr.ID {
		return ErrVerificationFailed
	}

	if time.Now().UTC().After(token.ValidUntilUtc.Time) {
		return ErrVerificationFailed
	}

	usr.Verified = usr.Verified.SetTrue()

	err = repository.SaveUser(ctx, s.queries, *usr)
	if err != nil {
		return fmt.Errorf("could not save user: %w", err)
	}

	return nil
}
