-- Reverse RLS on system tables added in 030_rls_system_tables.up.sql

ALTER TABLE schema_migrations DISABLE ROW LEVEL SECURITY;
ALTER TABLE agent_config DISABLE ROW LEVEL SECURITY;
