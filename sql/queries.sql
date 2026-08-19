-- name: CreateJob :one
INSERT INTO video_jobs (
    filename
) VALUES (
             $1
         )
    RETURNING *;

-- name: LockAndMarkProcessing :many
UPDATE video_jobs
SET status    = 'PROCESSING',
    locked_at = now(),
    locked_by = $1
WHERE id IN (
    SELECT id
    FROM video_jobs
    WHERE status = 'PENDING'
      AND (next_retry_at IS NULL OR next_retry_at <= now())
    ORDER BY created_at
    LIMIT $2
    FOR UPDATE SKIP LOCKED
            )
            RETURNING *;

-- name: MarkDone :exec
UPDATE video_jobs
SET status    = 'DONE',
    locked_at = NULL,
    locked_by = NULL,
    updated_at = now()
WHERE id = $1;

-- name: MarkRetry :exec
UPDATE video_jobs
SET retry_count   = retry_count + 1,
    next_retry_at = now() + ($1::int * interval '1 second'),
    status        = 'PENDING',
    locked_at     = NULL,
    locked_by     = NULL,
    updated_at    = now()
WHERE id = $2;

-- name: MarkFailed :exec
UPDATE video_jobs
SET status    = 'FAILED',
    locked_at = NULL,
    locked_by = NULL,
    updated_at = now()
WHERE id = $1;

-- name: GetJobByID :one
SELECT *
FROM video_jobs
WHERE id = $1;

-- name: JobExists :one
SELECT EXISTS (
    SELECT 1 FROM video_jobs WHERE id = $1
) AS exists;

-- name: CountByStatus :many
SELECT status, COUNT(*) AS count
FROM video_jobs
GROUP BY status;

-- name: ResetStuckJobs :exec
UPDATE video_jobs
SET status    = 'PENDING',
    locked_at = NULL,
    locked_by = NULL,
    updated_at = now()
WHERE status = 'PROCESSING'
  AND locked_at < now() - interval '10 minutes';

-- name: DeleteOldDoneJobs :exec
DELETE FROM video_jobs
WHERE status = 'DONE'
  AND updated_at < now() - interval '7 days';