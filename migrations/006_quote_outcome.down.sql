ALTER TABLE quote_requests
  DROP COLUMN IF EXISTS outcome,
  DROP COLUMN IF EXISTS outcome_note,
  DROP COLUMN IF EXISTS closed_at;
