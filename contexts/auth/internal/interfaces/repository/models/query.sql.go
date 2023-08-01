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
INSERT INTO auth.user (login, password_hash, verified_at)
VALUES ($1, $2, $3)
RETURNING id, created_at, updated_at, login, password_hash, first_name, last_name, name, birthday, locale, time_zone, picture_url, profile, verified_at, blocked_at, super_user_at
`

type CreateUserParams struct {
	Login        string
	PasswordHash string
	VerifiedAt   pgtype.Timestamptz
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (AuthUser, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Login, arg.PasswordHash, arg.VerifiedAt)
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
