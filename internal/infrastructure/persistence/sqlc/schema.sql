-- =========================
-- Table principale
-- =========================
CREATE TABLE IF NOT EXISTS video_jobs (
                                        id BIGSERIAL PRIMARY KEY,

  -- Nom du fichier (ou clé S3)
                                        filename TEXT NOT NULL,

  -- Statut du job
                                        status TEXT NOT NULL DEFAULT 'PENDING',

  -- Nombre de retries effectués
                                        retry_count INT NOT NULL DEFAULT 0,

  -- Date du prochain retry
                                        next_retry_at TIMESTAMP NULL,

  -- Dates de suivi
                                        created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now()
  );

-- =========================
-- Contraintes
-- =========================

-- Statuts autorisés
ALTER TABLE video_jobs
  ADD CONSTRAINT video_jobs_status_check
    CHECK (status IN ('PENDING', 'PROCESSING', 'DONE', 'FAILED'));

-- =========================
-- Index (performance)
-- =========================

-- Sélection rapide des jobs à traiter
CREATE INDEX IF NOT EXISTS idx_video_jobs_pending
  ON video_jobs (status, retry_count, created_at);

-- Suivi des jobs en cours (cleanup)
CREATE INDEX IF NOT EXISTS idx_video_jobs_processing
  ON video_jobs (status, updated_at);

-- Retry scheduling
CREATE INDEX IF NOT EXISTS idx_video_jobs_retry
  ON video_jobs (status, next_retry_at);

-- =========================
-- Trigger updated_at auto
-- =========================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_updated_at
  BEFORE UPDATE ON video_jobs
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();
