package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"

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
		User User
	}
)

func LoginUser(queries *models.Queries) func(context.Context, LoginUserRequest) (LoginUserResponse, error) {
	return func(ctx context.Context, in LoginUserRequest) (LoginUserResponse, error) {
		user, err := queries.FindUserByLogin(ctx, in.LoginEmail)
		if err != nil {
			return LoginUserResponse{}, ErrLoginFailed
		}

		hash := PasswordHash(user.UserPasswordHash)
		if !hash.Matches(in.Password) {
			return LoginUserResponse{}, ErrLoginFailed
		}

		u := User{
			ID:    ID(user.ID.String()),
			Login: Login(user.UserLogin),
			// todo mapping
		}
		return LoginUserResponse{User: u}, nil
	}
}

type (
	RegisterUserRequest struct {
		RegisterEmail        string `form:"login" validate:"max=1024,required,email"`
		Password             string `form:"password" validate:"max=1024,min=8"`
		PasswordConfirmation string `form:"password_confirmation" validate:"max=1024,eqfield=Password"`
	}
	RegisterUserResponse struct {
		User User
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

		user := User{
			ID:           NewID(),
			Name:         "",
			Login:        Login(in.RegisterEmail),
			PasswordHash: pwHash,
			IsVerified:   Verified{},
			IsBlocked:    BlockedFlag{},
			IsAdmin:      Admin{},
			Profile:      nil,
		}

		// TODO take the user and persist is completely
		_, err = queries.CreateUser(ctx, models.CreateUserParams{
			UserLogin:        string(user.Login),
			UserPasswordHash: string(user.PasswordHash),
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
func hashStringPassword(password string) (PasswordHash, error) {
	if isWeakPassword(password) {
		return "", ErrPasswordTooWeak
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return PasswordHash(hash), err
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
	ID                string
	Login             string
	Email             string
	PasswordHash      string
	Verified          time.Time
	BlockedFlag       time.Time
	Admin             time.Time
	VerificationToken string
	User              struct {
		ID           ID
		Name         string
		Login        Login // email, or phone, or nickname, or whatever the developer wants to have as a login
		PasswordHash PasswordHash
		IsVerified   Verified
		IsBlocked    BlockedFlag
		IsAdmin      Admin
		Profile      map[string]string // a quick helper for simple stuff, if you have a complicated profile => do it in your Context, as it's the better place
		// TenantID tenant.ID
	}
	UserRegistered struct {
		ID         ID
		RecordedAt time.Time
	}
)

func (pw PasswordHash) Matches(checkPW string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(string(pw)), []byte(checkPW)); err == nil {
		return true
	}

	return false
}
func (pw PasswordHash) String() string { return "xxxxxx" }

func (t Verified) IsVerified() bool   { return false }
func (t BlockedFlag) IsBlocked() bool { return false }
func (t Admin) IsAdmin() bool         { return false }

func NewUser(...any) User { return User{} }

func NewID() ID {
	return ID(uuid.NewString())
}
