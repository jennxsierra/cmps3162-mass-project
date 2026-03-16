-- migrations/000001_create_schema.down.sql
-- Drops the entire Medical Appointment Scheduling (MASS) database schema (Schema v1.0).

-- APPOINTMENT CANCELLATIONS
DROP TABLE IF EXISTS appt_cancellation;

-- APPOINTMENTS
DROP TABLE IF EXISTS appointment;

-- MANY-TO-MANY
DROP TABLE IF EXISTS provider_specialty;

-- REFERENCE TABLES
DROP TABLE IF EXISTS cancellation_reason;
DROP TABLE IF EXISTS appt_type;
DROP TABLE IF EXISTS specialty;

-- ROLES (SUBTYPES OF PERSON)
DROP TABLE IF EXISTS patient;
DROP TABLE IF EXISTS staff;
DROP TABLE IF EXISTS provider;

-- CORE ENTITIES
DROP TABLE IF EXISTS person_contact;
DROP TABLE IF EXISTS contact_type;
DROP TABLE IF EXISTS person;

-- EXTENSIONS
DROP EXTENSION IF EXISTS citext;