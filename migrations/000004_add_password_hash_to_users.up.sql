-- Add password_hash column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

-- Add index on password_hash for performance (optional)
CREATE INDEX IF NOT EXISTS idx_users_password_hash ON users(password_hash);

-- Add comment for documentation
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password - never expose in API responses'; 