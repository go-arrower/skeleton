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

SELECT key, data, expires_at_utc, user_id, user_agent, created_at, updated_at
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
			&i.ExpiresAtUtc,
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

SELECT id, created_at, updated_at, login, password_hash, name_firstname, name_lastname, name_displayname, birthday, locale, time_zone, picture_url, profile, verified_at_utc, blocked_at_utc, superuser_at_utc
FROM auth.user
WHERE TRUE
     AND (CASE WHEN $2::TEXT <> '' THEN $2 < login ELSE TRUE END)
ORDER BY login
LIMIT $1
`

type AllUsersParams struct {
	Limit int32
	Login string
}

// ----------------
// ---- User ------
// ----------------
func (q *Queries) AllUsers(ctx context.Context, arg AllUsersParams) ([]AuthUser, error) {
	rows, err := q.db.Query(ctx, allUsers, arg.Limit, arg.Login)
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
			&i.NameFirstname,
			&i.NameLastname,
			&i.NameDisplayname,
			&i.Birthday,
			&i.Locale,
			&i.TimeZone,
			&i.PictureUrl,
			&i.Profile,
			&i.VerifiedAtUtc,
			&i.BlockedAtUtc,
			&i.SuperuserAtUtc,
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

const allUsersByIDs = `-- name: AllUsersByIDs :many
SELECT id, created_at, updated_at, login, password_hash, name_firstname, name_lastname, name_displayname, birthday, locale, time_zone, picture_url, profile, verified_at_utc, blocked_at_utc, superuser_at_utc
FROM auth.user
WHERE id = ANY ($1::uuid[])
`

func (q *Queries) AllUsersByIDs(ctx context.Context, dollar_1 []uuid.UUID) ([]AuthUser, error) {
	rows, err := q.db.Query(ctx, allUsersByIDs, dollar_1)
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
			&i.NameFirstname,
			&i.NameLastname,
			&i.NameDisplayname,
			&i.Birthday,
			&i.Locale,
			&i.TimeZone,
			&i.PictureUrl,
			&i.Profile,
			&i.VerifiedAtUtc,
			&i.BlockedAtUtc,
			&i.SuperuserAtUtc,
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

const countUsers = `-- name: CountUsers :one
SELECT COUNT(*)
FROM auth.user
`

func (q *Queries) CountUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countUsers)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createUser = `-- name: CreateUser :one
INSERT
INTO auth.user (id, login, password_hash, verified_at_utc, blocked_at_utc)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at, login, password_hash, name_firstname, name_lastname, name_displayname, birthday, locale, time_zone, picture_url, profile, verified_at_utc, blocked_at_utc, superuser_at_utc
`

type CreateUserParams struct {
	ID            uuid.UUID
	Login         string
	PasswordHash  string
	VerifiedAtUtc pgtype.Timestamptz
	BlockedAtUtc  pgtype.Timestamptz
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (AuthUser, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.ID,
		arg.Login,
		arg.PasswordHash,
		arg.VerifiedAtUtc,
		arg.BlockedAtUtc,
	)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Login,
		&i.PasswordHash,
		&i.NameFirstname,
		&i.NameLastname,
		&i.NameDisplayname,
		&i.Birthday,
		&i.Locale,
		&i.TimeZone,
		&i.PictureUrl,
		&i.Profile,
		&i.VerifiedAtUtc,
		&i.BlockedAtUtc,
		&i.SuperuserAtUtc,
	)
	return i, err
}

const createVerificationToken = `-- name: CreateVerificationToken :exec
INSERT INTO auth.user_verification(token, user_id, valid_until_utc)
VALUES ($1, $2, $3)
`

type CreateVerificationTokenParams struct {
	Token         uuid.UUID
	UserID        uuid.UUID
	ValidUntilUtc pgtype.Timestamptz
}

func (q *Queries) CreateVerificationToken(ctx context.Context, arg CreateVerificationTokenParams) error {
	_, err := q.db.Exec(ctx, createVerificationToken, arg.Token, arg.UserID, arg.ValidUntilUtc)
	return err
}

const deleteAllUsers = `-- name: DeleteAllUsers :exec
DELETE
FROM auth.user
`

func (q *Queries) DeleteAllUsers(ctx context.Context) error {
	_, err := q.db.Exec(ctx, deleteAllUsers)
	return err
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

const deleteUser = `-- name: DeleteUser :exec
DELETE
FROM auth.user
WHERE id = ANY ($1::uuid[])
`

func (q *Queries) DeleteUser(ctx context.Context, dollar_1 []uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteUser, dollar_1)
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
SELECT key, data, expires_at_utc, user_id, user_agent, created_at, updated_at
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
			&i.ExpiresAtUtc,
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
SELECT id, created_at, updated_at, login, password_hash, name_firstname, name_lastname, name_displayname, birthday, locale, time_zone, picture_url, profile, verified_at_utc, blocked_at_utc, superuser_at_utc
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
		&i.NameFirstname,
		&i.NameLastname,
		&i.NameDisplayname,
		&i.Birthday,
		&i.Locale,
		&i.TimeZone,
		&i.PictureUrl,
		&i.Profile,
		&i.VerifiedAtUtc,
		&i.BlockedAtUtc,
		&i.SuperuserAtUtc,
	)
	return i, err
}

const findUserByLogin = `-- name: FindUserByLogin :one
SELECT id, created_at, updated_at, login, password_hash, name_firstname, name_lastname, name_displayname, birthday, locale, time_zone, picture_url, profile, verified_at_utc, blocked_at_utc, superuser_at_utc
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
		&i.NameFirstname,
		&i.NameLastname,
		&i.NameDisplayname,
		&i.Birthday,
		&i.Locale,
		&i.TimeZone,
		&i.PictureUrl,
		&i.Profile,
		&i.VerifiedAtUtc,
		&i.BlockedAtUtc,
		&i.SuperuserAtUtc,
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
INSERT INTO auth.session (key, data, expires_at_utc)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE SET (data, expires_at_utc) = ($2, $3)
`

type UpsertSessionDataParams struct {
	Key          []byte
	Data         []byte
	ExpiresAtUtc pgtype.Timestamptz
}

func (q *Queries) UpsertSessionData(ctx context.Context, arg UpsertSessionDataParams) error {
	_, err := q.db.Exec(ctx, upsertSessionData, arg.Key, arg.Data, arg.ExpiresAtUtc)
	return err
}

const upsertUser = `-- name: UpsertUser :one
INSERT INTO auth.user(id, created_at, login, password_hash, name_firstname, name_lastname, name_displayname, birthday,
                      locale, time_zone,
                      picture_url, profile, verified_at_utc, blocked_at_utc, superuser_at_utc)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (id) DO UPDATE SET (login, password_hash, name_firstname, name_lastname, name_displayname, birthday, locale,
                                time_zone,
                                picture_url, profile, verified_at_utc, blocked_at_utc,
                                superuser_at_utc) = ($3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING id, created_at, updated_at, login, password_hash, name_firstname, name_lastname, name_displayname, birthday, locale, time_zone, picture_url, profile, verified_at_utc, blocked_at_utc, superuser_at_utc
`

type UpsertUserParams struct {
	ID              uuid.UUID
	CreatedAt       pgtype.Timestamptz
	Login           string
	PasswordHash    string
	NameFirstname   string
	NameLastname    string
	NameDisplayname string
	Birthday        pgtype.Date
	Locale          string
	TimeZone        string
	PictureUrl      string
	Profile         pgtype.Hstore
	VerifiedAtUtc   pgtype.Timestamptz
	BlockedAtUtc    pgtype.Timestamptz
	SuperuserAtUtc  pgtype.Timestamptz
}

func (q *Queries) UpsertUser(ctx context.Context, arg UpsertUserParams) (AuthUser, error) {
	row := q.db.QueryRow(ctx, upsertUser,
		arg.ID,
		arg.CreatedAt,
		arg.Login,
		arg.PasswordHash,
		arg.NameFirstname,
		arg.NameLastname,
		arg.NameDisplayname,
		arg.Birthday,
		arg.Locale,
		arg.TimeZone,
		arg.PictureUrl,
		arg.Profile,
		arg.VerifiedAtUtc,
		arg.BlockedAtUtc,
		arg.SuperuserAtUtc,
	)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Login,
		&i.PasswordHash,
		&i.NameFirstname,
		&i.NameLastname,
		&i.NameDisplayname,
		&i.Birthday,
		&i.Locale,
		&i.TimeZone,
		&i.PictureUrl,
		&i.Profile,
		&i.VerifiedAtUtc,
		&i.BlockedAtUtc,
		&i.SuperuserAtUtc,
	)
	return i, err
}

const userExistsByID = `-- name: UserExistsByID :one
SELECT EXISTS(SELECT 1 FROM auth.user WHERE id = $1) AS "exists"
`

func (q *Queries) UserExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	row := q.db.QueryRow(ctx, userExistsByID, id)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const userExistsByLogin = `-- name: UserExistsByLogin :one
SELECT EXISTS(SELECT 1 FROM auth.user WHERE login = $1) AS "exists"
`

func (q *Queries) UserExistsByLogin(ctx context.Context, login string) (bool, error) {
	row := q.db.QueryRow(ctx, userExistsByLogin, login)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const verificationTokenByToken = `-- name: VerificationTokenByToken :one
SELECT token, user_id, valid_until_utc, created_at, updated_at
FROM auth.user_verification
WHERE token = $1
`

func (q *Queries) VerificationTokenByToken(ctx context.Context, token uuid.UUID) (AuthUserVerification, error) {
	row := q.db.QueryRow(ctx, verificationTokenByToken, token)
	var i AuthUserVerification
	err := row.Scan(
		&i.Token,
		&i.UserID,
		&i.ValidUntilUtc,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
