// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthCredential string

const (
	AuthCredentialUser AuthCredential = "user"
	AuthCredentialApi  AuthCredential = "api"
)

func (e *AuthCredential) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AuthCredential(s)
	case string:
		*e = AuthCredential(s)
	default:
		return fmt.Errorf("unsupported scan type for AuthCredential: %T", src)
	}
	return nil
}

type NullAuthCredential struct {
	AuthCredential AuthCredential
	Valid          bool // Valid is true if AuthCredential is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAuthCredential) Scan(value interface{}) error {
	if value == nil {
		ns.AuthCredential, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AuthCredential.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAuthCredential) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.AuthCredential), nil
}

type AuthSession struct {
	ID           int64
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	UserID       uuid.UUID
	Key          []byte
	Data         []byte
	ExpiresOn    pgtype.Timestamptz
	LastDevice   string
	LastLocation string
	LastTimezone string
}

type AuthTenant struct {
	ID        uuid.UUID
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	Name      string
	IsActive  bool
}

type AuthUser struct {
	ID                  uuid.UUID
	CreatedAt           pgtype.Timestamptz
	UpdatedAt           pgtype.Timestamptz
	CredentialType      AuthCredential
	IsActive            bool
	UserLogin           string
	UserPasswordHash    string
	UserLoginVerifiedAt pgtype.Timestamptz
	ApiName             string
	ApiKeyPrefix        string
	IsAdmin             bool
}

type GueJob struct {
	JobID      string
	Priority   int16
	RunAt      pgtype.Timestamptz
	JobType    string
	Args       []byte
	ErrorCount int32
	LastError  pgtype.Text
	Queue      string
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
}

type GueJobsHistory struct {
	JobID      string
	Priority   int16
	RunAt      pgtype.Timestamptz
	JobType    string
	Args       []byte
	Queue      string
	RunCount   int32
	RunError   pgtype.Text
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	Success    bool
	FinishedAt pgtype.Timestamptz
}

type GueJobsWorkerPool struct {
	ID        string
	Queue     string
	Workers   int16
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}
