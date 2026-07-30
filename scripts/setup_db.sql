-- Chạy với superuser (postgres hoặc user macOS):
--   make setup-db
-- Hoặc: psql -U postgres -f scripts/setup_db.sql

DO $$ BEGIN
  CREATE USER identity WITH PASSWORD 'identity';
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

SELECT 'CREATE DATABASE identity_core OWNER identity'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'identity_core')\gexec

GRANT ALL PRIVILEGES ON DATABASE identity_core TO identity;
