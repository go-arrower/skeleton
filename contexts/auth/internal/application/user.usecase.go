package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-arrower/arrower/alog"
	"github.com/go-arrower/arrower/jobs"
	"github.com/google/uuid"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"github.com/go-arrower/skeleton/contexts/auth/internal/infrastructure"
)

var (
	ErrLoginFailed  = errors.New("login failed")
	ErrInvalidInput = errors.New("invalid input")
)

type (
	LoginUserRequest struct { //nolint:govet // fieldalignment less important than grouping of params.
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
		IP         user.ResolvedIP
		Device     user.Device
		// Ip Location
	}
)

func LoginUser(
	logger alog.Logger,
	repo user.Repository,
	queue jobs.Enqueuer,
) func(context.Context, LoginUserRequest) (LoginUserResponse, error) {
	var ip user.IPResolver = infrastructure.NewIP2LocationService("")
	authenticator := user.NewAuthenticationService()

	return func(ctx context.Context, in LoginUserRequest) (LoginUserResponse, error) {
		usr, err := repo.FindByLogin(ctx, user.Login(in.LoginEmail))
		if err != nil {
			logger.Log(ctx, slog.LevelInfo, "login failed",
				slog.String("email", in.LoginEmail),
				slog.String("ip", in.IP),
			)

			return LoginUserResponse{}, ErrLoginFailed
		}

		if !authenticator.Authenticate(ctx, usr, in.Password) {
			logger.Log(ctx, slog.LevelInfo, "login failed",
				slog.String("email", in.LoginEmail),
				slog.String("ip", in.IP),
			)

			return LoginUserResponse{}, ErrLoginFailed
		}

		// The session is not valid until the end of the controller.
		// Thus, the session is created here and very short-lived, as the controller will update it with the right values.
		usr.Sessions = append(usr.Sessions, user.Session{
			ID:        in.SessionKey,
			Device:    user.NewDevice(in.UserAgent),
			CreatedAt: time.Now().UTC(),
			// ExpiresAt: // will be set & updated via the session store
		})

		err = repo.Save(ctx, usr)
		if err != nil {
			return LoginUserResponse{}, fmt.Errorf("could not update user session: %w", err)
		}
		// FIXME: add a method to user or a domain service, that ensures session is not added, if one with same ID already exists.

		if in.IsNewDevice {
			resolved, err := ip.ResolveIP(in.IP)
			if err != nil {
				return LoginUserResponse{}, fmt.Errorf("could not resolve ip address: %w", err)
			}

			err = queue.Enqueue(ctx, SendConfirmationNewDeviceLoggedIn{
				UserID:     usr.ID,
				OccurredAt: time.Now().UTC(),
				IP:         resolved,
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
	RegisterUserRequest struct { //nolint:govet // fieldalignment less important than grouping of params.
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

	NewUserVerificationEmail struct {
		UserID     user.ID
		OccurredAt time.Time
		IP         user.ResolvedIP
		Device     user.Device
	}
)

func RegisterUser(
	logger alog.Logger,
	repo user.Repository,
	registrator *user.RegistrationService,
	queue jobs.Enqueuer,
) func(context.Context, RegisterUserRequest) (RegisterUserResponse, error) {
	var ip user.IPResolver = infrastructure.NewIP2LocationService("")

	return func(ctx context.Context, in RegisterUserRequest) (RegisterUserResponse, error) {
		usr, err := registrator.RegisterNewUser(ctx, in.RegisterEmail, in.Password)
		if err != nil {
			if errors.Is(err, user.ErrUserAlreadyExists) {
				logger.Log(ctx, slog.LevelInfo, "register new user failed",
					slog.String("email", in.RegisterEmail),
					slog.String("ip", in.IP),
				)
			}

			return RegisterUserResponse{}, fmt.Errorf("%w", err)
		}

		// The session is not valid until the end of the controller.
		// Thus, the session is created here and very short-lived, as the controller will update it with the right values.
		usr.Sessions = append(usr.Sessions, user.Session{
			ID:        in.SessionKey,
			Device:    user.NewDevice(in.UserAgent),
			CreatedAt: time.Now().UTC(),
			// ExpiresAt: // will be set & updated via the session store
		})

		err = repo.Save(ctx, usr)
		if err != nil {
			return RegisterUserResponse{}, fmt.Errorf("could not save new user: %w", err)
		}

		resolved, err := ip.ResolveIP(in.IP)
		if err != nil {
			return RegisterUserResponse{}, fmt.Errorf("could not resolve ip address: %w", err)
		}

		// !!! CONSIDER !!! if the email output port is async (outbox pattern) call it directly instead of a job
		err = queue.Enqueue(ctx, NewUserVerificationEmail{
			UserID:     usr.ID,
			OccurredAt: time.Now().UTC(),
			IP:         resolved,
			Device:     user.NewDevice(in.UserAgent),
		})
		if err != nil {
			return RegisterUserResponse{}, fmt.Errorf("could not queue job to send verification email: %w", err)
		}

		// todo return a short "UserDescriptor" or something instead of a partial user.
		return RegisterUserResponse{User: user.User{ //nolint:exhaustruct // at this point the user has not more information.
			ID:    usr.ID,
			Login: usr.Login,
		}}, nil
	}
}

func SendNewUserVerificationEmail(
	logger alog.Logger,
	repo user.Repository,
) func(context.Context, NewUserVerificationEmail) error {
	return func(ctx context.Context, in NewUserVerificationEmail) error {
		usr, err := repo.FindByID(ctx, in.UserID)
		if err != nil {
			return fmt.Errorf("could not get user: %w", err)
		}

		verify := user.NewVerificationService(repo)

		token, err := verify.NewVerificationToken(ctx, usr)
		if err != nil {
			return fmt.Errorf("could not generate verification token: %w", err)
		}

		// later: instead of logging this => send it to an email output port
		logger.InfoContext(ctx, "send verification email to user",
			slog.String("token", token.Token().String()),
			slog.String("device", in.Device.Name()+" "+in.Device.OS()),
			slog.String("ip", in.IP.IP.String()),
			slog.String("time", in.OccurredAt.String()),
			slog.String("email", string(usr.Login)),
		)

		return nil
	}
}

type (
	VerifyUserRequest struct {
		UserID user.ID   `validate:"required"`
		Token  uuid.UUID `validate:"required"`
	}
)

func VerifyUser(repo user.Repository) func(context.Context, VerifyUserRequest) error {
	return func(ctx context.Context, in VerifyUserRequest) error {
		usr, err := repo.FindByID(ctx, in.UserID)
		if err != nil {
			return fmt.Errorf("could not get user: %w", err)
		}

		verify := user.NewVerificationService(repo)

		err = verify.Verify(ctx, &usr, in.Token)
		if err != nil {
			return fmt.Errorf("could not verify user: %w", err)
		}

		return nil
	}
}

type (
	ShowUserRequest struct {
		UserID user.ID
	}
	ShowUserResponse struct {
		User user.User
	}
)

func ShowUser(repo user.Repository) func(context.Context, ShowUserRequest) (ShowUserResponse, error) {
	return func(ctx context.Context, in ShowUserRequest) (ShowUserResponse, error) {
		if in.UserID == "" {
			return ShowUserResponse{}, ErrInvalidInput
		}

		usr, err := repo.FindByID(ctx, in.UserID)
		if err != nil {
			return ShowUserResponse{}, fmt.Errorf("could not get user: %w", err)
		}

		return ShowUserResponse{User: usr}, nil
	}
}

type (
	BlockUserRequest struct {
		UserID user.ID `validate:"required"`
	}
	BlockUserResponse struct {
		UserID  user.ID
		Blocked user.BoolFlag
	}
)

func BlockUser(repo user.Repository) func(context.Context, BlockUserRequest) (BlockUserResponse, error) {
	return func(ctx context.Context, in BlockUserRequest) (BlockUserResponse, error) {
		usr, err := repo.FindByID(ctx, in.UserID)
		if err != nil {
			return BlockUserResponse{}, fmt.Errorf("could not get user: %w", err)
		}

		usr.Block()

		err = repo.Save(ctx, usr)
		if err != nil {
			return BlockUserResponse{}, fmt.Errorf("could not get user: %w", err)
		}

		return BlockUserResponse{
			UserID:  usr.ID,
			Blocked: usr.Blocked,
		}, nil
	}
}

func UnblockUser(repo user.Repository) func(context.Context, BlockUserRequest) (BlockUserResponse, error) {
	return func(ctx context.Context, in BlockUserRequest) (BlockUserResponse, error) {
		usr, err := repo.FindByID(ctx, in.UserID)
		if err != nil {
			return BlockUserResponse{}, fmt.Errorf("could not get user: %w", err)
		}

		usr.Unblock()

		err = repo.Save(ctx, usr)
		if err != nil {
			return BlockUserResponse{}, fmt.Errorf("could not get user: %w", err)
		}

		return BlockUserResponse{
			UserID:  usr.ID,
			Blocked: usr.Blocked,
		}, nil
	}
}
