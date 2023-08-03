// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthSession struct {
	Key       []byte
	Data      []byte
	ExpiresAt pgtype.Timestamptz
	UserID    uuid.NullUUID
	UserAgent string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type AuthUser struct {
	ID           uuid.UUID
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	Login        string
	PasswordHash string
	FirstName    string
	LastName     string
	Name         string
	Birthday     pgtype.Date
	Locale       string
	TimeZone     string
	PictureUrl   string
	Profile      pgtype.Hstore
	VerifiedAt   pgtype.Timestamptz
	BlockedAt    pgtype.Timestamptz
	SuperUserAt  pgtype.Timestamptz
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
