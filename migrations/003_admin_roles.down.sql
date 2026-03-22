DELETE FROM users WHERE email = 'admin@proindustrial.local';

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_role_check;

ALTER TABLE users
DROP COLUMN IF EXISTS role;
