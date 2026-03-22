CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE companies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(100) UNIQUE NOT NULL,
    name        VARCHAR(200) NOT NULL,
    tagline     VARCHAR(200),
    description TEXT,
    location    VARCHAR(100),
    region      VARCHAR(100),
    categories  TEXT[] DEFAULT '{}',
    services    TEXT[] DEFAULT '{}',
    phone       VARCHAR(50),
    email       VARCHAR(200),
    website     VARCHAR(300),
    years_active INTEGER,
    featured    BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(200) UNIQUE NOT NULL,
    password_hash VARCHAR(200) NOT NULL,
    company_id    UUID REFERENCES companies(id),
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE quote_requests (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_name    VARCHAR(200) NOT NULL,
    requester_company VARCHAR(200),
    requester_email   VARCHAR(200) NOT NULL,
    requester_phone   VARCHAR(50),
    service           VARCHAR(200) NOT NULL,
    description       TEXT,
    location          VARCHAR(200),
    target_company_id UUID REFERENCES companies(id),
    status            VARCHAR(20) DEFAULT 'new' CHECK (status IN ('new','read','responded')),
    created_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE company_services (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID REFERENCES companies(id) ON DELETE CASCADE,
    name        VARCHAR(200) NOT NULL,
    category    VARCHAR(100),
    description TEXT,
    status      VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active','draft')),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE provider_registrations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name VARCHAR(200) NOT NULL,
    email        VARCHAR(200) NOT NULL,
    phone        VARCHAR(50),
    region       VARCHAR(100),
    services     TEXT[] DEFAULT '{}',
    description  TEXT,
    status       VARCHAR(20) DEFAULT 'pending',
    created_at   TIMESTAMPTZ DEFAULT NOW()
);
