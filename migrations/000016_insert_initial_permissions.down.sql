-- migrations/000016_insert_initial_permissions.down.sql
-- Removes seeded baseline permission codes.

DELETE FROM permissions
WHERE code IN (
    'system:read',
    'appointments:read',
    'appointments:write',
    'patients:read',
    'patients:write',
    'providers:read',
    'providers:write',
    'staff:read',
    'staff:write',
    'appointment-types:read',
    'appointment-types:write',
    'specialties:read',
    'specialties:write',
    'contact-types:read',
    'cancellation-reasons:read',
    'person-contacts:read',
    'person-contacts:write'
);
