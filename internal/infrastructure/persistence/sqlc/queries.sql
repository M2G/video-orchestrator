-- name: LockAndMarkProcessing :many
WITH jobs AS (
  SELECT id
  FROM video_jobs
  WHERE status = 'PENDING'
  ORDER BY created_at
  LIMIT $1
  FOR UPDATE SKIP LOCKED
        )
UPDATE video_jobs
SET status = 'PROCESSING',
    updated_at = now()
WHERE id IN (SELECT id FROM jobs)
  RETURNING id, filename, retry_count;

-- name: MarkDone :exec
UPDATE video_jobs
SET status = 'DONE',
    updated_at = now()
WHERE id = $1;

-- name: MarkRetry :exec
UPDATE video_jobs
SET retry_count = retry_count + 1,
    next_retry_at = now() + ($1::int * interval '1 second'),
    status = 'PENDING',
    updated_at = now()
WHERE id = $2;

-- name: MarkFailed :exec
UPDATE video_jobs
SET status = 'FAILED',
    updated_at = now()
WHERE id = $1;

-- name: GetJobByID :one
SELECT id FROM video_jobs WHERE id = $1;

