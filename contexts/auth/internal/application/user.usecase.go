package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/go-arrower/arrower/alog"
	"github.com/go-arrower/arrower/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository"
	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrPasswordTooWeak   = errors.New("password too weak")
	ErrLoginFailed       = errors.New("login failed")
	ErrInvalidInput      = errors.New("invalid input")
)

type (
	LoginUserRequest struct {
		LoginEmail string `form:"login" validate:"max=1024,required,email"`
		Password   string `form:"password" validate:"max=1024,min=8"`

		IsNewDevice bool
		UserAgent   string
		IP          string `validate:"ip"`
		SessionKey  string
	}
	LoginUserResponse struct {
		User user.User
	}

	SendConfirmationNewDeviceLoggedIn struct {
		UserID     user.ID
		OccurredAt time.Time
		IP         string
		Device     user.Device
		// Ip Location
	}
)

func LoginUser(logger alog.Logger, queries *models.Queries, queue jobs.Enqueuer) func(context.Context, LoginUserRequest) (LoginUserResponse, error) {
	return func(ctx context.Context, in LoginUserRequest) (LoginUserResponse, error) {
		usr, err := repository.RepoGetUserByLogin(ctx, queries, in.LoginEmail)
		if err != nil {
			logger.Log(ctx, alog.LevelInfo, "login failed",
				slog.String("email", in.LoginEmail),
				slog.String("ip", in.IP),
			)

			return LoginUserResponse{}, ErrLoginFailed
		}

		if !usr.Verified.IsVerified() {
			logger.Log(ctx, alog.LevelInfo, "login failed",
				slog.String("email", in.LoginEmail),
				slog.String("ip", in.IP),
			)

			return LoginUserResponse{}, ErrLoginFailed
		}

		if usr.Blocked.IsBlocked() {
			logger.Log(ctx, alog.LevelInfo, "login failed",
				slog.String("email", in.LoginEmail),
				slog.String("ip", in.IP),
			)

			return LoginUserResponse{}, ErrLoginFailed
		}

		if !usr.PasswordHash.Matches(in.Password) {
			logger.Log(ctx, alog.LevelInfo, "login failed",
				slog.String("email", in.LoginEmail),
				slog.String("ip", in.IP),
			)

			return LoginUserResponse{}, ErrLoginFailed
		}

		// The session is not persisted until the end of the controller.
		// Thus, the session is created here and very short-lived, as the controller will update it with the right values.
		err = queries.UpsertSession(ctx, models.UpsertSessionParams{
			Key:       []byte(in.SessionKey),
			Data:      []byte(""),
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Second), Valid: true},
			UserAgent: in.UserAgent,
		})
		if err != nil {
			return LoginUserResponse{}, fmt.Errorf("could not update session with user agent: %w", err)
		}

		if in.IsNewDevice {
			err = queue.Enqueue(ctx, SendConfirmationNewDeviceLoggedIn{
				UserID:     usr.ID,
				OccurredAt: time.Now().UTC(),
				IP:         in.IP,
				Device:     user.NewDevice(in.UserAgent),
			})
			if err != nil {
				return LoginUserResponse{}, fmt.Errorf("could not queue confirmation about new device: %w", err)
			}
		}

		return LoginUserResponse{User: usr}, nil
	}
}

type (
	RegisterUserRequest struct {
		RegisterEmail          string `form:"login" validate:"max=1024,required,email"`
		Password               string `form:"password" validate:"max=1024,min=8"`
		PasswordConfirmation   string `form:"password_confirmation" validate:"max=1024,eqfield=Password"`
		AcceptedTermsOfService bool   `form:"tos" validate:"required"`

		UserAgent  string
		IP         string `validate:"ip"`
		SessionKey string
	}
	RegisterUserResponse struct {
		User user.User
	}

	SendNewUserVerificationEmail struct {
		UserID     user.ID
		OccurredAt time.Time
		IP         string
		Device     user.Device
		// Ip Location
	}
)

