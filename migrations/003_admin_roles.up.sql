ALTER TABLE users
ADD COLUMN IF NOT EXISTS role VARCHAR(20);

UPDATE users
SET role = 'provider'
WHERE role IS NULL;

ALTER TABLE users
ALTER COLUMN role SET DEFAULT 'provider';

ALTER TABLE users
ALTER COLUMN role SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'users_role_check'
  ) THEN
    ALTER TABLE users
    ADD CONSTRAINT users_role_check CHECK (role IN ('provider', 'admin'));
  END IF;
END $$;

INSERT INTO users (email, password_hash, company_id, role)
VALUES (
  'admin@puntofusion.local',
  '$2a$10$r1zFxutzHJAHUYG.5UCOdeh4AyJru1vWjfhG3sklvd9Ml0JTILRWy',
  NULL,
  'admin'
)
ON CONFLICT (email) DO UPDATE SET role = EXCLUDED.role;
