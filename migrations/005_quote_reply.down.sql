ALTER TABLE quote_requests
  DROP COLUMN IF EXISTS reply_note,
  DROP COLUMN IF EXISTS replied_at;
