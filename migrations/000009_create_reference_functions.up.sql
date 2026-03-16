-- migrations/000009_create_reference_functions.up.sql
-- Creates CRUD + list/search functions for reference tables.

-- ====================================================================================
-- SPECIALTY
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_specialty(
    p_specialty_name TEXT
)
RETURNS TABLE (
    specialty_id INT,
    specialty_name TEXT
) AS $$
BEGIN
    IF p_specialty_name IS NULL OR btrim(p_specialty_name) = '' THEN
        RAISE EXCEPTION '[specialty-invalid] specialty_name is required';
    END IF;

    INSERT INTO specialty (specialty_name)
    VALUES (btrim(p_specialty_name))
    RETURNING specialty.specialty_id, specialty.specialty_name
    INTO specialty_id, specialty_name;

    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_specialty(
    p_specialty_id INT
)
RETURNS TABLE (
    specialty_id INT,
    specialty_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT s.specialty_id, s.specialty_name
    FROM specialty s
    WHERE s.specialty_id = p_specialty_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[specialty-not-found] specialty with specialty_id=% does not exist', p_specialty_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_specialties(
    p_search TEXT DEFAULT NULL,
    p_limit INT DEFAULT 50,
    p_offset INT DEFAULT 0
)
RETURNS TABLE (
    total_records BIGINT,
    specialty_id INT,
    specialty_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH filtered AS (
        SELECT s.specialty_id, s.specialty_name
        FROM specialty s
        WHERE p_search IS NULL
           OR s.specialty_name ILIKE '%' || p_search || '%'
    )
    SELECT
        count(*) OVER() AS total_records,
        f.specialty_id,
        f.specialty_name
    FROM filtered f
    ORDER BY f.specialty_name
    LIMIT GREATEST(COALESCE(p_limit, 50), 0)
    OFFSET GREATEST(COALESCE(p_offset, 0), 0);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_update_specialty(
    p_specialty_id INT,
    p_specialty_name TEXT
)
RETURNS VOID AS $$
BEGIN
    UPDATE specialty
       SET specialty_name = COALESCE(NULLIF(btrim(p_specialty_name), ''), specialty_name)
     WHERE specialty_id = p_specialty_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[specialty-not-found] specialty with specialty_id=% does not exist', p_specialty_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_delete_specialty(
    p_specialty_id INT
)
RETURNS VOID AS $$
BEGIN
    PERFORM 1
    FROM provider_specialty ps
    WHERE ps.specialty_id = p_specialty_id;

    IF FOUND THEN
        RAISE EXCEPTION '[specialty-in-use] specialty_id=% is referenced by provider_specialty', p_specialty_id;
    END IF;

    DELETE FROM specialty
    WHERE specialty_id = p_specialty_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[specialty-not-found] specialty with specialty_id=% does not exist', p_specialty_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- APPT_TYPE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_appt_type(
    p_appt_type_name TEXT
)
RETURNS TABLE (
    appt_type_id INT,
    appt_type_name TEXT
) AS $$
BEGIN
    IF p_appt_type_name IS NULL OR btrim(p_appt_type_name) = '' THEN
        RAISE EXCEPTION '[appt-type-invalid] appt_type_name is required';
    END IF;

    INSERT INTO appt_type (appt_type_name)
    VALUES (btrim(p_appt_type_name))
    RETURNING appt_type.appt_type_id, appt_type.appt_type_name
    INTO appt_type_id, appt_type_name;

    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_appt_type(
    p_appt_type_id INT
)
RETURNS TABLE (
    appt_type_id INT,
    appt_type_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT a.appt_type_id, a.appt_type_name
    FROM appt_type a
    WHERE a.appt_type_id = p_appt_type_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appt-type-not-found] appt_type with appt_type_id=% does not exist', p_appt_type_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_appt_types(
    p_search TEXT DEFAULT NULL,
    p_limit INT DEFAULT 50,
    p_offset INT DEFAULT 0
)
RETURNS TABLE (
    total_records BIGINT,
    appt_type_id INT,
    appt_type_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH filtered AS (
        SELECT a.appt_type_id, a.appt_type_name
        FROM appt_type a
        WHERE p_search IS NULL
           OR a.appt_type_name ILIKE '%' || p_search || '%'
    )
    SELECT
        count(*) OVER() AS total_records,
        f.appt_type_id,
        f.appt_type_name
    FROM filtered f
    ORDER BY f.appt_type_name
    LIMIT GREATEST(COALESCE(p_limit, 50), 0)
    OFFSET GREATEST(COALESCE(p_offset, 0), 0);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_update_appt_type(
    p_appt_type_id INT,
    p_appt_type_name TEXT
)
RETURNS VOID AS $$
BEGIN
    UPDATE appt_type
       SET appt_type_name = COALESCE(NULLIF(btrim(p_appt_type_name), ''), appt_type_name)
     WHERE appt_type_id = p_appt_type_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appt-type-not-found] appt_type with appt_type_id=% does not exist', p_appt_type_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_delete_appt_type(
    p_appt_type_id INT
)
RETURNS VOID AS $$
BEGIN
    PERFORM 1
    FROM appointment a
    WHERE a.appt_type_id = p_appt_type_id;

    IF FOUND THEN
        RAISE EXCEPTION '[appt-type-in-use] appt_type_id=% is referenced by appointment', p_appt_type_id;
    END IF;

    DELETE FROM appt_type
    WHERE appt_type_id = p_appt_type_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[appt-type-not-found] appt_type with appt_type_id=% does not exist', p_appt_type_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- CANCELLATION_REASON
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_cancellation_reason(
    p_reason_name TEXT
)
RETURNS TABLE (
    reason_id INT,
    reason_name TEXT
) AS $$
BEGIN
    IF p_reason_name IS NULL OR btrim(p_reason_name) = '' THEN
        RAISE EXCEPTION '[cancellation-reason-invalid] reason_name is required';
    END IF;

    INSERT INTO cancellation_reason (reason_name)
    VALUES (btrim(p_reason_name))
    RETURNING cancellation_reason.reason_id, cancellation_reason.reason_name
    INTO reason_id, reason_name;

    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_cancellation_reason(
    p_reason_id INT
)
RETURNS TABLE (
    reason_id INT,
    reason_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT c.reason_id, c.reason_name
    FROM cancellation_reason c
    WHERE c.reason_id = p_reason_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[cancellation-reason-not-found] cancellation_reason with reason_id=% does not exist', p_reason_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_cancellation_reasons(
    p_search TEXT DEFAULT NULL,
    p_limit INT DEFAULT 50,
    p_offset INT DEFAULT 0
)
RETURNS TABLE (
    total_records BIGINT,
    reason_id INT,
    reason_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH filtered AS (
        SELECT c.reason_id, c.reason_name
        FROM cancellation_reason c
        WHERE p_search IS NULL
           OR c.reason_name ILIKE '%' || p_search || '%'
    )
    SELECT
        count(*) OVER() AS total_records,
        f.reason_id,
        f.reason_name
    FROM filtered f
    ORDER BY f.reason_name
    LIMIT GREATEST(COALESCE(p_limit, 50), 0)
    OFFSET GREATEST(COALESCE(p_offset, 0), 0);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_update_cancellation_reason(
    p_reason_id INT,
    p_reason_name TEXT
)
RETURNS VOID AS $$
BEGIN
    UPDATE cancellation_reason
       SET reason_name = COALESCE(NULLIF(btrim(p_reason_name), ''), reason_name)
     WHERE reason_id = p_reason_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[cancellation-reason-not-found] cancellation_reason with reason_id=% does not exist', p_reason_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_delete_cancellation_reason(
    p_reason_id INT
)
RETURNS VOID AS $$
BEGIN
    PERFORM 1
    FROM appt_cancellation ac
    WHERE ac.reason_id = p_reason_id;

    IF FOUND THEN
        RAISE EXCEPTION '[cancellation-reason-in-use] reason_id=% is referenced by appt_cancellation', p_reason_id;
    END IF;

    DELETE FROM cancellation_reason
    WHERE reason_id = p_reason_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[cancellation-reason-not-found] cancellation_reason with reason_id=% does not exist', p_reason_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- CONTACT_TYPE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_contact_type(
    p_contact_type_name TEXT
)
RETURNS TABLE (
    contact_type_id INT,
    contact_type_name TEXT
) AS $$
BEGIN
    IF p_contact_type_name IS NULL OR btrim(p_contact_type_name) = '' THEN
        RAISE EXCEPTION '[contact-type-invalid] contact_type_name is required';
    END IF;

    INSERT INTO contact_type (contact_type_name)
    VALUES (btrim(p_contact_type_name))
    RETURNING contact_type.contact_type_id, contact_type.contact_type_name
    INTO contact_type_id, contact_type_name;

    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_contact_type(
    p_contact_type_id INT
)
RETURNS TABLE (
    contact_type_id INT,
    contact_type_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT c.contact_type_id, c.contact_type_name
    FROM contact_type c
    WHERE c.contact_type_id = p_contact_type_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[contact-type-not-found] contact_type with contact_type_id=% does not exist', p_contact_type_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_get_contact_types(
    p_search TEXT DEFAULT NULL,
    p_limit INT DEFAULT 50,
    p_offset INT DEFAULT 0
)
RETURNS TABLE (
    total_records BIGINT,
    contact_type_id INT,
    contact_type_name TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH filtered AS (
        SELECT c.contact_type_id, c.contact_type_name
        FROM contact_type c
        WHERE p_search IS NULL
           OR c.contact_type_name ILIKE '%' || p_search || '%'
    )
    SELECT
        count(*) OVER() AS total_records,
        f.contact_type_id,
        f.contact_type_name
    FROM filtered f
    ORDER BY f.contact_type_name
    LIMIT GREATEST(COALESCE(p_limit, 50), 0)
    OFFSET GREATEST(COALESCE(p_offset, 0), 0);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_update_contact_type(
    p_contact_type_id INT,
    p_contact_type_name TEXT
)
RETURNS VOID AS $$
BEGIN
    UPDATE contact_type
       SET contact_type_name = COALESCE(NULLIF(btrim(p_contact_type_name), ''), contact_type_name)
     WHERE contact_type_id = p_contact_type_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[contact-type-not-found] contact_type with contact_type_id=% does not exist', p_contact_type_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_delete_contact_type(
    p_contact_type_id INT
)
RETURNS VOID AS $$
BEGIN
    PERFORM 1
    FROM person_contact pc
    WHERE pc.contact_type_id = p_contact_type_id;

    IF FOUND THEN
        RAISE EXCEPTION '[contact-type-in-use] contact_type_id=% is referenced by person_contact', p_contact_type_id;
    END IF;

    DELETE FROM contact_type
    WHERE contact_type_id = p_contact_type_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[contact-type-not-found] contact_type with contact_type_id=% does not exist', p_contact_type_id;
    END IF;
END;
$$ LANGUAGE plpgsql;
