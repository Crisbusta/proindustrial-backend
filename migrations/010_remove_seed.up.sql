-- ============================================================
-- Elimina las empresas SEED de demo (por slug) + sus datos
-- dependientes. NO toca ninguna empresa productiva.
--
-- Se aplica automáticamente en el próximo deploy vía golang-migrate.
-- El orden respeta las FKs que NO tienen ON DELETE CASCADE
-- (quote_requests.target_company_id y users.company_id).
-- El resto de tablas hijas cascadean solas.
-- ============================================================

-- (1) quote_requests apuntando a empresas seed (FK sin cascade)
DELETE FROM quote_requests
WHERE target_company_id IN (
  SELECT id FROM companies WHERE slug IN (
    'proveedora-aceros-pacifico',
    'tuberias-del-sur',
    'electro-industrial-spa',
    'hormigonsur',
    'hidro-norte',
    'geomembranas-atacama',
    'fusiones-pacifico',
    'montajes-valparaiso'
  )
);

-- (2) usuarios de las empresas seed (FK sin cascade)
DELETE FROM users
WHERE company_id IN (
  SELECT id FROM companies WHERE slug IN (
    'proveedora-aceros-pacifico',
    'tuberias-del-sur',
    'electro-industrial-spa',
    'hormigonsur',
    'hidro-norte',
    'geomembranas-atacama',
    'fusiones-pacifico',
    'montajes-valparaiso'
  )
);

-- (3) las empresas seed (cascade limpia services, events, media,
--     certifications, projects, project_images, service_regions,
--     service_images)
DELETE FROM companies
WHERE slug IN (
  'proveedora-aceros-pacifico',
  'tuberias-del-sur',
  'electro-industrial-spa',
  'hormigonsur',
  'hidro-norte',
  'geomembranas-atacama',
  'fusiones-pacifico',
  'montajes-valparaiso'
);
