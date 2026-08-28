-- ============================================================
-- 011 — Endurecimiento del flujo de registro → aprobación
--
-- IMPORTANTE: RunMigrations hace log.Fatalf ante cualquier error,
-- así que una migración que falle impide arrancar la aplicación.
-- Por eso TODO lo que puede chocar con datos ya existentes va
-- protegido por guardas: ante datos sucios se emite un WARNING y
-- se continúa, en vez de abortar el deploy.
-- ============================================================

-- ────────────────────────────────────────────────────────────
-- (1) Normalización de correos
--     Se bajan a minúsculas SOLO las filas que no colisionan con
--     otra ya existente. Las colisiones se dejan intactas y se
--     reportan, para resolverlas a mano sin bloquear el arranque.
-- ────────────────────────────────────────────────────────────

UPDATE users u
SET email = lower(btrim(u.email))
WHERE u.email <> lower(btrim(u.email))
  AND NOT EXISTS (
    SELECT 1 FROM users x
    WHERE x.id <> u.id AND x.email = lower(btrim(u.email))
  );

UPDATE companies c
SET email = lower(btrim(c.email))
WHERE c.email IS NOT NULL
  AND c.email <> lower(btrim(c.email));

UPDATE provider_registrations r
SET email = lower(btrim(r.email))
WHERE r.email <> lower(btrim(r.email));

DO $$
DECLARE dup_count INT;
BEGIN
  SELECT count(*) INTO dup_count FROM (
    SELECT lower(email) FROM users GROUP BY 1 HAVING count(*) > 1
  ) d;

  IF dup_count = 0 THEN
    CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uniq ON users (lower(email));
  ELSE
    RAISE WARNING '011: % correos de usuario duplicados al normalizar mayúsculas. Índice único omitido; resolver manualmente y crearlo después.', dup_count;
  END IF;
END $$;

-- ────────────────────────────────────────────────────────────
-- (2) Vínculo explícito registro → empresa/usuario
--     Reemplaza la búsqueda por email de DeleteApprovedCompanyByRegistration,
--     que podía borrar la empresa equivocada o no encontrar ninguna.
-- ────────────────────────────────────────────────────────────

ALTER TABLE provider_registrations
  ADD COLUMN IF NOT EXISTS company_id       UUID REFERENCES companies(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS user_id          UUID REFERENCES users(id)     ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS approved_at      TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS email_status     VARCHAR(20),
  ADD COLUMN IF NOT EXISTS email_note       TEXT,
  -- Motivo del rechazo: queda registrado para que la decisión sea auditable
  -- y para poder explicársela al proveedor si vuelve a preguntar.
  ADD COLUMN IF NOT EXISTS rejection_reason TEXT,
  ADD COLUMN IF NOT EXISTS rejected_at      TIMESTAMPTZ;

-- Backfill de registros ya aprobados: solo cuando el correo identifica
-- una única empresa / un único usuario. Los ambiguos quedan en NULL y
-- el código cae al camino de compatibilidad por email.
UPDATE provider_registrations r
SET company_id = c.id
FROM companies c
WHERE r.status = 'approved'
  AND r.company_id IS NULL
  AND lower(c.email) = lower(r.email)
  AND (SELECT count(*) FROM companies x WHERE lower(x.email) = lower(r.email)) = 1;

UPDATE provider_registrations r
SET user_id = u.id
FROM users u
WHERE r.status = 'approved'
  AND r.user_id IS NULL
  AND lower(u.email) = lower(r.email)
  AND (SELECT count(*) FROM users x WHERE lower(x.email) = lower(r.email)) = 1;

UPDATE provider_registrations
SET approved_at = created_at
WHERE status = 'approved' AND approved_at IS NULL;

UPDATE provider_registrations
SET rejected_at = created_at
WHERE status = 'rejected' AND rejected_at IS NULL;

-- ────────────────────────────────────────────────────────────
-- (3) Topes de longitud
--     NOT VALID: las filas históricas que ya exceden el límite
--     (incluida la que originó esta revisión) se conservan; el
--     tope solo aplica a escrituras nuevas. El código recorta al
--     copiar a companies, para que esos registros igual se puedan aprobar.
-- ────────────────────────────────────────────────────────────

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'provider_registrations_description_len'
  ) THEN
    ALTER TABLE provider_registrations
      ADD CONSTRAINT provider_registrations_description_len
      CHECK (description IS NULL OR char_length(description) <= 600) NOT VALID;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'provider_registrations_rejection_len'
  ) THEN
    ALTER TABLE provider_registrations
      ADD CONSTRAINT provider_registrations_rejection_len
      CHECK (rejection_reason IS NULL OR char_length(rejection_reason) <= 500) NOT VALID;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'companies_description_len'
  ) THEN
    ALTER TABLE companies
      ADD CONSTRAINT companies_description_len
      CHECK (description IS NULL OR char_length(description) <= 600) NOT VALID;
  END IF;
END $$;

-- ────────────────────────────────────────────────────────────
-- (4) Una sola solicitud pendiente por correo
-- ────────────────────────────────────────────────────────────

DO $$
DECLARE dup_count INT;
BEGIN
  SELECT count(*) INTO dup_count FROM (
    SELECT lower(email) FROM provider_registrations
    WHERE status = 'pending' GROUP BY 1 HAVING count(*) > 1
  ) d;

  IF dup_count = 0 THEN
    CREATE UNIQUE INDEX IF NOT EXISTS provider_registrations_pending_email_uniq
      ON provider_registrations (lower(email)) WHERE status = 'pending';
  ELSE
    RAISE WARNING '011: % correos con más de una solicitud pendiente. Índice único omitido; depurar la cola y crearlo después.', dup_count;
  END IF;
END $$;

-- ────────────────────────────────────────────────────────────
-- (5) Índice para la bandeja del admin
-- ────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_provider_registrations_status_created
  ON provider_registrations (status, created_at DESC);
