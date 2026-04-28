-- Event tracking
CREATE TABLE company_events (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  event_type  TEXT NOT NULL,
  visitor_id  TEXT,
  referrer    TEXT,
  ip_hash     TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_events_company_time ON company_events(company_id, created_at DESC);
CREATE INDEX idx_events_type        ON company_events(event_type);

-- Full-text search on companies
ALTER TABLE companies ADD COLUMN tsv tsvector
  GENERATED ALWAYS AS (
    setweight(to_tsvector('spanish', coalesce(name,'')), 'A') ||
    setweight(to_tsvector('spanish', coalesce(tagline,'')), 'B') ||
    setweight(to_tsvector('spanish', coalesce(description,'')), 'C')
  ) STORED;
CREATE INDEX idx_companies_tsv ON companies USING GIN(tsv);
