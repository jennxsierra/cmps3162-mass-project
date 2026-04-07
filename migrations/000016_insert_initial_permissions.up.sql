-- migrations/000016_insert_initial_permissions.up.sql
-- Seeds baseline permission codes for the medical appointment scheduling API.

INSERT INTO permissions (code)
VALUES
    ('appointments:read'),
    ('appointments:write'),
    ('patients:read'),
    ('patients:write'),
    ('providers:read'),
    ('providers:write');
