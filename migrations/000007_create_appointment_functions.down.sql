-- migrations/000007_create_appointment_functions.down.sql
-- Drops appointment availability helpers + CRUD/search functions.

-- ====================================================================================
-- DROP APPOINTMENT CRUD / LIST FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_appointment(BIGINT);
DROP FUNCTION IF EXISTS fn_update_appointment(BIGINT, TIMESTAMPTZ, TIMESTAMPTZ, TEXT, BIGINT, BIGINT, BIGINT, INT);
DROP FUNCTION IF EXISTS fn_get_appointments(BIGINT, BIGINT, INT, TIMESTAMPTZ, TIMESTAMPTZ, BOOLEAN, INT, INT);
DROP FUNCTION IF EXISTS fn_get_appointment(BIGINT);
DROP FUNCTION IF EXISTS fn_create_appointment(TIMESTAMPTZ, TIMESTAMPTZ, TEXT, BIGINT, BIGINT, BIGINT, INT);

-- ====================================================================================
-- DROP AVAILABILITY HELPER FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_get_patient_is_available(BIGINT, TIMESTAMPTZ, TIMESTAMPTZ, BIGINT);
DROP FUNCTION IF EXISTS fn_get_provider_is_available(BIGINT, TIMESTAMPTZ, TIMESTAMPTZ, BIGINT);
