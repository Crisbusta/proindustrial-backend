-- ============================================================
-- BORRADO QUIRÚRGICO de empresas SEED
-- Borra SOLO las 8 empresas de demo (por slug) + sus datos
-- dependientes. NO toca ninguna empresa productiva.
--
-- Correr en Railway → PostgreSQL → Data (editor SQL)
-- Se ejecuta dentro de una transacción: si algo falla, ROLLBACK.
-- ============================================================

BEGIN;

-- Slugs de las empresas seed (única fuente de verdad del filtro)
CREATE TEMP TABLE seed_slugs (slug TEXT) ON COMMIT DROP;
INSERT INTO seed_slugs (slug) VALUES
  ('proveedora-aceros-pacifico'),
  ('tuberias-del-sur'),
  ('electro-industrial-spa'),
  ('hormigonsur'),
  ('hidro-norte'),
  ('geomembranas-atacama'),
  ('fusiones-pacifico'),
  ('montajes-valparaiso');

-- (0) Ver qué se va a borrar ANTES de borrar
SELECT id, slug, name FROM companies
WHERE slug IN (SELECT slug FROM seed_slugs)
ORDER BY slug;

-- (1) quote_requests apuntando a empresas seed (FK sin cascade)
DELETE FROM quote_requests
WHERE target_company_id IN (
  SELECT id FROM companies WHERE slug IN (SELECT slug FROM seed_slugs)
);

-- (2) usuarios de las empresas seed (FK sin cascade)
DELETE FROM users
WHERE company_id IN (
  SELECT id FROM companies WHERE slug IN (SELECT slug FROM seed_slugs)
);

-- (3) las empresas seed
--     Cascade limpia automáticamente: company_services, service_images,
--     company_events, company_media, company_certifications,
--     company_projects, company_project_images, company_service_regions
DELETE FROM companies
WHERE slug IN (SELECT slug FROM seed_slugs);

-- (4) Verificar: debe devolver 0 filas
SELECT id, slug, name FROM companies
WHERE slug IN (SELECT slug FROM seed_slugs);

-- Revisa los resultados de arriba. Si todo se ve bien:
COMMIT;
-- Si algo salió mal, en vez de COMMIT ejecuta:  ROLLBACK;
