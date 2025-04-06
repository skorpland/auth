#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE USER powerbase_admin LOGIN CREATEROLE CREATEDB REPLICATION BYPASSRLS;

    -- Powerbase super admin
    CREATE USER powerbase_auth_admin NOINHERIT CREATEROLE LOGIN NOREPLICATION PASSWORD 'root';
    CREATE SCHEMA IF NOT EXISTS $DB_NAMESPACE AUTHORIZATION powerbase_auth_admin;
    GRANT CREATE ON DATABASE postgres TO powerbase_auth_admin;
    ALTER USER powerbase_auth_admin SET search_path = '$DB_NAMESPACE';
EOSQL
