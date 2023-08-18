package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrVerificationFailed = errors.New("verification failed")

func NewVerificationToken(token uuid.UUID, userID ID, validUntilUTC time.Time) VerificationToken {
	return VerificationToken{
		validUntil: validUntilUTC,
		userID:     userID,
		token:      token,
	}
}

type VerificationToken struct {
	validUntil time.Time
	userID     ID
	token      uuid.UUID
}

func (t VerificationToken) Token() uuid.UUID {
	return t.token
}

func (t VerificationToken) UserID() ID {
	return t.userID
}

func (t VerificationToken) ValidUntilUTC() time.Time {
	return t.validUntil
}

type VerificationOpt func(vs *VerificationService)

// WithValidTime overwrites the time a VerificationToken is valid.
func WithValidTime(validTime time.Duration) VerificationOpt {
	return func(vs *VerificationService) {
		vs.validTime = validTime
	}
}

func NewVerificationService(repo Repository, opts ...VerificationOpt) *VerificationService {
	const oneWeek = time.Hour * 24 * 7 // default time a token is valid.

	verificationService := &VerificationService{
		repo:      repo,
		validTime: oneWeek,
	}

	for _, opt := range opts {
		opt(verificationService)
	}

	return verificationService
}

type VerificationService struct {
	repo      Repository
	validTime time.Duration
}

// todo add docs and rename more descriptive.
func (s *VerificationService) NewVerificationToken(ctx context.Context, user User) (VerificationToken, error) {
	token := VerificationToken{
		token:      uuid.New(),
		validUntil: time.Now().UTC().Add(s.validTime),
		userID:     user.ID,
	}

	err := s.repo.CreateVerificationToken(ctx, token)
	if err != nil {
		return VerificationToken{}, fmt.Errorf("could not save new verification token: %w", err)
	}

	return token, nil
}

func (s *VerificationService) Verify(ctx context.Context, usr *User, rawToken uuid.UUID) error {
	token, err := s.repo.VerificationTokenByToken(ctx, rawToken)
	if err != nil {
		return fmt.Errorf("%w: could not fetch verification token: %v", ErrVerificationFailed, err)
	}

	if token.UserID() != usr.ID {
		return ErrVerificationFailed
	}

	if time.Now().UTC().After(token.ValidUntilUTC()) {
		return ErrVerificationFailed
	}

	usr.Verified = usr.Verified.SetTrue()

	err = s.repo.Save(ctx, *usr)
	if err != nil {
		return fmt.Errorf("could not save user: %w", err)
	}

	return nil
}
