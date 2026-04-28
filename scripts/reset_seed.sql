-- ============================================================
-- RESET + RE-SEED de base de datos PuntoFusión
-- Contraseña de todos los usuarios: demo123
-- Admin: admin@puntofusion.local / demo123
--
-- Correr en Railway → PostgreSQL → Data (editor SQL)
-- ============================================================

-- 1. Limpiar datos (CASCADE respeta foreign keys)
TRUNCATE company_events       CASCADE;
TRUNCATE company_project_images CASCADE;
TRUNCATE company_projects     CASCADE;
TRUNCATE company_certifications CASCADE;
TRUNCATE service_images       CASCADE;
TRUNCATE quote_requests       CASCADE;
TRUNCATE company_services     CASCADE;
TRUNCATE company_service_regions CASCADE;
TRUNCATE companies            CASCADE;
TRUNCATE users                CASCADE;

-- 2. Empresas seed
INSERT INTO companies (slug, name, tagline, description, location, region, categories, services, phone, email, years_active, featured) VALUES
(
  'proveedora-aceros-pacifico',
  'Proveedora Aceros del Pacífico',
  'Especialistas en termofusión y tuberías PEAD desde 1998',
  'Empresa con más de 25 años de experiencia en instalación y mantención de sistemas de tuberías PEAD mediante termofusión. Trabajamos con proyectos de agua potable, riego tecnificado y minería en todo el norte de Chile.',
  'Antofagasta', 'Antofagasta',
  ARRAY['termofusion', 'tuberias-industriales'],
  ARRAY['Termofusión de tuberías PEAD','Instalación de tuberías a presión','Reparación de redes de agua potable','Proyectos mineros de transporte de pulpa'],
  '+56 55 234 5678', 'contacto@acerospacifico.cl', 25, true
),
(
  'tuberias-del-sur',
  'Tuberías del Sur S.A.',
  'Geomembranas e impermeabilización para proyectos de gran escala',
  'Líderes en instalación de geomembranas HDPE en la zona sur de Chile. Ejecutamos impermeabilización de tranques de relave, piscinas de proceso, rellenos sanitarios y proyectos de acuicultura.',
  'Puerto Montt', 'Los Lagos',
  ARRAY['geomembranas', 'obras-civiles'],
  ARRAY['Instalación de geomembranas HDPE','Impermeabilización de tranques','Liners para piscinas industriales','Rellenos sanitarios','Proyectos de acuicultura'],
  '+56 65 221 8900', 'proyectos@tuberiasdelsur.cl', 18, true
),
(
  'electro-industrial-spa',
  'Electro Industrial SpA',
  'Montaje electromecánico e industrial para minería y energía',
  'Empresa especializada en montaje industrial, instalaciones electromecánicas y servicios de mantención para la industria minera y energética. Operamos en las regiones de Atacama y Coquimbo.',
  'Copiapó', 'Atacama',
  ARRAY['montaje-industrial', 'servicios-hidraulicos'],
  ARRAY['Montaje electromecánico','Instalación de bombas y válvulas','Sistemas hidráulicos industriales','Mantención de plantas','Servicios de turnaround'],
  '+56 52 261 3344', 'info@electroindustrial.cl', 12, true
),
(
  'hormigonsur',
  'HormigonSur Ltda.',
  'Obras civiles y fundaciones para proyectos industriales del sur',
  'Contratista de obras civiles con foco en fundaciones, pavimentos industriales y hormigón para proyectos de infraestructura. Trabajamos con constructoras, salmoneras y mineras en la zona sur.',
  'Temuco', 'La Araucanía',
  ARRAY['obras-civiles', 'montaje-industrial'],
  ARRAY['Fundaciones industriales','Pavimentos de hormigón','Movimiento de tierras','Estructuras de hormigón'],
  '+56 45 244 7711', 'obras@hormigonsur.cl', 9, false
),
(
  'hidro-norte',
  'Hidro Norte Ingeniería',
  'Soluciones hidráulicas y de tuberías para el norte minero',
  'Empresa de ingeniería especializada en sistemas hidráulicos para la minería del norte grande. Diseño, suministro e instalación de sistemas de bombeo, tuberías y control de fluidos.',
  'Calama', 'Antofagasta',
  ARRAY['servicios-hidraulicos', 'tuberias-industriales'],
  ARRAY['Diseño de sistemas de bombeo','Instalación de cañerías industriales','Control de válvulas y actuadores','Estudios de golpe de ariete'],
  '+56 55 278 9900', 'contacto@hidronorte.cl', 14, false
),
(
  'geomembranas-atacama',
  'Geomembranas Atacama',
  'Instalación de geomembranas y soluciones de contención minera',
  'Especialistas en instalación de geomembranas para la industria minera en la Región de Atacama y Antofagasta. Proyectos de impermeabilización de pilas de lixiviación, tranques y botaderos.',
  'Antofagasta', 'Antofagasta',
  ARRAY['geomembranas'],
  ARRAY['Geomembranas HDPE para pilas de lixiviación','Impermeabilización de tranques de relave','Liners para botaderos','Reparación y mantención de geomembranas'],
  '+56 55 291 4433', 'ventas@geomembranasatacama.cl', 16, false
),
(
  'fusiones-pacifico',
  'Fusiones Pacífico',
  'Termofusión certificada para agua potable y riego',
  'Empresa certificada en termofusión de tuberías PEAD y PP-R para proyectos de agua potable rural, riego tecnificado y redes urbanas. Operamos en la Región del Maule y O''Higgins.',
  'Talca', 'Maule',
  ARRAY['termofusion'],
  ARRAY['Termofusión PEAD a tope y electrofusión','Redes de agua potable rural','Sistemas de riego tecnificado','Instalación en zanjas y microtúnel'],
  '+56 71 224 6622', 'info@fusionespacifico.cl', 11, false
),
(
  'montajes-valparaiso',
  'Montajes Valparaíso S.A.',
  'Montaje industrial y mantenimiento en la zona central',
  'Empresa de montaje industrial con base en Valparaíso. Especialistas en instalaciones mecánicas, estructuras metálicas y servicios de mantención para la industria portuaria, química y energética.',
  'Valparaíso', 'Valparaíso',
  ARRAY['montaje-industrial', 'tuberias-industriales'],
  ARRAY['Montaje de estructuras metálicas','Instalación de cañerías industriales','Mantención de plantas industriales','Servicios portuarios'],
  '+56 32 255 1199', 'licitaciones@montajesvalparaiso.cl', 22, true
);

