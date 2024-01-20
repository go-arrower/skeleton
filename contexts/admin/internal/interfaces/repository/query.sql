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