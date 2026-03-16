-- migrations/000011_seed_mass_demo_data.down.sql
-- Removes data seeded by 000011_seed_mass_demo_data.up.sql.

-- ====================================================================================
-- APPOINTMENTS + CANCELLATIONS
-- ====================================================================================

DELETE FROM appt_cancellation ac
USING appointment a
JOIN patient pa ON pa.patient_id = a.patient_id
WHERE ac.appointment_id = a.appointment_id
  AND pa.patient_no LIKE 'SEED-PAT-%';

DELETE FROM appointment a
USING patient pa
WHERE a.patient_id = pa.patient_id
  AND pa.patient_no LIKE 'SEED-PAT-%';

-- ====================================================================================
-- PROVIDER-SPECIALTY LINKS
-- ====================================================================================

DELETE FROM provider_specialty ps
USING provider pr
WHERE ps.provider_id = pr.provider_id
  AND pr.license_no LIKE 'SEED-LIC-%';

-- ====================================================================================
-- PEOPLE DATA (PATIENTS, PROVIDERS, STAFF)
-- ====================================================================================

DELETE FROM person p
USING patient pa
WHERE p.person_id = pa.patient_id
  AND pa.patient_no LIKE 'SEED-PAT-%';

DELETE FROM person p
USING provider pr
WHERE p.person_id = pr.provider_id
  AND pr.license_no LIKE 'SEED-LIC-%';

DELETE FROM person p
USING staff s
WHERE p.person_id = s.staff_id
  AND s.staff_no LIKE 'SEED-STF-%';

-- ====================================================================================
-- REFERENCE DATA
-- ====================================================================================

DELETE FROM specialty
WHERE specialty_name IN (
  'Cardiology',
  'Dermatology',
  'Endocrinology',
  'Family Medicine',
  'Gastroenterology',
  'Neurology',
  'Obstetrics & Gynecology',
  'Oncology',
  'Pediatrics',
  'Psychiatry'
);

DELETE FROM appt_type
WHERE appt_type_name IN (
  'New Patient Consultation',
  'Follow-up Visit',
  'Annual Physical Exam',
  'Telehealth Check-in'
);

DELETE FROM cancellation_reason
WHERE reason_name IN (
  'Patient requested reschedule',
  'Provider emergency leave',
  'Transportation issue',
  'Severe weather conditions',
  'Insurance authorization pending'
);

DELETE FROM contact_type
WHERE contact_type_name IN (
  'Mobile Phone',
  'Work Phone',
  'Personal Email',
  'Emergency Contact Phone'
);