-- migrations/000010_create_person_contact_functions.down.sql
-- Drops person_contact CRUD + list functions and primary-contact uniqueness index.

-- ====================================================================================
-- DROP PERSON CONTACT FUNCTIONS (REVERSE ORDER)
-- ====================================================================================

DROP FUNCTION IF EXISTS fn_delete_person_contact(BIGINT);
DROP FUNCTION IF EXISTS fn_update_person_contact(BIGINT, TEXT, BOOLEAN, INT);
DROP FUNCTION IF EXISTS fn_get_person_contacts(BIGINT, INT, BOOLEAN);
DROP FUNCTION IF EXISTS fn_get_person_contact(BIGINT);
DROP FUNCTION IF EXISTS fn_create_person_contact(BIGINT, INT, TEXT, BOOLEAN);

-- ====================================================================================
-- DROP INDEXES / CONSTRAINT HELPERS
-- ====================================================================================

DROP INDEX IF EXISTS ux_person_contact_primary_per_type;
