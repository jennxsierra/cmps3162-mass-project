-- migrations/000009_create_reference_functions.down.sql
-- Drops CRUD + list/search functions for reference tables.

-- ====================================================================================
-- DROP CONTACT_TYPE FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_contact_type(INT);
DROP FUNCTION IF EXISTS fn_update_contact_type(INT, TEXT);
DROP FUNCTION IF EXISTS fn_get_contact_types(TEXT, INT, INT);
DROP FUNCTION IF EXISTS fn_get_contact_type(INT);
DROP FUNCTION IF EXISTS fn_create_contact_type(TEXT);

-- ====================================================================================
-- DROP CANCELLATION_REASON FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_cancellation_reason(INT);
DROP FUNCTION IF EXISTS fn_update_cancellation_reason(INT, TEXT);
DROP FUNCTION IF EXISTS fn_get_cancellation_reasons(TEXT, INT, INT);
DROP FUNCTION IF EXISTS fn_get_cancellation_reason(INT);
DROP FUNCTION IF EXISTS fn_create_cancellation_reason(TEXT);

-- ====================================================================================
-- DROP APPT_TYPE FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_appt_type(INT);
DROP FUNCTION IF EXISTS fn_update_appt_type(INT, TEXT);
DROP FUNCTION IF EXISTS fn_get_appt_types(TEXT, INT, INT);
DROP FUNCTION IF EXISTS fn_get_appt_type(INT);
DROP FUNCTION IF EXISTS fn_create_appt_type(TEXT);

-- ====================================================================================
-- DROP SPECIALTY FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_specialty(INT);
DROP FUNCTION IF EXISTS fn_update_specialty(INT, TEXT);
DROP FUNCTION IF EXISTS fn_get_specialties(TEXT, INT, INT);
DROP FUNCTION IF EXISTS fn_get_specialty(INT);
DROP FUNCTION IF EXISTS fn_create_specialty(TEXT);
