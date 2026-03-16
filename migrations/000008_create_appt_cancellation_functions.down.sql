-- migrations/000008_create_appt_cancellation_functions.down.sql
-- Drops appointment cancellation CRUD functions.

-- ====================================================================================
-- DROP FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_appt_cancellation(BIGINT);
DROP FUNCTION IF EXISTS fn_update_appt_cancellation(BIGINT, TEXT, INT, BIGINT);
DROP FUNCTION IF EXISTS fn_get_appt_cancellation(BIGINT);
DROP FUNCTION IF EXISTS fn_create_appt_cancellation(BIGINT, TEXT, INT, BIGINT, TIMESTAMPTZ);
