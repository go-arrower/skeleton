// Package auth is the intraprocess API of what this Context is exposing to other Contexts to use.
package auth

import (
	"context"
	"time"
)

const (
	RouteLogin       = "auth.login"
	RouteLogout      = "auth.logout"
	RouteVerifyEmail = ""
	RouteResetPW     = ""
)

type Tenant struct{}

type User struct {
	ID     string
	Tenant string
	Login  string // UserName

	FirstName         string
	LastName          string
	Name              string // DisplayName
	Birthday          string // make struct to prevent issues with tz or define format?
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
	IsAdmin       bool
	AdminSince    time.Time
}

type APIKey struct{}

// see CurrentUserID
func UserID(ctx context.Context) string { return "" } // or just ID()

func UserFromContext(ctx context.Context) User { // or just User()
	return User{}
}

func TenantFromContext(ctx context.Context) Tenant {
	return Tenant{}
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
