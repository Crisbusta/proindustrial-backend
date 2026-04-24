ALTER TABLE quote_requests
  ADD COLUMN reply_note TEXT,
  ADD COLUMN replied_at TIMESTAMPTZ;
