-- migrations/000001_create_schema.up.sql
-- Creates the entire Medical Appointment Scheduling (MASS) database schema (Schema v1.0).

-- ====================================================================================
-- EXTENSIONS
-- ====================================================================================

CREATE EXTENSION IF NOT EXISTS citext;

-- ====================================================================================
-- CORE ENTITIES
-- ====================================================================================

CREATE TABLE person (
    person_id   BIGSERIAL PRIMARY KEY,
    first_name  TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    date_of_birth DATE,
    gender      TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE contact_type (
    contact_type_id   SERIAL PRIMARY KEY,
    contact_type_name TEXT UNIQUE NOT NULL
);

CREATE TABLE person_contact (
    person_contact_id BIGSERIAL PRIMARY KEY,
    contact_value     TEXT NOT NULL,
    is_primary        BOOLEAN NOT NULL DEFAULT FALSE,
    person_id         BIGINT NOT NULL REFERENCES person(person_id) ON DELETE CASCADE,
    contact_type_id   INT NOT NULL REFERENCES contact_type(contact_type_id)
);

-- ====================================================================================
-- ROLES (SUBTYPES OF PERSON)
-- ====================================================================================

CREATE TABLE provider (
    provider_id BIGINT PRIMARY KEY REFERENCES person(person_id) ON DELETE CASCADE,
    license_no  TEXT UNIQUE NOT NULL
);

CREATE TABLE staff (
    staff_id  BIGINT PRIMARY KEY REFERENCES person(person_id) ON DELETE CASCADE,
    staff_no  TEXT UNIQUE NOT NULL
);

CREATE TABLE patient (
    patient_id  BIGINT PRIMARY KEY REFERENCES person(person_id) ON DELETE CASCADE,
    patient_no  TEXT UNIQUE NOT NULL,
    ssn         TEXT UNIQUE NOT NULL
);

-- ====================================================================================
-- REFERENCE TABLES
-- ====================================================================================

CREATE TABLE specialty (
    specialty_id   SERIAL PRIMARY KEY,
    specialty_name TEXT UNIQUE NOT NULL
);

CREATE TABLE appt_type (
    appt_type_id   SERIAL PRIMARY KEY,
    appt_type_name TEXT UNIQUE NOT NULL
);

CREATE TABLE cancellation_reason (
    reason_id   SERIAL PRIMARY KEY,
    reason_name TEXT UNIQUE NOT NULL
);

-- ====================================================================================
-- MANY-TO-MANY PROVIDER-SPECIALTY RELATIONSHIP
-- ====================================================================================

CREATE TABLE provider_specialty (
    provider_id  BIGINT NOT NULL REFERENCES provider(provider_id) ON DELETE CASCADE,
    specialty_id INT NOT NULL REFERENCES specialty(specialty_id) ON DELETE CASCADE,
    PRIMARY KEY (provider_id, specialty_id)
);

-- ====================================================================================
-- APPOINTMENTS
-- ====================================================================================

CREATE TABLE appointment (
    appointment_id BIGSERIAL PRIMARY KEY,
    start_time     TIMESTAMPTZ NOT NULL,
    end_time       TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ,
    reason         TEXT,
    patient_id     BIGINT NOT NULL REFERENCES patient(patient_id),
    provider_id    BIGINT NOT NULL REFERENCES provider(provider_id),
    created_by     BIGINT NOT NULL REFERENCES staff(staff_id),
    appt_type_id   INT NOT NULL REFERENCES appt_type(appt_type_id),
    CONSTRAINT appointment_end_after_start CHECK (end_time > start_time)
);

-- ====================================================================================
-- APPOINTMENT CANCELLATIONS
-- ====================================================================================

CREATE TABLE appt_cancellation (
    appointment_id BIGINT PRIMARY KEY REFERENCES appointment(appointment_id) ON DELETE CASCADE,
    cancelled_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note           TEXT,
    reason_id      INT NOT NULL REFERENCES cancellation_reason(reason_id),
    recorded_by    BIGINT NOT NULL REFERENCES staff(staff_id)
);