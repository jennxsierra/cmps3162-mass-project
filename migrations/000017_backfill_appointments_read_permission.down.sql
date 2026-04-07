-- migrations/000017_backfill_appointments_read_permission.down.sql
-- Removes backfilled appointments:read permission assignments.

DELETE FROM users_permissions up
USING permissions p
WHERE up.permission_id = p.id
  AND p.code = 'appointments:read';
