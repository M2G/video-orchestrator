DO $$ BEGIN
CREATE TYPE video_job_status AS ENUM (
'PENDING',
'PROCESSING',
'DONE',
'FAILED'
);
EXCEPTION
WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS video_jobs (
id BIGSERIAL PRIMARY KEY,
filename TEXT NOT NULL,
status video_job_status NOT NULL DEFAULT 'PENDING',
retry_count INT NOT NULL DEFAULT 0,
next_retry_at TIMESTAMPTZ NULL,
locked_at TIMESTAMPTZ NULL,
locked_by TEXT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Idx opti.
-- Jobs
CREATE INDEX IF NOT EXISTS idx_jobs_ready
ON video_jobs (status, next_retry_at)
WHERE status = 'PENDING';

-- Jobs en cours (monitoring + recovery)
CREATE INDEX IF NOT EXISTS idx_jobs_processing
ON video_jobs (locked_at)
WHERE status = 'PROCESSING';

-- Retry (perf)
CREATE INDEX IF NOT EXISTS idx_jobs_retry
ON video_jobs (next_retry_at)
WHERE status = 'PENDING';

-- Trigger updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_updated_at ON video_jobs;

CREATE TRIGGER trigger_update_updated_at
BEFORE UPDATE ON video_jobs
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