func RegisterUser(logger alog.Logger, queries *models.Queries, queue jobs.Enqueuer) func(context.Context, RegisterUserRequest) (RegisterUserResponse, error) {
	return func(ctx context.Context, in RegisterUserRequest) (RegisterUserResponse, error) {
		//if !mw.PassedValidation(ctx) { /* validate OR return err */ }

		if _, err := queries.FindUserByLogin(ctx, in.RegisterEmail); err == nil {
			logger.Log(ctx, alog.LevelInfo, "register new user failed",
				slog.String("email", in.RegisterEmail),
				slog.String("ip", in.IP),
			)

			return RegisterUserResponse{}, ErrUserAlreadyExists
		}

		pwHash, err := hashStringPassword(in.Password)
		if err != nil {
			logger.Log(ctx, alog.LevelInfo, "register new user failed",
				slog.String("email", in.RegisterEmail),
				slog.String("ip", in.IP),
			)

			return RegisterUserResponse{}, err
		}

		usr, err := queries.CreateUser(ctx, models.CreateUserParams{
			ID:           uuid.MustParse(string(user.NewID())),
			Login:        in.RegisterEmail,
			PasswordHash: string(pwHash),
		})
		if err != nil {
			return RegisterUserResponse{}, fmt.Errorf("could not create user: %w", err)
		}

		err = queue.Enqueue(ctx, SendNewUserVerificationEmail{
			UserID:     user.ID(usr.ID.String()),
			OccurredAt: time.Now().UTC(),
			IP:         in.IP,
			Device:     user.NewDevice(in.UserAgent),
		})
		if err != nil {
			return RegisterUserResponse{}, fmt.Errorf("could not queue job to send verification email: %w", err)
		}

		// The session is not persisted until the end of the controller.
		// Thus, the session is created here and very short-lived, as the controller will update it with the right values.
		err = queries.UpsertSession(ctx, models.UpsertSessionParams{
			Key:       []byte(in.SessionKey),
			Data:      []byte(""),
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Second), Valid: true},
			UserAgent: in.UserAgent,
		})
		if err != nil {
			return RegisterUserResponse{}, fmt.Errorf("could not update session with user agent: %w", err)
		}

		return RegisterUserResponse{User: user.User{
			ID:    user.ID(usr.ID.String()),
			Login: user.Login(usr.Login),
		}}, nil
	}
}

// todo move to domain
func hashStringPassword(password string) (user.PasswordHash, error) {
	if isWeakPassword(password) {
		return "", ErrPasswordTooWeak
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return user.PasswordHash(hash), err
}

var (
	upperCase   = regexp.MustCompile("[A-Z]")
	lowerCase   = regexp.MustCompile("[a-z]")
	number      = regexp.MustCompile("[0-9]")
	specialChar = regexp.MustCompile("[!@#$%^&*]")
)

// isWeakPassword required the password to be:
// - 8 characters or longer
// - contain at least one lower case letter
// - contain at least one upper case letter
// - contain at least one number
// - contain at least one special character
func isWeakPassword(password string) bool {
	minPasswordLength := 8
	if len(password) < minPasswordLength {
		return true
	}

	matchRules := []*regexp.Regexp{upperCase, lowerCase, number, specialChar}
	mPW := []byte(password)

	for _, r := range matchRules {
		if !r.Match(mPW) {
			return true
		}
	}

	return false
}

type (
	ShowUserRequest struct {
		UserID user.ID
	}
	ShowUserResponse struct {
		User user.User
	}
)

func ShowUser(queries *models.Queries) func(context.Context, ShowUserRequest) (ShowUserResponse, error) {
	return func(ctx context.Context, in ShowUserRequest) (ShowUserResponse, error) {
		if in.UserID == "" {
			return ShowUserResponse{}, ErrInvalidInput
		}

		usr, err := repository.RepoGetUserByID(ctx, queries, in.UserID)
		if err != nil {
			return ShowUserResponse{}, fmt.Errorf("could not get user: %w", err)
		}

		return ShowUserResponse{User: usr}, nil
	}
}
