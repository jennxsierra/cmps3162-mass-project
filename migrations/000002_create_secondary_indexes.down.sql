-- migrations/000002_create_indexes.down.sql
-- Drops all secondary indexes.

DROP INDEX IF EXISTS idx_appt_cancellation_cancelled_at;
DROP INDEX IF EXISTS idx_appt_cancellation_recorded_by;
DROP INDEX IF EXISTS idx_appt_cancellation_reason_id;

DROP INDEX IF EXISTS idx_appointment_created_by;
DROP INDEX IF EXISTS idx_appointment_appt_type_id;
DROP INDEX IF EXISTS idx_appointment_start_time;
DROP INDEX IF EXISTS idx_appointment_patient_time_window;
DROP INDEX IF EXISTS idx_appointment_provider_time_window;
DROP INDEX IF EXISTS idx_appointment_patient_start;
DROP INDEX IF EXISTS idx_appointment_provider_start;

DROP INDEX IF EXISTS idx_provider_specialty_specialty_id;

DROP INDEX IF EXISTS idx_person_contact_value;
DROP INDEX IF EXISTS idx_person_contact_primary;
DROP INDEX IF EXISTS idx_person_contact_contact_type_id;
DROP INDEX IF EXISTS idx_person_contact_person_id;