-- migrations/000016_insert_initial_permissions.down.sql
-- Removes seeded baseline permission codes.

DELETE FROM permissions
WHERE code IN (
    'appointments:read',
    'appointments:write',
    'patients:read',
    'patients:write',
    'providers:read',
    'providers:write'
);
