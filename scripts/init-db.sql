-- Initialize databases for Clean DDD ES Template
-- This script creates the necessary databases for write, read, and event storage

-- Create read database
CREATE DATABASE clean_ddd_read_db;

-- Create event database  
CREATE DATABASE clean_ddd_event_db;

-- Grant permissions to postgres user
GRANT ALL PRIVILEGES ON DATABASE clean_ddd_read_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE clean_ddd_event_db TO postgres;

-- Connect to read database and create extensions
\c clean_ddd_read_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect to event database and create extensions
\c clean_ddd_event_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Connect back to write database
\c clean_ddd_write_db; 