-- migrations/000018_grant_john_doe_write_permissions.down.sql
-- Removes write permissions from user John Doe.

DELETE FROM users_permissions up
USING users u, permissions p
WHERE up.user_id = u.id
  AND up.permission_id = p.id
  AND lower(u.username) = lower('John Doe')
  AND p.code LIKE '%:write';
