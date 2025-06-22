-- Initial database setup for Nginx Manager development
-- This script runs when the PostgreSQL container is first created

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create additional schemas if needed
-- CREATE SCHEMA IF NOT EXISTS analytics;
-- CREATE SCHEMA IF NOT EXISTS monitoring;

-- Set timezone
SET timezone = 'UTC';

-- Create a test user for development (optional)
-- This is just for development - production should use proper user management
DO $$
BEGIN
    -- You can add any initial data setup here
    -- For example, creating a default admin user
    -- INSERT INTO users (email, password_hash, role) VALUES 
    --   ('admin@nginx-manager.local', crypt('admin123', gen_salt('bf')), 'admin');
    
    RAISE NOTICE 'Database initialization completed successfully';
END $$;

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE nginx_manager TO nginx_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO nginx_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO nginx_user;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO nginx_user;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO nginx_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO nginx_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO nginx_user;