-- 3. Usuarios proveedor (password = "demo123")
INSERT INTO users (email, password_hash, company_id, role, must_change_password) VALUES
  ('contacto@acerospacifico.cl',         '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', (SELECT id FROM companies WHERE slug='proveedora-aceros-pacifico'), 'provider', false),
  ('proyectos@tuberiasdelsur.cl',        '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', (SELECT id FROM companies WHERE slug='tuberias-del-sur'),           'provider', false),
  ('info@electroindustrial.cl',          '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', (SELECT id FROM companies WHERE slug='electro-industrial-spa'),    'provider', false),
  ('obras@hormigonsur.cl',               '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', (SELECT id FROM companies WHERE slug='hormigonsur'),               'provider', false),
  ('contacto@hidronorte.cl',             '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', (SELECT id FROM companies WHERE slug='hidro-norte'),               'provider', false),
  ('ventas@geomembranasatacama.cl',      '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', (SELECT id FROM companies WHERE slug='geomembranas-atacama'),      'provider', false),
  ('info@fusionespacifico.cl',           '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', (SELECT id FROM companies WHERE slug='fusiones-pacifico'),          'provider', false),
  ('licitaciones@montajesvalparaiso.cl', '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', (SELECT id FROM companies WHERE slug='montajes-valparaiso'),        'provider', false);

-- 4. Usuario admin (password = "demo123")
INSERT INTO users (email, password_hash, company_id, role, must_change_password) VALUES
  ('admin@puntofusion.local', '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy', NULL, 'admin', false);

