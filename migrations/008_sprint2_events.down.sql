DROP INDEX IF EXISTS idx_companies_tsv;
ALTER TABLE companies DROP COLUMN IF EXISTS tsv;
DROP INDEX IF EXISTS idx_events_type;
DROP INDEX IF EXISTS idx_events_company_time;
DROP TABLE IF EXISTS company_events;
