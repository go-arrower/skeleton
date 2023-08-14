// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: query.sql

package models

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const allSessions = `-- name: AllSessions :many

SELECT key, data, expires_at, user_id, user_agent, created_at, updated_at
FROM auth.session
ORDER BY created_at ASC
`

// -------------------
// ---- Session ------
// -------------------
func (q *Queries) AllSessions(ctx context.Context) ([]AuthSession, error) {
	rows, err := q.db.Query(ctx, allSessions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AuthSession
	for rows.Next() {
		var i AuthSession
		if err := rows.Scan(
			&i.Key,
			&i.Data,
			&i.ExpiresAt,
			&i.UserID,
			&i.UserAgent,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const allUsers = `-- name: AllUsers :many

SELECT id, created_at, updated_at, login, password_hash, first_name, last_name, name, birthday, locale, time_zone, picture_url, profile, verified_at, blocked_at, super_user_at
FROM auth.user
ORDER BY login
`

// ----------------
// ---- User ------
// ----------------
func (q *Queries) AllUsers(ctx context.Context) ([]AuthUser, error) {
	rows, err := q.db.Query(ctx, allUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AuthUser
	for rows.Next() {
		var i AuthUser
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Login,
			&i.PasswordHash,
			&i.FirstName,
			&i.LastName,
			&i.Name,
			&i.Birthday,
			&i.Locale,
			&i.TimeZone,
			&i.PictureUrl,
			&i.Profile,
			&i.VerifiedAt,
			&i.BlockedAt,
			&i.SuperUserAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createUser = `-- name: CreateUser :one
INSERT INTO auth.user (id, login, password_hash, verified_at, blocked_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at, login, password_hash, first_name, last_name, name, birthday, locale, time_zone, picture_url, profile, verified_at, blocked_at, super_user_at
`

type CreateUserParams struct {
	ID           uuid.UUID
	Login        string
	PasswordHash string
	VerifiedAt   pgtype.Timestamptz
	BlockedAt    pgtype.Timestamptz
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (AuthUser, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.ID,
		arg.Login,
		arg.PasswordHash,
		arg.VerifiedAt,
		arg.BlockedAt,
	)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Login,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Name,
		&i.Birthday,
		&i.Locale,
		&i.TimeZone,
		&i.PictureUrl,
		&i.Profile,
		&i.VerifiedAt,
		&i.BlockedAt,
		&i.SuperUserAt,
	)
	return i, err
}

const deleteSessionByKey = `-- name: DeleteSessionByKey :exec
DELETE
FROM auth.session
WHERE key = $1
`

func (q *Queries) DeleteSessionByKey(ctx context.Context, key []byte) error {
	_, err := q.db.Exec(ctx, deleteSessionByKey, key)
	return err
}

const findSessionDataByKey = `-- name: FindSessionDataByKey :one
SELECT data
FROM auth.session
WHERE key = $1
`

func (q *Queries) FindSessionDataByKey(ctx context.Context, key []byte) ([]byte, error) {
	row := q.db.QueryRow(ctx, findSessionDataByKey, key)
	var data []byte
	err := row.Scan(&data)
	return data, err
}

const findSessionsByUserID = `-- name: FindSessionsByUserID :many
SELECT key, data, expires_at, user_id, user_agent, created_at, updated_at
FROM auth.session
WHERE user_id = $1
ORDER BY created_at
`

func (q *Queries) FindSessionsByUserID(ctx context.Context, userID uuid.NullUUID) ([]AuthSession, error) {
	rows, err := q.db.Query(ctx, findSessionsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AuthSession
	for rows.Next() {
		var i AuthSession
		if err := rows.Scan(
			&i.Key,
			&i.Data,
			&i.ExpiresAt,
			&i.UserID,
			&i.UserAgent,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findUserByID = `-- name: FindUserByID :one
SELECT id, created_at, updated_at, login, password_hash, first_name, last_name, name, birthday, locale, time_zone, picture_url, profile, verified_at, blocked_at, super_user_at
FROM auth.user
WHERE id = $1
`

func (q *Queries) FindUserByID(ctx context.Context, id uuid.UUID) (AuthUser, error) {
	row := q.db.QueryRow(ctx, findUserByID, id)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Login,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Name,
		&i.Birthday,
		&i.Locale,
		&i.TimeZone,
		&i.PictureUrl,
		&i.Profile,
		&i.VerifiedAt,
		&i.BlockedAt,
		&i.SuperUserAt,
	)
	return i, err
}

const findUserByLogin = `-- name: FindUserByLogin :one
SELECT id, created_at, updated_at, login, password_hash, first_name, last_name, name, birthday, locale, time_zone, picture_url, profile, verified_at, blocked_at, super_user_at
FROM auth.user
WHERE login = $1
`

func (q *Queries) FindUserByLogin(ctx context.Context, login string) (AuthUser, error) {
	row := q.db.QueryRow(ctx, findUserByLogin, login)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Login,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Name,
		&i.Birthday,
		&i.Locale,
		&i.TimeZone,
		&i.PictureUrl,
		&i.Profile,
		&i.VerifiedAt,
		&i.BlockedAt,
		&i.SuperUserAt,
	)
	return i, err
}

const upsertNewSession = `-- name: UpsertNewSession :exec
INSERT INTO auth.session (key, user_id, user_agent)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET (user_id, user_agent) = ($2, $3)
`

type UpsertNewSessionParams struct {
	Key       []byte
	UserID    uuid.NullUUID
	UserAgent string
}

func (q *Queries) UpsertNewSession(ctx context.Context, arg UpsertNewSessionParams) error {
	_, err := q.db.Exec(ctx, upsertNewSession, arg.Key, arg.UserID, arg.UserAgent)
	return err
}

const upsertSessionData = `-- name: UpsertSessionData :exec
INSERT INTO auth.session (key, data, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET (data, expires_at) = ($2, $3)
`

type UpsertSessionDataParams struct {
	Key       []byte
	Data      []byte
	ExpiresAt pgtype.Timestamptz
}

func (q *Queries) UpsertSessionData(ctx context.Context, arg UpsertSessionDataParams) error {
	_, err := q.db.Exec(ctx, upsertSessionData, arg.Key, arg.Data, arg.ExpiresAt)
	return err
}

const upsertUser = `-- name: UpsertUser :one
INSERT INTO auth.user(id, created_at, login, password_hash, first_name, last_name, name, birthday, locale, time_zone,
                      picture_url, profile, verified_at, blocked_at, super_user_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (id) DO UPDATE SET (login, password_hash, first_name, last_name, name, birthday, locale, time_zone,
                                picture_url, profile, verified_at, blocked_at, super_user_at) = ($3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING id, created_at, updated_at, login, password_hash, first_name, last_name, name, birthday, locale, time_zone, picture_url, profile, verified_at, blocked_at, super_user_at
`

type UpsertUserParams struct {
	ID           uuid.UUID
	CreatedAt    pgtype.Timestamptz
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

func (q *Queries) UpsertUser(ctx context.Context, arg UpsertUserParams) (AuthUser, error) {
	row := q.db.QueryRow(ctx, upsertUser,
		arg.ID,
		arg.CreatedAt,
		arg.Login,
		arg.PasswordHash,
		arg.FirstName,
		arg.LastName,
		arg.Name,
		arg.Birthday,
		arg.Locale,
		arg.TimeZone,
		arg.PictureUrl,
		arg.Profile,
		arg.VerifiedAt,
		arg.BlockedAt,
		arg.SuperUserAt,
	)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Login,
		&i.PasswordHash,
		&i.FirstName,
		&i.LastName,
		&i.Name,
		&i.Birthday,
		&i.Locale,
		&i.TimeZone,
		&i.PictureUrl,
		&i.Profile,
		&i.VerifiedAt,
		&i.BlockedAt,
		&i.SuperUserAt,
	)
	return i, err
}
