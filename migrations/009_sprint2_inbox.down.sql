ALTER TABLE quote_requests
  DROP COLUMN IF EXISTS outcome_amount_clp,
  DROP COLUMN IF EXISTS follow_up_at,
  DROP COLUMN IF EXISTS tags,
  DROP COLUMN IF EXISTS first_response_at;
