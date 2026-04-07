-- migrations/000018_grant_john_doe_write_permissions.up.sql
-- Grants all write permissions to user John Doe.

INSERT INTO users_permissions (user_id, permission_id)
SELECT u.id, p.id
FROM users u
JOIN permissions p ON p.code LIKE '%:write'
WHERE lower(u.username) = lower('John Doe')
ON CONFLICT (user_id, permission_id) DO NOTHING;
