-- migrations/000002_create_indexes.up.sql
-- Adds all secondary indexes for MASS schema.

-- ==============================
-- PERSON CONTACT
-- ==============================

-- Lookup contacts by person
CREATE INDEX IF NOT EXISTS idx_person_contact_person_id
ON person_contact (person_id);

-- Lookup contacts by type
CREATE INDEX IF NOT EXISTS idx_person_contact_contact_type_id
ON person_contact (contact_type_id);

-- Fast lookup of primary contact
CREATE INDEX IF NOT EXISTS idx_person_contact_primary
ON person_contact (person_id)
WHERE is_primary = TRUE;

-- Optional search by contact value (email/phone)
CREATE INDEX IF NOT EXISTS idx_person_contact_value
ON person_contact (contact_value);

-- ==============================
-- PROVIDER SPECIALTY
-- ==============================

-- Find providers by specialty
CREATE INDEX IF NOT EXISTS idx_provider_specialty_specialty_id
ON provider_specialty (specialty_id);

-- ==============================
-- APPOINTMENTS (AVAILABILITY)
-- ==============================

-- Provider schedule lookups
CREATE INDEX IF NOT EXISTS idx_appointment_provider_start
ON appointment (provider_id, start_time);

-- Patient schedule lookups
CREATE INDEX IF NOT EXISTS idx_appointment_patient_start
ON appointment (patient_id, start_time);

-- Overlap checks for provider
CREATE INDEX IF NOT EXISTS idx_appointment_provider_time_window
ON appointment (provider_id, start_time, end_time);

-- Overlap checks for patient
CREATE INDEX IF NOT EXISTS idx_appointment_patient_time_window
ON appointment (patient_id, start_time, end_time);

-- General time filtering
CREATE INDEX IF NOT EXISTS idx_appointment_start_time
ON appointment (start_time);

-- Lookup by appointment type
CREATE INDEX IF NOT EXISTS idx_appointment_appt_type_id
ON appointment (appt_type_id);

-- Lookup by staff creator
CREATE INDEX IF NOT EXISTS idx_appointment_created_by
ON appointment (created_by);

-- ==============================
-- CANCELLATIONS
-- ==============================

-- Join by cancellation reason
CREATE INDEX IF NOT EXISTS idx_appt_cancellation_reason_id
ON appt_cancellation (reason_id);

-- Audit by staff
CREATE INDEX IF NOT EXISTS idx_appt_cancellation_recorded_by
ON appt_cancellation (recorded_by);

-- Filter by cancellation time
CREATE INDEX IF NOT EXISTS idx_appt_cancellation_cancelled_at
ON appt_cancellation (cancelled_at);