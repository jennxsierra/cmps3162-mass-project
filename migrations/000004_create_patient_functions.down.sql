-- migrations/000004_create_patient_functions.down.sql
-- Drops patient CRUD + search functions.

-- ====================================================================================
-- DROP FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_patient(TEXT);
DROP FUNCTION IF EXISTS fn_update_patient(TEXT, TEXT, TEXT, DATE, TEXT, TEXT, TEXT);
DROP FUNCTION IF EXISTS fn_get_patients(TEXT, TEXT, INT, INT);
DROP FUNCTION IF EXISTS fn_get_patient(TEXT);
DROP FUNCTION IF EXISTS fn_create_patient(TEXT, TEXT, DATE, TEXT, TEXT, TEXT);
