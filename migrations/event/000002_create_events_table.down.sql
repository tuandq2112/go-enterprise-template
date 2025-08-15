-- Migration: 000002_create_events_table
-- Description: Rollback events table

DROP INDEX IF EXISTS idx_events_aggregate_version_unique;
DROP INDEX IF EXISTS idx_events_aggregate_version;
DROP INDEX IF EXISTS idx_events_created_at;
DROP INDEX IF EXISTS idx_events_event_type;
DROP INDEX IF EXISTS idx_events_aggregate_type;
DROP INDEX IF EXISTS idx_events_aggregate_id;
DROP TABLE IF EXISTS events; 