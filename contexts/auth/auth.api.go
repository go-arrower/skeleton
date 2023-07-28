// Package auth is the intraprocess API of what this Context is exposing to other Contexts to use.
package auth

import (
	"context"
	"time"
)

// API is the api of the auth Context.
//
// SHOULD IT BE MORE TYPED? UserID instead of string, Credentials instead of string,string pair?
type API interface {
	User(ctx context.Context) User
	All() ([]User, error)
	UserByID(id string) (User, error)
	UserByLogin(login string) (User, error)
	Register(info ...any) (User, error)
	Validate(id string, token string) error
	Authenticate(username string, password string) (bool, error)
	Logout(id string) error
	//ResetPW
}

const (
	RouteLogin       = "auth.login"
	RouteLogout      = "auth.logout"
	RouteVerifyEmail = ""
	RouteResetPW     = ""
)

type User struct {
	ID    string
	Login string // UserName

	FirstName         string
	LastName          string
	Name              string // DisplayName
	Birthday          string // make struct to prevent issues with tz or define format? // TYPES OR PLAIN?
	Locale            string
	TimeZone          string
	ProfilePictureURL string
	Data              map[string]string // limit the length of keys & values // { plan: 'silver', team_id: 'a111' }
	// nickname, gender, email, phone, website???

	RegisteredAt  time.Time
	IsVerified    bool
	VerifiedSince time.Time
	IsBlocked     bool
	BlockedSince  time.Time
}

type APIKey struct{}

// see CurrentUserID
func UserID(ctx context.Context) string { return "" } // or just ID()

func UserFromContext(ctx context.Context) User { // or just User()
	return User{}
}

func IsLoggedInAsOtherUser(ctx context.Context) bool {
	return false
}

// --- --- ---
// methods are part of auth api and not static auth package:

// if develoepr wants to do the auth himself, instead of the web route
func Authenticate(cred any) (worked bool, validationErrs error) { return false, nil }

func Logout(userID any) bool { return false }

// --- --- ---
// events emitted by this Context

/*
	- RegisteredUser
	- AuthenticationAttempt
	- Authenticated
 	- SuccessfulLogin
	- FailedLogin
	- Verified
	- SuccessfulLogout
	- CurrentDeviceLogout
	- OtherDeviceLogout
	- PasswordReset
*/
