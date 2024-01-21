-- name: PendingJobs :many
SELECT bins.*, COUNT(t)
FROM (SELECT date_bin($1, finished_at, TIMESTAMP WITH TIME ZONE'2001-01-01')::TIMESTAMPTZ as t
      FROM arrower.gue_jobs_history
      WHERE finished_at > $2) bins
GROUP BY bins.t
ORDER BY bins.t DESC
LIMIT $3;

-- name: JobTypes :many
SELECT DISTINCT(job_type)
FROM arrower.gue_jobs_history
WHERE queue = $1;

-- name: ScheduleJobs :copyfrom
INSERT INTO arrower.gue_jobs (job_id, created_at, updated_at, queue, job_type, priority, run_at, args)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: JobTableSize :one
SELECT pg_size_pretty(pg_total_relation_size('arrower.gue_jobs'))         as jobs,
       pg_size_pretty(pg_total_relation_size('arrower.gue_jobs_history')) as history;