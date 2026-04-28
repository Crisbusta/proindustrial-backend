ALTER TABLE quote_requests
  ADD COLUMN IF NOT EXISTS first_response_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS tags              TEXT[]    NOT NULL DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS follow_up_at      TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS outcome_amount_clp BIGINT;
