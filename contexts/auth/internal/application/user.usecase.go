package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/go-arrower/skeleton/contexts/auth/internal/application/user"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-arrower/skeleton/contexts/auth/internal/interfaces/repository/models"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrPasswordTooWeak   = errors.New("password too weak")
	ErrLoginFailed       = errors.New("login failed")
)

type (
	LoginUserRequest struct {
		LoginEmail string `form:"login" validate:"max=1024,required,email"`
		Password   string `form:"password" validate:"max=1024,min=8"`
	}
	LoginUserResponse struct {
		User user.User
	}
)

func LoginUser(queries *models.Queries) func(context.Context, LoginUserRequest) (LoginUserResponse, error) {
	return func(ctx context.Context, in LoginUserRequest) (LoginUserResponse, error) {
		user, err := repoGetUserByLogin(ctx, queries, in.LoginEmail)
		if err != nil {
			return LoginUserResponse{}, ErrLoginFailed
		}

		if !user.Verified.IsVerified() {
			return LoginUserResponse{}, ErrLoginFailed
		}

		if !user.PasswordHash.Matches(in.Password) {
			return LoginUserResponse{}, ErrLoginFailed
		}

		return LoginUserResponse{User: user}, nil
	}
}

func repoGetUserByLogin(ctx context.Context, queries *models.Queries, loginEmail string) (user.User, error) {
	u, err := queries.FindUserByLogin(ctx, loginEmail)
	if err != nil {
		return user.User{}, ErrLoginFailed
	}

	var p = make(map[string]*string)
	profile := u.Profile.Scan(&p)
	_ = profile
	_ = p
	_ = u.Profile.Value

	return user.User{
		ID:                user.ID(u.ID.String()),
		Login:             user.Login(u.Login),
		PasswordHash:      user.PasswordHash(u.PasswordHash),
		RegisteredAt:      u.CreatedAt.Time,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		Name:              u.Name,
		Birthday:          user.Birthday{}, //todo
		Locale:            user.Locale{},   //todo
		TimeZone:          user.TimeZone(u.TimeZone),
		ProfilePictureURL: user.URL(u.PictureUrl),
		Profile2:          p, //todo
		Verified:          user.VerifiedFlag(u.VerifiedAt.Time),
		Blocked:           user.BlockedFlag(u.BlockedAt.Time),
		SuperUser:         user.SuperUserFlag(u.SuperUserAt.Time),
	}, nil
}

type (
	RegisterUserRequest struct {
		RegisterEmail          string `form:"login" validate:"max=1024,required,email"`
		Password               string `form:"password" validate:"max=1024,min=8"`
		PasswordConfirmation   string `form:"password_confirmation" validate:"max=1024,eqfield=Password"`
		AcceptedTermsOfService bool   `form:"toc" validate:"required"`
	}
	RegisterUserResponse struct {
		User user.User
	}
)

func RegisterUser(queries *models.Queries) func(context.Context, RegisterUserRequest) (RegisterUserResponse, error) {
	return func(ctx context.Context, in RegisterUserRequest) (RegisterUserResponse, error) {
		//if !mw.PassedValidation(ctx) { /* validate OR return err */ }

		if _, err := queries.FindUserByLogin(ctx, in.RegisterEmail); err == nil {
			return RegisterUserResponse{}, ErrUserAlreadyExists
		}

		pwHash, err := hashStringPassword(in.Password)
		if err != nil {
			return RegisterUserResponse{}, err
		}

		// TODO  Gather metadata: device info, location, timezone?

		user := user.User{
			ID:           user.NewID(),
			Name:         "",
			Login:        user.Login(in.RegisterEmail),
			PasswordHash: pwHash,
			Verified:     user.VerifiedFlag{},
			Blocked:      user.BlockedFlag{},
			SuperUser:    user.SuperUserFlag{},
			Profile:      nil,
		}

		// TODO take the user and persist is completely
		_, err = queries.CreateUser(ctx, models.CreateUserParams{
			Login:        string(user.Login),
			PasswordHash: string(user.PasswordHash),
		})
		if err != nil {
			return RegisterUserResponse{}, fmt.Errorf("could not create user: %w", err)
		}

		/*
		* Send activation message as job (email or sms or nothing depending on the login type)
		* Emit event of new user
		 */

		return RegisterUserResponse{}, nil
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
