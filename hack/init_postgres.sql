CREATE USER powerbase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;

-- Powerbase super admin
CREATE USER powerbase_auth_admin NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD 'root';
CREATE SCHEMA IF NOT EXISTS auth AUTHORIZATION powerbase_auth_admin;
GRANT CREATE ON DATABASE postgres TO powerbase_auth_admin;
ALTER USER powerbase_auth_admin SET search_path = 'auth';
