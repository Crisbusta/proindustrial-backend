DROP INDEX IF EXISTS idx_provider_registrations_status_created;
DROP INDEX IF EXISTS provider_registrations_pending_email_uniq;
DROP INDEX IF EXISTS users_email_lower_uniq;

ALTER TABLE companies              DROP CONSTRAINT IF EXISTS companies_description_len;
ALTER TABLE provider_registrations DROP CONSTRAINT IF EXISTS provider_registrations_description_len;
ALTER TABLE provider_registrations DROP CONSTRAINT IF EXISTS provider_registrations_rejection_len;

ALTER TABLE provider_registrations
  DROP COLUMN IF EXISTS rejected_at,
  DROP COLUMN IF EXISTS rejection_reason,
  DROP COLUMN IF EXISTS email_note,
  DROP COLUMN IF EXISTS email_status,
  DROP COLUMN IF EXISTS approved_at,
  DROP COLUMN IF EXISTS user_id,
  DROP COLUMN IF EXISTS company_id;

-- La normalización de correos a minúsculas no se revierte: es idempotente
-- y revertirla exigiría conocer las mayúsculas originales.