-- 5. Servicios de empresas destacadas
INSERT INTO company_services (company_id, name, category, description, status) VALUES
(
  (SELECT id FROM companies WHERE slug='proveedora-aceros-pacifico'),
  'Termofusión a tope PEAD', 'termofusion',
  'Unión de tuberías PEAD por fusión a tope con máquinas CNC. Diámetros DN63 a DN630.',
  'active'
),
(
  (SELECT id FROM companies WHERE slug='proveedora-aceros-pacifico'),
  'Electrofusión en terreno', 'termofusion',
  'Unión por electrofusión en zanjas y espacios reducidos. Certificado conforme a NCh.',
  'active'
),
(
  (SELECT id FROM companies WHERE slug='tuberias-del-sur'),
  'Instalación liner HDPE', 'geomembranas',
  'Instalación de geomembranas HDPE 1.0mm, 1.5mm y 2.0mm con ensayos de hermeticidad.',
  'active'
),
(
  (SELECT id FROM companies WHERE slug='tuberias-del-sur'),
  'Reparación geomembranas', 'geomembranas',
  'Detección de fallas y reparación de liners existentes. Uso de equipo de chispa y cuña.',
  'active'
),
(
  (SELECT id FROM companies WHERE slug='montajes-valparaiso'),
  'Montaje estructuras metálicas', 'montaje-industrial',
  'Montaje de estructuras de acero para plantas industriales y portuarias.',
  'active'
);

-- 6. Solicitudes de cotización de ejemplo
INSERT INTO quote_requests (requester_name, requester_company, requester_email, requester_phone, service, description, location, target_company_id, status) VALUES
(
  'Juan Pérez', 'Minera Collahuasi', 'jperez@collahuasi.cl', '+56 9 8765 4321',
  'Termofusión de tuberías PEAD',
  'Necesitamos instalar 2km de tubería PEAD DN200 para transporte de agua de proceso en faena.',
  'Iquique',
  (SELECT id FROM companies WHERE slug='proveedora-aceros-pacifico'),
  'new'
),
(
  'María González', 'Constructora del Norte', 'mgonzalez@cnorte.cl', '+56 9 7654 3210',
  'Instalación de tuberías a presión',
  'Proyecto de agua potable rural en sector cordillerano. Aproximadamente 800m de tubería.',
  'Antofagasta',
  (SELECT id FROM companies WHERE slug='proveedora-aceros-pacifico'),
  'read'
),
(
  'Carlos Rojas', 'Salmonera Austral', 'crojas@salmonera.cl', '+56 9 6543 2109',
  'Instalación de geomembranas HDPE',
  'Impermeabilización de 3 piscinas de cultivo de 50x30m cada una.',
  'Puerto Montt',
  (SELECT id FROM companies WHERE slug='tuberias-del-sur'),
  'new'
),
(
  'Ana Muñoz', 'Aguas del Sur', 'amunoz@aguassur.cl', '+56 9 5432 1098',
  'Impermeabilización de tranques',
  'Tranque de regulación de 2 hectáreas, requiere liner HDPE 1.5mm.',
  'Temuco',
  (SELECT id FROM companies WHERE slug='tuberias-del-sur'),
  'responded'
),
(
  'Pedro Silva', 'Constructora Vial', 'psilva@cvial.cl', '+56 9 4321 0987',
  'Montaje electromecánico',
  'Montaje de planta procesadora. Incluye tuberías de acero inox y sistema de bombeo.',
  'Copiapó',
  (SELECT id FROM companies WHERE slug='electro-industrial-spa'),
  'new'
);

-- 7. Verificar resultado
SELECT 'companies' AS tabla, COUNT(*) AS registros FROM companies
UNION ALL
SELECT 'users',         COUNT(*) FROM users
UNION ALL
SELECT 'services',      COUNT(*) FROM company_services
UNION ALL
SELECT 'quotes',        COUNT(*) FROM quote_requests;
