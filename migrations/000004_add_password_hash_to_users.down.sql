-- Remove password_hash column from users table
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

-- Remove index if it exists
DROP INDEX IF EXISTS idx_users_password_hash; 