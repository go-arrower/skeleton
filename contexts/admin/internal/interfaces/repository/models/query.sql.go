// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: query.sql

package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteJob = `-- name: DeleteJob :exec
DELETE
FROM arrower.gue_jobs
WHERE job_id = $1
`

func (q *Queries) DeleteJob(ctx context.Context, jobID string) error {
	_, err := q.db.Exec(ctx, deleteJob, jobID)
	return err
}

const getPendingJobs = `-- name: GetPendingJobs :many
SELECT job_id, priority, run_at, job_type, args, error_count, last_error, queue, created_at, updated_at
FROM arrower.gue_jobs
WHERE queue = $1
ORDER BY priority, run_at ASC
LIMIT 100
`

func (q *Queries) GetPendingJobs(ctx context.Context, queue string) ([]ArrowerGueJob, error) {
	rows, err := q.db.Query(ctx, getPendingJobs, queue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ArrowerGueJob
	for rows.Next() {
		var i ArrowerGueJob
		if err := rows.Scan(
			&i.JobID,
			&i.Priority,
			&i.RunAt,
			&i.JobType,
			&i.Args,
			&i.ErrorCount,
			&i.LastError,
			&i.Queue,
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

const getQueues = `-- name: GetQueues :many
SELECT queue
FROM arrower.gue_jobs
UNION
SELECT queue
FROM arrower.gue_jobs_history
`

func (q *Queries) GetQueues(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, getQueues)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var queue string
		if err := rows.Scan(&queue); err != nil {
			return nil, err
		}
		items = append(items, queue)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getWorkerPools = `-- name: GetWorkerPools :many
SELECT id, queue, workers, version, job_types, created_at, updated_at
FROM arrower.gue_jobs_worker_pool
WHERE updated_at > NOW() - INTERVAL '2 minutes'
ORDER BY queue, id
`

func (q *Queries) GetWorkerPools(ctx context.Context) ([]ArrowerGueJobsWorkerPool, error) {
	rows, err := q.db.Query(ctx, getWorkerPools)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ArrowerGueJobsWorkerPool
	for rows.Next() {
		var i ArrowerGueJobsWorkerPool
		if err := rows.Scan(
			&i.ID,
			&i.Queue,
			&i.Workers,
			&i.Version,
			&i.JobTypes,
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

const jobHistoryPayloadSize = `-- name: JobHistoryPayloadSize :one
SELECT COALESCE(pg_size_pretty(SUM(pg_column_size(arrower.gue_jobs_history.args))), '')
FROM arrower.gue_jobs_history
WHERE queue = $1
  AND created_at <= $2
  AND args <> ''
`

type JobHistoryPayloadSizeParams struct {
	Queue     string
	CreatedAt pgtype.Timestamptz
}

func (q *Queries) JobHistoryPayloadSize(ctx context.Context, arg JobHistoryPayloadSizeParams) (interface{}, error) {
	row := q.db.QueryRow(ctx, jobHistoryPayloadSize, arg.Queue, arg.CreatedAt)
	var coalesce interface{}
	err := row.Scan(&coalesce)
	return coalesce, err
}

const jobHistorySize = `-- name: JobHistorySize :one
SELECT COALESCE(pg_size_pretty(SUM(pg_column_size(arrower.gue_jobs_history.*))), '')
FROM arrower.gue_jobs_history
WHERE created_at <= $1
`

func (q *Queries) JobHistorySize(ctx context.Context, createdAt pgtype.Timestamptz) (interface{}, error) {
	row := q.db.QueryRow(ctx, jobHistorySize, createdAt)
	var coalesce interface{}
	err := row.Scan(&coalesce)
	return coalesce, err
}

const jobTableSize = `-- name: JobTableSize :one
SELECT pg_size_pretty(pg_total_relation_size('arrower.gue_jobs'))         as jobs,
       pg_size_pretty(pg_total_relation_size('arrower.gue_jobs_history')) as history
`

type JobTableSizeRow struct {
	Jobs    string
	History string
}

func (q *Queries) JobTableSize(ctx context.Context) (JobTableSizeRow, error) {
	row := q.db.QueryRow(ctx, jobTableSize)
	var i JobTableSizeRow
	err := row.Scan(&i.Jobs, &i.History)
	return i, err
}

const jobTypes = `-- name: JobTypes :many
SELECT DISTINCT(job_type)
FROM arrower.gue_jobs_history
WHERE queue = $1
`

func (q *Queries) JobTypes(ctx context.Context, queue string) ([]string, error) {
	rows, err := q.db.Query(ctx, jobTypes, queue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var job_type string
		if err := rows.Scan(&job_type); err != nil {
			return nil, err
		}
		items = append(items, job_type)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const lastHistoryPayloads = `-- name: LastHistoryPayloads :many
SELECT args
FROM arrower.gue_jobs_history
WHERE queue = $1
  AND job_type = $2
ORDER BY created_at DESC
LIMIT 5
`

type LastHistoryPayloadsParams struct {
	Queue   string
	JobType string
}

func (q *Queries) LastHistoryPayloads(ctx context.Context, arg LastHistoryPayloadsParams) ([][]byte, error) {
	rows, err := q.db.Query(ctx, lastHistoryPayloads, arg.Queue, arg.JobType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items [][]byte
	for rows.Next() {
		var args []byte
		if err := rows.Scan(&args); err != nil {
			return nil, err
		}
		items = append(items, args)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const pendingJobs = `-- name: PendingJobs :many
SELECT bins.t, COUNT(t)
FROM (SELECT date_bin($1, finished_at, TIMESTAMP WITH TIME ZONE'2001-01-01')::TIMESTAMPTZ as t
      FROM arrower.gue_jobs_history
      WHERE finished_at > $2) bins
GROUP BY bins.t
ORDER BY bins.t DESC
LIMIT $3
`

type PendingJobsParams struct {
	DateBin    pgtype.Interval
	FinishedAt pgtype.Timestamptz
	Limit      int32
}

type PendingJobsRow struct {
	T     pgtype.Timestamptz
	Count int64
}

func (q *Queries) PendingJobs(ctx context.Context, arg PendingJobsParams) ([]PendingJobsRow, error) {
	rows, err := q.db.Query(ctx, pendingJobs, arg.DateBin, arg.FinishedAt, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PendingJobsRow
	for rows.Next() {
		var i PendingJobsRow
		if err := rows.Scan(&i.T, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const pruneHistory = `-- name: PruneHistory :exec
DELETE
FROM arrower.gue_jobs_history
WHERE created_at <= $1
`

func (q *Queries) PruneHistory(ctx context.Context, createdAt pgtype.Timestamptz) error {
	_, err := q.db.Exec(ctx, pruneHistory, createdAt)
	return err
}

const pruneHistoryPayload = `-- name: PruneHistoryPayload :exec
UPDATE arrower.gue_jobs_history
SET args      = ''::BYTEA,
    pruned_at = NOW()
WHERE queue = $1
  AND created_at <= $2
`

type PruneHistoryPayloadParams struct {
	Queue     string
	CreatedAt pgtype.Timestamptz
}

func (q *Queries) PruneHistoryPayload(ctx context.Context, arg PruneHistoryPayloadParams) error {
	_, err := q.db.Exec(ctx, pruneHistoryPayload, arg.Queue, arg.CreatedAt)
	return err
}

type ScheduleJobsParams struct {
	JobID     string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	Queue     string
	JobType   string
	Priority  int16
	RunAt     pgtype.Timestamptz
	Args      []byte
}

const statsAvgDurationOfJobs = `-- name: StatsAvgDurationOfJobs :one
SELECT COALESCE(AVG(EXTRACT(MICROSECONDS FROM (finished_at - created_at))), 0)::FLOAT AS durration_in_microseconds
FROM arrower.gue_jobs_history
WHERE queue = $1
`

func (q *Queries) StatsAvgDurationOfJobs(ctx context.Context, queue string) (float64, error) {
	row := q.db.QueryRow(ctx, statsAvgDurationOfJobs, queue)
	var durration_in_microseconds float64
	err := row.Scan(&durration_in_microseconds)
	return durration_in_microseconds, err
}

const statsFailedJobs = `-- name: StatsFailedJobs :one
SELECT COUNT(*)
FROM arrower.gue_jobs
WHERE queue = $1
  AND error_count <> 0
`

func (q *Queries) StatsFailedJobs(ctx context.Context, queue string) (int64, error) {
	row := q.db.QueryRow(ctx, statsFailedJobs, queue)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const statsPendingJobs = `-- name: StatsPendingJobs :one
SELECT COUNT(*)
FROM arrower.gue_jobs
WHERE queue = $1
`

func (q *Queries) StatsPendingJobs(ctx context.Context, queue string) (int64, error) {
	row := q.db.QueryRow(ctx, statsPendingJobs, queue)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const statsPendingJobsPerType = `-- name: StatsPendingJobsPerType :many
SELECT job_type, COUNT(*) as count
FROM arrower.gue_jobs
WHERE queue = $1
GROUP BY job_type
`

type StatsPendingJobsPerTypeRow struct {
	JobType string
	Count   int64
}

func (q *Queries) StatsPendingJobsPerType(ctx context.Context, queue string) ([]StatsPendingJobsPerTypeRow, error) {
	rows, err := q.db.Query(ctx, statsPendingJobsPerType, queue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StatsPendingJobsPerTypeRow
	for rows.Next() {
		var i StatsPendingJobsPerTypeRow
		if err := rows.Scan(&i.JobType, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const statsProcessedJobs = `-- name: StatsProcessedJobs :one
SELECT COUNT(DISTINCT job_id)
FROM arrower.gue_jobs_history
WHERE queue = $1
  AND success = true
`

func (q *Queries) StatsProcessedJobs(ctx context.Context, queue string) (int64, error) {
	row := q.db.QueryRow(ctx, statsProcessedJobs, queue)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const statsQueueWorkerPoolSize = `-- name: StatsQueueWorkerPoolSize :one
SELECT COALESCE(SUM(workers), 0)::INTEGER
FROM arrower.gue_jobs_worker_pool
WHERE queue = $1
  AND updated_at > NOW() - INTERVAL '1 minutes'
`

func (q *Queries) StatsQueueWorkerPoolSize(ctx context.Context, queue string) (int32, error) {
	row := q.db.QueryRow(ctx, statsQueueWorkerPoolSize, queue)
	var column_1 int32
	err := row.Scan(&column_1)
	return column_1, err
}

const updateRunAt = `-- name: UpdateRunAt :exec
UPDATE arrower.gue_jobs
SET run_at = $1
WHERE job_id = $2
`

type UpdateRunAtParams struct {
	RunAt pgtype.Timestamptz
	JobID string
}

func (q *Queries) UpdateRunAt(ctx context.Context, arg UpdateRunAtParams) error {
	_, err := q.db.Exec(ctx, updateRunAt, arg.RunAt, arg.JobID)
	return err
}

const upsertWorkerToPool = `-- name: UpsertWorkerToPool :exec
INSERT INTO arrower.gue_jobs_worker_pool (id, queue, workers, created_at, updated_at)
VALUES ($1, $2, $3, STATEMENT_TIMESTAMP(), $4)
ON CONFLICT (id, queue) DO UPDATE SET updated_at = STATEMENT_TIMESTAMP(),
                                      workers    = $3
`

type UpsertWorkerToPoolParams struct {
	ID        string
	Queue     string
	Workers   int16
	UpdatedAt pgtype.Timestamptz
}

func (q *Queries) UpsertWorkerToPool(ctx context.Context, arg UpsertWorkerToPoolParams) error {
	_, err := q.db.Exec(ctx, upsertWorkerToPool,
		arg.ID,
		arg.Queue,
		arg.Workers,
		arg.UpdatedAt,
	)
	return err
}
