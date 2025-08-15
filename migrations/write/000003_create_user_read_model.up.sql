-- Migration: 000003_create_user_read_model
-- Description: Create user read model table for CQRS

CREATE TABLE IF NOT EXISTS user_read_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create indexes for read model
CREATE INDEX IF NOT EXISTS idx_user_read_models_email ON user_read_models(email);
CREATE INDEX IF NOT EXISTS idx_user_read_models_created_at ON user_read_models(created_at);
CREATE INDEX IF NOT EXISTS idx_user_read_models_deleted_at ON user_read_models(deleted_at);

-- Create trigger to update updated_at
CREATE TRIGGER update_user_read_models_updated_at 
    BEFORE UPDATE ON user_read_models 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 