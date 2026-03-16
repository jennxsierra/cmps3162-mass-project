-- migrations/000011_seed_mass_demo_data.up.sql
-- Seeds MASS demo data:
-- - 20 patients
-- - 10 providers
-- - 5 staff
-- - 10 specialties
-- - 15 active appointments (5 updated)
-- - 5 cancelled appointments

-- ====================================================================================
-- REFERENCE DATA
-- ====================================================================================

INSERT INTO specialty (specialty_name) VALUES
    ('Cardiology'),
    ('Dermatology'),
    ('Endocrinology'),
    ('Family Medicine'),
    ('Gastroenterology'),
    ('Neurology'),
    ('Obstetrics & Gynecology'),
    ('Oncology'),
    ('Pediatrics'),
    ('Psychiatry');

INSERT INTO appt_type (appt_type_name) VALUES
    ('New Patient Consultation'),
    ('Follow-up Visit'),
    ('Annual Physical Exam'),
    ('Telehealth Check-in');

INSERT INTO cancellation_reason (reason_name) VALUES
    ('Patient requested reschedule'),
    ('Provider emergency leave'),
    ('Transportation issue'),
    ('Severe weather conditions'),
    ('Insurance authorization pending');

INSERT INTO contact_type (contact_type_name) VALUES
    ('Mobile Phone'),
    ('Work Phone'),
    ('Personal Email'),
    ('Emergency Contact Phone');

-- ====================================================================================
-- PEOPLE DATA (PATIENTS, PROVIDERS, STAFF)
-- ====================================================================================

DO $$
DECLARE
    i INT;
    v_gender TEXT;
    v_patient_first_names TEXT[] := ARRAY[
        'Aaliyah','Brandon','Camila','Darius','Elena','Francis','Gabriela','Hector','Isabel','Jamal',
        'Kiara','Liam','Maya','Noah','Olivia','Priya','Quentin','Rosa','Samuel','Tiana'
    ];
    v_patient_last_names TEXT[] := ARRAY[
        'Bennett','Castillo','Diaz','Edwards','Flores','Grant','Hernandez','Irving','James','King',
        'Lopez','Martinez','Nunez','Owens','Peters','Quinn','Reyes','Santos','Turner','Vargas'
    ];
    v_provider_first_names TEXT[] := ARRAY[
        'Adrian','Bianca','Carlos','Danielle','Evan','Farah','Gavin','Hanna','Isaac','Jasmine'
    ];
    v_provider_last_names TEXT[] := ARRAY[
        'Coleman','Delgado','Ellis','Foster','Green','Hall','Ingram','Jordan','Khan','Lewis'
    ];
    v_staff_first_names TEXT[] := ARRAY[
        'Monica','Nathan','Paula','Renee','Tyler'
    ];
    v_staff_last_names TEXT[] := ARRAY[
        'Morris','Nelson','Ortiz','Price','Robinson'
    ];
BEGIN
    FOR i IN 1..20 LOOP
        v_gender := CASE WHEN i % 2 = 0 THEN 'Female' ELSE 'Male' END;

        PERFORM *
        FROM fn_create_patient(
            p_first_name => v_patient_first_names[i],
            p_last_name => v_patient_last_names[i],
            p_date_of_birth => (DATE '1980-01-01' + (i * INTERVAL '120 days'))::DATE,
            p_gender => v_gender,
            p_patient_no => format('SEED-PAT-%s', lpad(i::TEXT, 3, '0')),
            p_ssn => format('900-00-%s', lpad(i::TEXT, 4, '0'))
        );
    END LOOP;

    FOR i IN 1..10 LOOP
        v_gender := CASE WHEN i % 2 = 0 THEN 'Female' ELSE 'Male' END;

        PERFORM *
        FROM fn_create_provider(
            p_first_name => v_provider_first_names[i],
            p_last_name => v_provider_last_names[i],
            p_date_of_birth => (DATE '1975-01-01' + (i * INTERVAL '220 days'))::DATE,
            p_gender => v_gender,
            p_license_no => format('SEED-LIC-%s', lpad(i::TEXT, 3, '0'))
        );
    END LOOP;

    FOR i IN 1..5 LOOP
        v_gender := CASE WHEN i % 2 = 0 THEN 'Female' ELSE 'Male' END;

        PERFORM *
        FROM fn_create_staff(
            p_first_name => v_staff_first_names[i],
            p_last_name => v_staff_last_names[i],
            p_date_of_birth => (DATE '1988-01-01' + (i * INTERVAL '300 days'))::DATE,
            p_gender => v_gender,
            p_staff_no => format('SEED-STF-%s', lpad(i::TEXT, 3, '0'))
        );
    END LOOP;
END;
$$;

-- ====================================================================================
-- PERSON CONTACTS
-- ====================================================================================

