-- migrations/000005_create_provider_functions.down.sql
-- Drops provider CRUD + specialty management functions.

-- ====================================================================================
-- DROP PROVIDER SPECIALTY FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_get_provider_specialties(TEXT);
DROP FUNCTION IF EXISTS fn_delete_provider_specialty(TEXT, INT);
DROP FUNCTION IF EXISTS fn_create_provider_specialty(TEXT, INT);

-- ====================================================================================
-- DROP PROVIDER FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_provider(TEXT);
DROP FUNCTION IF EXISTS fn_update_provider(TEXT, TEXT, TEXT, DATE, TEXT, TEXT);
DROP FUNCTION IF EXISTS fn_get_providers(TEXT, INT, INT, INT);
DROP FUNCTION IF EXISTS fn_get_provider(TEXT);
DROP FUNCTION IF EXISTS fn_create_provider(TEXT, TEXT, DATE, TEXT, TEXT);
