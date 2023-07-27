// Package auth is the intraprocess API of what this Context is exposing to other Contexts to use.
package auth

import (
	"context"
	"time"
)

type TenantID string

type Tenant struct{}

type UserID string

type User struct {
	ID            UserID
	Tenant        TenantID
	Login         string
	Name          string
	IsVerified    bool
	VerifiedSince time.Time
	IsBlocked     bool
	BlockedSince  time.Time
	IsAdmin       bool
	AdminSince    time.Time
	Profile       map[string]string
}

type APIKey struct{}

func UserFromContext(ctx context.Context) User {
	return User{}
}
func TenantFromContext(ctx context.Context) Tenant {
	return Tenant{}
}

func IsLoggedInAsOtherUser(ctx context.Context) bool {
	return false
}
