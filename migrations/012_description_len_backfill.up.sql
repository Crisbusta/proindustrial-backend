-- ============================================================
-- 012 — Descripciones históricas por encima del tope
--
-- La migración 011 agregó los CHECK de longitud como NOT VALID
-- suponiendo que así las filas históricas quedaban exentas. NOT VALID
-- solo evita el escaneo inicial de la tabla: Postgres evalúa igual el
-- CHECK sobre la fila COMPLETA en cada INSERT y en cada UPDATE. Por eso
-- una solicitud con descripción de más de 600 caracteres quedó imposible
-- de resolver: tanto el UPDATE de rechazo como el de aprobación fallaban
-- contra provider_registrations_description_len sin tocar description.
--
-- Aquí se recorta el dato histórico al tope canónico (model.MaxDescription
-- = 600, espejado en frontend-public/src/constants/limits.ts) y se validan
-- los CHECK, para que el invariante valga en TODAS las filas y el problema
-- no pueda reaparecer con otra fila vieja.
--
-- Igual que en 011: ante datos inesperados se emite WARNING y se continúa,
-- porque RunMigrations hace log.Fatalf y una migración que falla impide
-- arrancar la aplicación.
-- ============================================================

-- ────────────────────────────────────────────────────────────
-- (1) Recorte de las filas que exceden el tope
--     left() cuenta caracteres, igual que el recorte por runas
--     de truncatedDescription en admin_repo.go.
-- ────────────────────────────────────────────────────────────

DO $$
DECLARE affected INT;
BEGIN
  UPDATE provider_registrations
  SET description = btrim(left(description, 600))
  WHERE description IS NOT NULL AND char_length(description) > 600;
  GET DIAGNOSTICS affected = ROW_COUNT;
  IF affected > 0 THEN
    RAISE NOTICE '012: % descripciones de solicitudes recortadas a 600 caracteres.', affected;
  END IF;

  UPDATE provider_registrations
  SET rejection_reason = btrim(left(rejection_reason, 500))
  WHERE rejection_reason IS NOT NULL AND char_length(rejection_reason) > 500;
  GET DIAGNOSTICS affected = ROW_COUNT;
  IF affected > 0 THEN
    RAISE NOTICE '012: % motivos de rechazo recortados a 500 caracteres.', affected;
  END IF;

  UPDATE companies
  SET description = btrim(left(description, 600))
  WHERE description IS NOT NULL AND char_length(description) > 600;
  GET DIAGNOSTICS affected = ROW_COUNT;
  IF affected > 0 THEN
    RAISE NOTICE '012: % descripciones de empresas recortadas a 600 caracteres.', affected;
  END IF;
END $$;

-- ────────────────────────────────────────────────────────────
-- (2) Validación de los CHECK
--     Ya sin filas infractoras, VALIDATE deja de marcarlos como
--     "no verificados" y documenta que el invariante vale para todo
--     el histórico. Es idempotente sobre un CHECK ya validado.
-- ────────────────────────────────────────────────────────────

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'provider_registrations_description_len') THEN
    ALTER TABLE provider_registrations VALIDATE CONSTRAINT provider_registrations_description_len;
  END IF;
EXCEPTION WHEN others THEN
  RAISE WARNING '012: no se pudo validar provider_registrations_description_len (%). El recorte igual se aplicó.', SQLERRM;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'provider_registrations_rejection_len') THEN
    ALTER TABLE provider_registrations VALIDATE CONSTRAINT provider_registrations_rejection_len;
  END IF;
EXCEPTION WHEN others THEN
  RAISE WARNING '012: no se pudo validar provider_registrations_rejection_len (%).', SQLERRM;
END $$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'companies_description_len') THEN
    ALTER TABLE companies VALIDATE CONSTRAINT companies_description_len;
  END IF;
EXCEPTION WHEN others THEN
  RAISE WARNING '012: no se pudo validar companies_description_len (%).', SQLERRM;
END $$;
