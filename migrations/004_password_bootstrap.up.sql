ALTER TABLE users
ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT FALSE;

UPDATE users
SET must_change_password = FALSE
WHERE must_change_password IS NULL;

ALTER TABLE users
ALTER COLUMN must_change_password SET NOT NULL;
