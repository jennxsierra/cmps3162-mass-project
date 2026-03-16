-- migrations/000004_create_patient_functions.up.sql
-- Creates patient CRUD + search functions.

-- ====================================================================================
-- PATIENT: CREATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_patient(
    p_first_name TEXT,
    p_last_name TEXT,
    p_date_of_birth DATE DEFAULT NULL,
    p_gender TEXT DEFAULT NULL,
    p_patient_no TEXT DEFAULT NULL,
    p_ssn TEXT DEFAULT NULL
)
RETURNS TABLE (
    patient_id BIGINT,
    patient_no TEXT,
    ssn TEXT,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    gender TEXT,
    created_at TIMESTAMPTZ
) AS $$
DECLARE
    v_patient_id BIGINT;
BEGIN
    IF p_first_name IS NULL OR btrim(p_first_name) = '' THEN
        RAISE EXCEPTION '[patient-invalid] first_name is required';
    END IF;

    IF p_last_name IS NULL OR btrim(p_last_name) = '' THEN
        RAISE EXCEPTION '[patient-invalid] last_name is required';
    END IF;

    IF p_patient_no IS NULL OR btrim(p_patient_no) = '' THEN
        RAISE EXCEPTION '[patient-invalid] patient_no is required';
    END IF;

    IF p_ssn IS NULL OR btrim(p_ssn) = '' THEN
        RAISE EXCEPTION '[patient-invalid] ssn is required';
    END IF;

    INSERT INTO person (first_name, last_name, date_of_birth, gender)
    VALUES (btrim(p_first_name), btrim(p_last_name), p_date_of_birth, p_gender)
    RETURNING person_id INTO v_patient_id;

    INSERT INTO patient (patient_id, patient_no, ssn)
    VALUES (v_patient_id, btrim(p_patient_no), btrim(p_ssn));

    RETURN QUERY
    SELECT p.person_id, pa.patient_no, pa.ssn, p.first_name, p.last_name, p.date_of_birth, p.gender, p.created_at
    FROM person p
    JOIN patient pa ON pa.patient_id = p.person_id
    WHERE p.person_id = v_patient_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PATIENT: READ (BY PATIENT NO)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_patient(
    p_patient_no TEXT
)
RETURNS TABLE (
    patient_id BIGINT,
    patient_no TEXT,
    ssn TEXT,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    gender TEXT,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT p.person_id, pa.patient_no, pa.ssn, p.first_name, p.last_name, p.date_of_birth, p.gender, p.created_at
    FROM patient pa
    JOIN person p ON p.person_id = pa.patient_id
    WHERE pa.patient_no = p_patient_no;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[patient-not-found] patient with patient_no=% does not exist', p_patient_no;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PATIENT: LIST / SEARCH
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_patients(
    p_search TEXT DEFAULT NULL,
    p_patient_no TEXT DEFAULT NULL,
    p_limit INT DEFAULT 50,
    p_offset INT DEFAULT 0
)
RETURNS TABLE (
    total_records BIGINT,
    patient_id BIGINT,
    patient_no TEXT,
    ssn TEXT,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    gender TEXT,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    WITH filtered AS (
        SELECT
            p.person_id,
            pa.patient_no,
            pa.ssn,
            p.first_name,
            p.last_name,
            p.date_of_birth,
            p.gender,
            p.created_at
        FROM patient pa
        JOIN person p ON p.person_id = pa.patient_id
        WHERE (
            p_search IS NULL
            OR p.first_name ILIKE '%' || p_search || '%'
            OR p.last_name ILIKE '%' || p_search || '%'
            OR (p.first_name || ' ' || p.last_name) ILIKE '%' || p_search || '%'
        )
          AND (
            p_patient_no IS NULL
            OR pa.patient_no ILIKE '%' || p_patient_no || '%'
        )
    )
    SELECT
        count(*) OVER() AS total_records,
        f.person_id,
        f.patient_no,
        f.ssn,
        f.first_name,
        f.last_name,
        f.date_of_birth,
        f.gender,
        f.created_at
    FROM filtered f
    ORDER BY f.last_name, f.first_name, f.person_id
    LIMIT GREATEST(COALESCE(p_limit, 50), 0)
    OFFSET GREATEST(COALESCE(p_offset, 0), 0);
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PATIENT: UPDATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_update_patient(
    p_patient_no TEXT,
    p_first_name TEXT DEFAULT NULL,
    p_last_name TEXT DEFAULT NULL,
    p_date_of_birth DATE DEFAULT NULL,
    p_gender TEXT DEFAULT NULL,
    p_new_patient_no TEXT DEFAULT NULL,
    p_new_ssn TEXT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_patient_id BIGINT;
BEGIN
    SELECT pa.patient_id
      INTO v_patient_id
      FROM patient pa
     WHERE pa.patient_no = p_patient_no
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[patient-not-found] patient with patient_no=% does not exist', p_patient_no;
    END IF;

    UPDATE person
       SET first_name = COALESCE(p_first_name, first_name),
           last_name = COALESCE(p_last_name, last_name),
           date_of_birth = COALESCE(p_date_of_birth, date_of_birth),
           gender = COALESCE(p_gender, gender)
     WHERE person_id = v_patient_id;

    UPDATE patient
       SET patient_no = COALESCE(NULLIF(btrim(p_new_patient_no), ''), patient_no),
           ssn = COALESCE(NULLIF(btrim(p_new_ssn), ''), ssn)
     WHERE patient_id = v_patient_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PATIENT: DELETE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_delete_patient(
    p_patient_no TEXT
)
RETURNS VOID AS $$
DECLARE
    v_patient_id BIGINT;
BEGIN
    SELECT pa.patient_id
      INTO v_patient_id
      FROM patient pa
     WHERE pa.patient_no = p_patient_no
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[patient-not-found] patient with patient_no=% does not exist', p_patient_no;
    END IF;

    DELETE FROM person
     WHERE person_id = v_patient_id;
END;
$$ LANGUAGE plpgsql;
