-- migrations/000007_create_appointment_functions.up.sql
-- Creates appointment availability helpers + CRUD/search functions.

-- ====================================================================================
-- APPOINTMENT: AVAILABILITY HELPER (PROVIDER)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_provider_is_available(
    p_provider_id BIGINT,
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ,
    p_ignore_appointment_id BIGINT DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    v_conflict_exists BOOLEAN;
BEGIN
    IF p_end_time <= p_start_time THEN
        RAISE EXCEPTION '[appointment-invalid] end_time must be greater than start_time';
    END IF;

    SELECT EXISTS (
        SELECT 1
        FROM appointment a
        LEFT JOIN appt_cancellation ac ON ac.appointment_id = a.appointment_id
        WHERE a.provider_id = p_provider_id
          AND ac.appointment_id IS NULL
          AND (p_ignore_appointment_id IS NULL OR a.appointment_id <> p_ignore_appointment_id)
          AND a.start_time < p_end_time
          AND a.end_time > p_start_time
    )
    INTO v_conflict_exists;

    RETURN NOT v_conflict_exists;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT: AVAILABILITY HELPER (PATIENT)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_patient_is_available(
    p_patient_id BIGINT,
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ,
    p_ignore_appointment_id BIGINT DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    v_conflict_exists BOOLEAN;
BEGIN
    IF p_end_time <= p_start_time THEN
        RAISE EXCEPTION '[appointment-invalid] end_time must be greater than start_time';
    END IF;

    SELECT EXISTS (
        SELECT 1
        FROM appointment a
        LEFT JOIN appt_cancellation ac ON ac.appointment_id = a.appointment_id
        WHERE a.patient_id = p_patient_id
          AND ac.appointment_id IS NULL
          AND (p_ignore_appointment_id IS NULL OR a.appointment_id <> p_ignore_appointment_id)
          AND a.start_time < p_end_time
          AND a.end_time > p_start_time
    )
    INTO v_conflict_exists;

    RETURN NOT v_conflict_exists;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT: CREATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_appointment(
    p_start_time TIMESTAMPTZ,
    p_end_time TIMESTAMPTZ,
    p_reason TEXT,
    p_patient_id BIGINT,
    p_provider_id BIGINT,
    p_created_by BIGINT,
    p_appt_type_id INT
)
RETURNS TABLE (
    appointment_id BIGINT
) AS $$
DECLARE
    v_appointment_id BIGINT;
BEGIN
    IF p_end_time <= p_start_time THEN
        RAISE EXCEPTION '[appointment-invalid] end_time must be greater than start_time';
    END IF;

    PERFORM 1 FROM patient pa WHERE pa.patient_id = p_patient_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[patient-not-found] patient with patient_id=% does not exist', p_patient_id;
    END IF;

    PERFORM 1 FROM provider pr WHERE pr.provider_id = p_provider_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-not-found] provider with provider_id=% does not exist', p_provider_id;
    END IF;

    PERFORM 1 FROM staff s WHERE s.staff_id = p_created_by;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[staff-not-found] staff with staff_id=% does not exist', p_created_by;
    END IF;

    PERFORM 1 FROM appt_type atp WHERE atp.appt_type_id = p_appt_type_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[appt-type-not-found] appt_type with appt_type_id=% does not exist', p_appt_type_id;
    END IF;

    IF NOT fn_get_provider_is_available(p_provider_id, p_start_time, p_end_time, NULL) THEN
        RAISE EXCEPTION '[provider-unavailable] provider_id=% has overlapping appointment for [% - %)', p_provider_id, p_start_time, p_end_time;
    END IF;

    IF NOT fn_get_patient_is_available(p_patient_id, p_start_time, p_end_time, NULL) THEN
        RAISE EXCEPTION '[patient-unavailable] patient_id=% has overlapping appointment for [% - %)', p_patient_id, p_start_time, p_end_time;
    END IF;

    INSERT INTO appointment (
        start_time,
        end_time,
        reason,
        patient_id,
        provider_id,
        created_by,
        appt_type_id
    )
    VALUES (
        p_start_time,
        p_end_time,
        p_reason,
        p_patient_id,
        p_provider_id,
        p_created_by,
        p_appt_type_id
    )
    RETURNING appointment.appointment_id INTO v_appointment_id;

    RETURN QUERY SELECT v_appointment_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT: READ (BY APPOINTMENT ID)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_appointment(
    p_appointment_id BIGINT
)
RETURNS TABLE (
    appointment_id BIGINT,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    reason TEXT,
    patient_id BIGINT,
    provider_id BIGINT,
    created_by BIGINT,
    appt_type_id INT,
    appt_type_name TEXT,
    patient_name TEXT,
    provider_name TEXT,
    is_cancelled BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        a.appointment_id,
        a.start_time,
        a.end_time,
        a.created_at,
        a.updated_at,
        a.reason,
        a.patient_id,
        a.provider_id,
        a.created_by,
        a.appt_type_id,
        atp.appt_type_name,
        (pp.first_name || ' ' || pp.last_name) AS patient_name,
        (prp.first_name || ' ' || prp.last_name) AS provider_name,
        (ac.appointment_id IS NOT NULL) AS is_cancelled
    FROM appointment a
    JOIN appt_type atp ON atp.appt_type_id = a.appt_type_id
    JOIN person pp ON pp.person_id = a.patient_id
    JOIN person prp ON prp.person_id = a.provider_id
    LEFT JOIN appt_cancellation ac ON ac.appointment_id = a.appointment_id
    WHERE a.appointment_id = p_appointment_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appointment-not-found] appointment with appointment_id=% does not exist', p_appointment_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT: LIST / SEARCH
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_appointments(
    p_provider_id BIGINT DEFAULT NULL,
    p_patient_id BIGINT DEFAULT NULL,
    p_appt_type_id INT DEFAULT NULL,
    p_start_from TIMESTAMPTZ DEFAULT NULL,
    p_start_to TIMESTAMPTZ DEFAULT NULL,
    p_include_cancelled BOOLEAN DEFAULT TRUE,
    p_limit INT DEFAULT 50,
    p_offset INT DEFAULT 0
)
RETURNS TABLE (
    total_records BIGINT,
    appointment_id BIGINT,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    reason TEXT,
    patient_id BIGINT,
    provider_id BIGINT,
    created_by BIGINT,
    appt_type_id INT,
    is_cancelled BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    WITH filtered AS (
        SELECT
            a.appointment_id,
            a.start_time,
            a.end_time,
            a.created_at,
            a.updated_at,
            a.reason,
            a.patient_id,
            a.provider_id,
            a.created_by,
            a.appt_type_id,
            (ac.appointment_id IS NOT NULL) AS is_cancelled
        FROM appointment a
        LEFT JOIN appt_cancellation ac ON ac.appointment_id = a.appointment_id
        WHERE (p_provider_id IS NULL OR a.provider_id = p_provider_id)
          AND (p_patient_id IS NULL OR a.patient_id = p_patient_id)
          AND (p_appt_type_id IS NULL OR a.appt_type_id = p_appt_type_id)
          AND (p_start_from IS NULL OR a.start_time >= p_start_from)
          AND (p_start_to IS NULL OR a.start_time <= p_start_to)
          AND (p_include_cancelled OR ac.appointment_id IS NULL)
    )
    SELECT
        count(*) OVER() AS total_records,
        f.appointment_id,
        f.start_time,
        f.end_time,
        f.created_at,
        f.updated_at,
        f.reason,
        f.patient_id,
        f.provider_id,
        f.created_by,
        f.appt_type_id,
        f.is_cancelled
    FROM filtered f
    ORDER BY f.start_time, f.appointment_id
    LIMIT GREATEST(COALESCE(p_limit, 50), 0)
    OFFSET GREATEST(COALESCE(p_offset, 0), 0);
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT: UPDATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_update_appointment(
    p_appointment_id BIGINT,
    p_start_time TIMESTAMPTZ DEFAULT NULL,
    p_end_time TIMESTAMPTZ DEFAULT NULL,
    p_reason TEXT DEFAULT NULL,
    p_patient_id BIGINT DEFAULT NULL,
    p_provider_id BIGINT DEFAULT NULL,
    p_created_by BIGINT DEFAULT NULL,
    p_appt_type_id INT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_start_time TIMESTAMPTZ;
    v_end_time TIMESTAMPTZ;
    v_reason TEXT;
    v_patient_id BIGINT;
    v_provider_id BIGINT;
    v_created_by BIGINT;
    v_appt_type_id INT;
BEGIN
    SELECT
        a.start_time,
        a.end_time,
        a.reason,
        a.patient_id,
        a.provider_id,
        a.created_by,
        a.appt_type_id
      INTO
        v_start_time,
        v_end_time,
        v_reason,
        v_patient_id,
        v_provider_id,
        v_created_by,
        v_appt_type_id
      FROM appointment a
     WHERE a.appointment_id = p_appointment_id
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appointment-not-found] appointment with appointment_id=% does not exist', p_appointment_id;
    END IF;

    v_start_time := COALESCE(p_start_time, v_start_time);
    v_end_time := COALESCE(p_end_time, v_end_time);
    v_reason := COALESCE(p_reason, v_reason);
    v_patient_id := COALESCE(p_patient_id, v_patient_id);
    v_provider_id := COALESCE(p_provider_id, v_provider_id);
    v_created_by := COALESCE(p_created_by, v_created_by);
    v_appt_type_id := COALESCE(p_appt_type_id, v_appt_type_id);

    IF v_end_time <= v_start_time THEN
        RAISE EXCEPTION '[appointment-invalid] end_time must be greater than start_time';
    END IF;

    PERFORM 1 FROM patient pa WHERE pa.patient_id = v_patient_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[patient-not-found] patient with patient_id=% does not exist', v_patient_id;
    END IF;

    PERFORM 1 FROM provider pr WHERE pr.provider_id = v_provider_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-not-found] provider with provider_id=% does not exist', v_provider_id;
    END IF;

    PERFORM 1 FROM staff s WHERE s.staff_id = v_created_by;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[staff-not-found] staff with staff_id=% does not exist', v_created_by;
    END IF;

    PERFORM 1 FROM appt_type atp WHERE atp.appt_type_id = v_appt_type_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[appt-type-not-found] appt_type with appt_type_id=% does not exist', v_appt_type_id;
    END IF;

    IF NOT fn_get_provider_is_available(v_provider_id, v_start_time, v_end_time, p_appointment_id) THEN
        RAISE EXCEPTION '[provider-unavailable] provider_id=% has overlapping appointment for [% - %)', v_provider_id, v_start_time, v_end_time;
    END IF;

    IF NOT fn_get_patient_is_available(v_patient_id, v_start_time, v_end_time, p_appointment_id) THEN
        RAISE EXCEPTION '[patient-unavailable] patient_id=% has overlapping appointment for [% - %)', v_patient_id, v_start_time, v_end_time;
    END IF;

    UPDATE appointment
       SET start_time = v_start_time,
           end_time = v_end_time,
           reason = v_reason,
           patient_id = v_patient_id,
           provider_id = v_provider_id,
           created_by = v_created_by,
           appt_type_id = v_appt_type_id,
           updated_at = NOW()
     WHERE appointment_id = p_appointment_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT: DELETE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_delete_appointment(
    p_appointment_id BIGINT
)
RETURNS VOID AS $$
BEGIN
    DELETE FROM appointment
     WHERE appointment_id = p_appointment_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appointment-not-found] appointment with appointment_id=% does not exist', p_appointment_id;
    END IF;
END;
$$ LANGUAGE plpgsql;