DO $$
DECLARE
    i INT;
    v_person_id BIGINT;
    v_mobile_phone_type_id INT;
    v_work_phone_type_id INT;
    v_personal_email_type_id INT;
    v_emergency_phone_type_id INT;
BEGIN
    SELECT ct.contact_type_id INTO v_mobile_phone_type_id
    FROM contact_type ct
    WHERE ct.contact_type_name = 'Mobile Phone';

    SELECT ct.contact_type_id INTO v_work_phone_type_id
    FROM contact_type ct
    WHERE ct.contact_type_name = 'Work Phone';

    SELECT ct.contact_type_id INTO v_personal_email_type_id
    FROM contact_type ct
    WHERE ct.contact_type_name = 'Personal Email';

    SELECT ct.contact_type_id INTO v_emergency_phone_type_id
    FROM contact_type ct
    WHERE ct.contact_type_name = 'Emergency Contact Phone';

    FOR i IN 1..20 LOOP
        SELECT pa.patient_id INTO v_person_id
        FROM patient pa
        WHERE pa.patient_no = format('SEED-PAT-%s', lpad(i::TEXT, 3, '0'));

        INSERT INTO person_contact (contact_value, is_primary, person_id, contact_type_id)
        VALUES (format('+1-555-210-%s', lpad(i::TEXT, 4, '0')), TRUE, v_person_id, v_mobile_phone_type_id);

        INSERT INTO person_contact (contact_value, is_primary, person_id, contact_type_id)
        VALUES (format('patient%s@massdemo.org', lpad(i::TEXT, 2, '0')), FALSE, v_person_id, v_personal_email_type_id);

        INSERT INTO person_contact (contact_value, is_primary, person_id, contact_type_id)
        VALUES (format('+1-555-910-%s', lpad(i::TEXT, 4, '0')), FALSE, v_person_id, v_emergency_phone_type_id);
    END LOOP;

    FOR i IN 1..10 LOOP
        SELECT pr.provider_id INTO v_person_id
        FROM provider pr
        WHERE pr.license_no = format('SEED-LIC-%s', lpad(i::TEXT, 3, '0'));

        INSERT INTO person_contact (contact_value, is_primary, person_id, contact_type_id)
        VALUES (format('+1-555-320-%s', lpad(i::TEXT, 4, '0')), TRUE, v_person_id, v_work_phone_type_id);

        INSERT INTO person_contact (contact_value, is_primary, person_id, contact_type_id)
        VALUES (format('provider%s@massdemo.org', lpad(i::TEXT, 2, '0')), FALSE, v_person_id, v_personal_email_type_id);
    END LOOP;

    FOR i IN 1..5 LOOP
        SELECT s.staff_id INTO v_person_id
        FROM staff s
        WHERE s.staff_no = format('SEED-STF-%s', lpad(i::TEXT, 3, '0'));

        INSERT INTO person_contact (contact_value, is_primary, person_id, contact_type_id)
        VALUES (format('+1-555-420-%s', lpad(i::TEXT, 4, '0')), TRUE, v_person_id, v_work_phone_type_id);

        INSERT INTO person_contact (contact_value, is_primary, person_id, contact_type_id)
        VALUES (format('staff%s@massdemo.org', lpad(i::TEXT, 2, '0')), FALSE, v_person_id, v_personal_email_type_id);
    END LOOP;
END;
$$;

-- ====================================================================================
-- PROVIDER-SPECIALTY LINKS
-- ====================================================================================

DO $$
DECLARE
    i INT;
BEGIN
    FOR i IN 1..10 LOOP
        PERFORM *
        FROM fn_create_provider_specialty(
            p_license_no => format('SEED-LIC-%s', lpad(i::TEXT, 3, '0')),
            p_specialty_id => i
        );
    END LOOP;
END;
$$;

-- ====================================================================================
-- APPOINTMENTS + CANCELLATIONS
-- ====================================================================================

