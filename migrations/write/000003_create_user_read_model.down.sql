-- Migration: 000003_create_user_read_model
-- Description: Rollback user read model table

DROP TRIGGER IF EXISTS update_user_read_models_updated_at ON user_read_models;
DROP INDEX IF EXISTS idx_user_read_models_deleted_at;
DROP INDEX IF EXISTS idx_user_read_models_created_at;
DROP INDEX IF EXISTS idx_user_read_models_email;
DROP TABLE IF EXISTS user_read_models; 