-- migrations/000014_create_permissions_table.up.sql
-- Adds permissions table for fine-grained authorization.

CREATE TABLE IF NOT EXISTS permissions (
    id bigserial PRIMARY KEY,
    code text NOT NULL
);
