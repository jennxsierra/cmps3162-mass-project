-- migrations/000013_create_tokens_table.up.sql
-- Adds token storage for activation and authentication workflows.

CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) WITH TIME ZONE NOT NULL,
    scope text NOT NULL
);
