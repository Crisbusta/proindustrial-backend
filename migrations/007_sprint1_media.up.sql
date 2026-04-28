-- ── Company media fields ─────────────────────────────────────────────────────
ALTER TABLE companies
  ADD COLUMN IF NOT EXISTS logo_url  TEXT,
  ADD COLUMN IF NOT EXISTS cover_url TEXT;

-- ── Service images ────────────────────────────────────────────────────────────
CREATE TABLE service_images (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service_id UUID NOT NULL REFERENCES company_services(id) ON DELETE CASCADE,
  url        TEXT NOT NULL,
  alt_text   TEXT,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_service_images_service ON service_images(service_id);

-- ── Certifications ────────────────────────────────────────────────────────────
CREATE TABLE company_certifications (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name         VARCHAR(200) NOT NULL,
  issuer       VARCHAR(200),
  document_url TEXT,
  issued_at    DATE,
  expires_at   DATE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_certifications_company ON company_certifications(company_id);

-- ── Projects / casos de éxito ─────────────────────────────────────────────────
CREATE TABLE company_projects (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  title       VARCHAR(300) NOT NULL,
  description TEXT,
  client_name VARCHAR(200),
  year        INTEGER,
  cover_url   TEXT,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_projects_company ON company_projects(company_id);

CREATE TABLE company_project_images (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES company_projects(id) ON DELETE CASCADE,
  url        TEXT NOT NULL,
  alt_text   TEXT,
  sort_order INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_project_images_project ON company_project_images(project_id);

-- ── Service regions (multi-región) ────────────────────────────────────────────
CREATE TABLE company_service_regions (
  company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  region     VARCHAR(100) NOT NULL,
  PRIMARY KEY (company_id, region)
);
CREATE INDEX idx_service_regions_company ON company_service_regions(company_id);

-- ── Indexes for existing tables (missing from Sprint 0) ─────────────────────
CREATE INDEX IF NOT EXISTS idx_quote_requests_company ON quote_requests(target_company_id);
CREATE INDEX IF NOT EXISTS idx_quote_requests_status  ON quote_requests(status);
CREATE INDEX IF NOT EXISTS idx_companies_region       ON companies(region);
CREATE INDEX IF NOT EXISTS idx_companies_featured     ON companies(featured);
