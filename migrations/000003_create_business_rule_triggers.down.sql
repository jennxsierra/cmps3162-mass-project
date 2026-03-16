-- migrations/000003_create_business_rule_triggers.down.sql
-- Drops business-rule trigger functions and triggers for MASS schema.

-- Drop triggers first

DROP TRIGGER IF EXISTS trg_appointment_block_edit_if_cancelled ON appointment;
DROP TRIGGER IF EXISTS trg_appt_cancellation_validate_window ON appt_cancellation;
DROP TRIGGER IF EXISTS trg_person_contact_normalize_value ON person_contact;
DROP TRIGGER IF EXISTS trg_person_contact_single_primary ON person_contact;
DROP TRIGGER IF EXISTS trg_appointment_prevent_overlap_patient ON appointment;
DROP TRIGGER IF EXISTS trg_appointment_prevent_overlap_provider ON appointment;
DROP TRIGGER IF EXISTS trg_appointment_set_updated_at ON appointment;

-- Drop functions

DROP FUNCTION IF EXISTS fn_appointment_block_edit_if_cancelled();
DROP FUNCTION IF EXISTS fn_appt_cancellation_validate_window();
DROP FUNCTION IF EXISTS fn_person_contact_normalize_value();
DROP FUNCTION IF EXISTS fn_person_contact_single_primary();
DROP FUNCTION IF EXISTS fn_appointment_prevent_overlap_patient();
DROP FUNCTION IF EXISTS fn_appointment_prevent_overlap_provider();
DROP FUNCTION IF EXISTS fn_appointment_set_updated_at();