DO $$
DECLARE
    i INT;
    v_appt_id BIGINT;
    v_start TIMESTAMPTZ;
    v_end TIMESTAMPTZ;
    v_staff_id BIGINT;
    v_cancel_reason_id INT;
    v_appt_type_ids INT[];
    v_staff_ids BIGINT[];
    v_cancel_reason_ids INT[];
    v_appt_reasons TEXT[] := ARRAY[
        'Annual wellness check and preventive care review',
        'Persistent migraines and headache management',
        'Follow-up on hypertension medication adjustment',
        'Routine pediatric growth and development visit',
        'Type 2 diabetes blood sugar review',
        'Dermatitis flare-up evaluation and treatment',
        'Prenatal consultation and prenatal vitamins review',
        'Post-hospital discharge medication reconciliation',
        'Chronic lower back pain assessment',
        'Thyroid lab review and dosage update',
        'Anxiety symptoms follow-up and coping plan',
        'Gastrointestinal discomfort and reflux evaluation',
        'Cardiology consult for palpitations',
        'Neurology follow-up for neuropathy symptoms',
        'Sports physical clearance exam',
        'Seasonal allergy symptom management',
        'Medication side-effect review',
        'Telehealth check for recent lab results',
        'Pediatric fever follow-up after urgent care',
        'Oncology treatment planning discussion'
    ];
    v_updated_reasons TEXT[] := ARRAY[
        'Annual wellness check updated with fasting-lab request',
        'Migraine follow-up updated with trigger journal review',
        'Hypertension visit updated with home BP log review',
        'Pediatric checkup updated with vaccination discussion',
        'Diabetes follow-up updated with nutrition counseling'
    ];
    v_cancel_notes TEXT[] := ARRAY[
        'Patient called to reschedule due to work conflict.',
        'Provider unavailable because of urgent hospital call.',
        'Patient could not arrange transportation in time.',
        'Clinic reduced operations due to severe weather alert.',
        'Visit postponed while insurance pre-authorization is processed.'
    ];
BEGIN
    SELECT array_agg(atp.appt_type_id ORDER BY atp.appt_type_name)
      INTO v_appt_type_ids
      FROM appt_type atp
     WHERE atp.appt_type_name IN (
         'New Patient Consultation',
         'Follow-up Visit',
         'Annual Physical Exam',
         'Telehealth Check-in'
     );

    SELECT array_agg(s.staff_id ORDER BY s.staff_no)
      INTO v_staff_ids
      FROM staff s
     WHERE s.staff_no LIKE 'SEED-STF-%';

    SELECT array_agg(cr.reason_id ORDER BY cr.reason_name)
      INTO v_cancel_reason_ids
      FROM cancellation_reason cr
         WHERE cr.reason_name IN (
                 'Patient requested reschedule',
                 'Provider emergency leave',
                 'Transportation issue',
                 'Severe weather conditions',
                 'Insurance authorization pending'
         );

    CREATE TEMP TABLE tmp_seed_appointments (
        seq INT PRIMARY KEY,
        appointment_id BIGINT NOT NULL
    ) ON COMMIT DROP;

    FOR i IN 1..20 LOOP
        v_start := TIMESTAMPTZ '2026-04-01 09:00:00+00' + ((i - 1) * INTERVAL '1 day');
        v_end := v_start + INTERVAL '45 minutes';
        v_staff_id := v_staff_ids[((i - 1) % 5) + 1];

        SELECT c.appointment_id
          INTO v_appt_id
          FROM fn_create_appointment(
              p_start_time => v_start,
              p_end_time => v_end,
              p_reason => v_appt_reasons[i],
              p_patient_id => (SELECT pa.patient_id FROM patient pa WHERE pa.patient_no = format('SEED-PAT-%s', lpad(i::TEXT, 3, '0'))),
              p_provider_id => (SELECT pr.provider_id FROM provider pr WHERE pr.license_no = format('SEED-LIC-%s', lpad((((i - 1) % 10) + 1)::TEXT, 3, '0'))),
              p_created_by => v_staff_id,
              p_appt_type_id => v_appt_type_ids[((i - 1) % array_length(v_appt_type_ids, 1)) + 1]
          ) AS c;

        INSERT INTO tmp_seed_appointments (seq, appointment_id)
        VALUES (i, v_appt_id);
    END LOOP;

    -- Update 5 active appointments
    FOR i IN 1..5 LOOP
        PERFORM fn_update_appointment(
            p_appointment_id => (SELECT tsa.appointment_id FROM tmp_seed_appointments tsa WHERE tsa.seq = i),
            p_reason => v_updated_reasons[i]
        );
    END LOOP;

    -- Cancel 5 appointments (seq 16-20), leaving 15 active
    FOR i IN 16..20 LOOP
        v_staff_id := v_staff_ids[((i - 1) % 5) + 1];
        v_cancel_reason_id := v_cancel_reason_ids[((i - 16) % array_length(v_cancel_reason_ids, 1)) + 1];

        PERFORM *
        FROM fn_create_appt_cancellation(
            p_appointment_id => (SELECT tsa.appointment_id FROM tmp_seed_appointments tsa WHERE tsa.seq = i),
            p_note => v_cancel_notes[i - 15],
            p_reason_id => v_cancel_reason_id,
            p_recorded_by => v_staff_id,
            p_cancelled_at => (SELECT a.start_time - INTERVAL '1 day' FROM appointment a WHERE a.appointment_id = (SELECT tsa.appointment_id FROM tmp_seed_appointments tsa WHERE tsa.seq = i))
        );
    END LOOP;
END;
$$;