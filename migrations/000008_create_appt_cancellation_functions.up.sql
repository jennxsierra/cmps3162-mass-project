-- migrations/000008_create_appt_cancellation_functions.up.sql
-- Creates appointment cancellation CRUD functions.

-- ====================================================================================
-- APPOINTMENT CANCELLATION: CREATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_appt_cancellation(
    p_appointment_id BIGINT,
    p_note TEXT,
    p_reason_id INT,
    p_recorded_by BIGINT,
    p_cancelled_at TIMESTAMPTZ DEFAULT NULL
)
RETURNS TABLE (
    appointment_id BIGINT,
    cancelled_at TIMESTAMPTZ,
    note TEXT,
    reason_id INT,
    recorded_by BIGINT
) AS $$
DECLARE
    v_cancelled_at TIMESTAMPTZ;
BEGIN
    PERFORM 1
      FROM appointment a
     WHERE a.appointment_id = p_appointment_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appointment-not-found] appointment with appointment_id=% does not exist', p_appointment_id;
    END IF;

    PERFORM 1
      FROM appt_cancellation ac
     WHERE ac.appointment_id = p_appointment_id;

    IF FOUND THEN
        RAISE EXCEPTION '[appointment-already-cancelled] appointment_id=% is already cancelled', p_appointment_id;
    END IF;

    PERFORM 1
      FROM cancellation_reason cr
     WHERE cr.reason_id = p_reason_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[cancellation-reason-not-found] cancellation_reason with reason_id=% does not exist', p_reason_id;
    END IF;

    PERFORM 1
      FROM staff s
     WHERE s.staff_id = p_recorded_by;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[staff-not-found] staff with staff_id=% does not exist', p_recorded_by;
    END IF;

    v_cancelled_at := COALESCE(p_cancelled_at, NOW());

    INSERT INTO appt_cancellation (
        appointment_id,
        cancelled_at,
        note,
        reason_id,
        recorded_by
    )
    VALUES (
        p_appointment_id,
        v_cancelled_at,
        p_note,
        p_reason_id,
        p_recorded_by
    );

    RETURN QUERY
    SELECT ac.appointment_id, ac.cancelled_at, ac.note, ac.reason_id, ac.recorded_by
    FROM appt_cancellation ac
    WHERE ac.appointment_id = p_appointment_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT CANCELLATION: READ (BY APPOINTMENT ID)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_appt_cancellation(
    p_appointment_id BIGINT
)
RETURNS TABLE (
    appointment_id BIGINT,
    cancelled_at TIMESTAMPTZ,
    note TEXT,
    reason_id INT,
    reason_name TEXT,
    recorded_by BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ac.appointment_id,
        ac.cancelled_at,
        ac.note,
        ac.reason_id,
        cr.reason_name,
        ac.recorded_by
    FROM appt_cancellation ac
    JOIN cancellation_reason cr ON cr.reason_id = ac.reason_id
    WHERE ac.appointment_id = p_appointment_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appointment-cancellation-not-found] cancellation for appointment_id=% does not exist', p_appointment_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT CANCELLATION: UPDATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_update_appt_cancellation(
    p_appointment_id BIGINT,
    p_note TEXT DEFAULT NULL,
    p_reason_id INT DEFAULT NULL,
    p_recorded_by BIGINT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_reason_id INT;
    v_recorded_by BIGINT;
BEGIN
    SELECT ac.reason_id, ac.recorded_by
      INTO v_reason_id, v_recorded_by
      FROM appt_cancellation ac
     WHERE ac.appointment_id = p_appointment_id
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appointment-cancellation-not-found] cancellation for appointment_id=% does not exist', p_appointment_id;
    END IF;

    v_reason_id := COALESCE(p_reason_id, v_reason_id);
    v_recorded_by := COALESCE(p_recorded_by, v_recorded_by);

    PERFORM 1
      FROM cancellation_reason cr
     WHERE cr.reason_id = v_reason_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[cancellation-reason-not-found] cancellation_reason with reason_id=% does not exist', v_reason_id;
    END IF;

    PERFORM 1
      FROM staff s
     WHERE s.staff_id = v_recorded_by;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[staff-not-found] staff with staff_id=% does not exist', v_recorded_by;
    END IF;

    UPDATE appt_cancellation
       SET note = COALESCE(p_note, note),
           reason_id = v_reason_id,
           recorded_by = v_recorded_by
     WHERE appointment_id = p_appointment_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPOINTMENT CANCELLATION: DELETE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_delete_appt_cancellation(
    p_appointment_id BIGINT
)
RETURNS VOID AS $$
BEGIN
    DELETE FROM appt_cancellation
     WHERE appointment_id = p_appointment_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appointment-cancellation-not-found] cancellation for appointment_id=% does not exist', p_appointment_id;
    END IF;
END;
$$ LANGUAGE plpgsql;
