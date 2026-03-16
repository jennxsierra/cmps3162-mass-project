-- migrations/000006_create_staff_functions.down.sql
-- Drops staff CRUD + search functions.

-- ====================================================================================
-- DROP FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_staff(TEXT);
DROP FUNCTION IF EXISTS fn_update_staff(TEXT, TEXT, TEXT, DATE, TEXT, TEXT);
DROP FUNCTION IF EXISTS fn_get_staffs(TEXT, INT, INT);
DROP FUNCTION IF EXISTS fn_get_staff(TEXT);
DROP FUNCTION IF EXISTS fn_create_staff(TEXT, TEXT, DATE, TEXT, TEXT);
