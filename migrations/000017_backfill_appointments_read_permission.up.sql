-- migrations/000017_backfill_appointments_read_permission.up.sql
-- Backfills appointments:read permission for existing users.

INSERT INTO users_permissions (user_id, permission_id)
SELECT u.id, p.id
FROM users u
CROSS JOIN permissions p
WHERE p.code = 'appointments:read'
ON CONFLICT (user_id, permission_id) DO NOTHING;
