ALTER TABLE quote_requests
  ADD COLUMN outcome      VARCHAR(50),
  ADD COLUMN outcome_note TEXT,
  ADD COLUMN closed_at    TIMESTAMPTZ;
