-- migrations/000015_create_users_permissions_table.down.sql
-- Removes users_permissions mapping table.

DROP TABLE IF EXISTS users_permissions;
