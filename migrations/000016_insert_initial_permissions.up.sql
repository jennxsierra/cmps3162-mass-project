-- migrations/000016_insert_initial_permissions.up.sql
-- Seeds baseline permission codes for the medical appointment scheduling API.

INSERT INTO permissions (code)
VALUES
    ('system:read'),
    ('appointments:read'),
    ('appointments:write'),
    ('patients:read'),
    ('patients:write'),
    ('providers:read'),
    ('providers:write'),
    ('staff:read'),
    ('staff:write'),
    ('appointment-types:read'),
    ('appointment-types:write'),
    ('specialties:read'),
    ('specialties:write'),
    ('contact-types:read'),
    ('cancellation-reasons:read'),
    ('person-contacts:read'),
    ('person-contacts:write');
