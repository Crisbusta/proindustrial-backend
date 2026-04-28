DROP INDEX IF EXISTS idx_companies_featured;
DROP INDEX IF EXISTS idx_companies_region;
DROP INDEX IF EXISTS idx_quote_requests_status;
DROP INDEX IF EXISTS idx_quote_requests_company;

DROP TABLE IF EXISTS company_service_regions;
DROP TABLE IF EXISTS company_project_images;
DROP TABLE IF EXISTS company_projects;
DROP TABLE IF EXISTS company_certifications;
DROP TABLE IF EXISTS service_images;

ALTER TABLE companies
  DROP COLUMN IF EXISTS logo_url,
  DROP COLUMN IF EXISTS cover_url;
