-- migrations/000010_create_person_contact_functions.up.sql
-- Creates person_contact CRUD + list functions and primary-contact uniqueness index.

-- ====================================================================================
-- PERSON CONTACT: PRIMARY-CONTACT CONSTRAINT HELPER
-- ====================================================================================

CREATE UNIQUE INDEX IF NOT EXISTS ux_person_contact_primary_per_type
ON person_contact (person_id, contact_type_id)
WHERE is_primary = TRUE;

-- ====================================================================================
-- PERSON CONTACT: CREATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_person_contact(
    p_person_id BIGINT,
    p_contact_type_id INT,
    p_contact_value TEXT,
    p_is_primary BOOLEAN DEFAULT FALSE
)
RETURNS TABLE (
    person_contact_id BIGINT,
    person_id BIGINT,
    contact_type_id INT,
    contact_value TEXT,
    is_primary BOOLEAN
) AS $$
DECLARE
    v_person_contact_id BIGINT;
    v_person_id BIGINT;
    v_contact_type_id INT;
    v_contact_value TEXT;
    v_is_primary BOOLEAN;
BEGIN
    IF p_contact_value IS NULL OR btrim(p_contact_value) = '' THEN
        RAISE EXCEPTION '[person-contact-invalid] contact_value is required';
    END IF;

    PERFORM 1 FROM person p WHERE p.person_id = p_person_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[person-not-found] person with person_id=% does not exist', p_person_id;
    END IF;

    PERFORM 1 FROM contact_type ct WHERE ct.contact_type_id = p_contact_type_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[contact-type-not-found] contact_type with contact_type_id=% does not exist', p_contact_type_id;
    END IF;

    IF COALESCE(p_is_primary, FALSE) THEN
        UPDATE person_contact
           SET is_primary = FALSE
         WHERE person_id = p_person_id
           AND is_primary = TRUE;
    END IF;

    INSERT INTO person_contact (
        contact_value,
        is_primary,
        person_id,
        contact_type_id
    )
    VALUES (
        btrim(p_contact_value),
        COALESCE(p_is_primary, FALSE),
        p_person_id,
        p_contact_type_id
    )
    RETURNING
        person_contact.person_contact_id,
        person_contact.person_id,
        person_contact.contact_type_id,
        person_contact.contact_value,
        person_contact.is_primary
    INTO
        v_person_contact_id,
        v_person_id,
        v_contact_type_id,
        v_contact_value,
        v_is_primary;

    RETURN QUERY
    SELECT
        v_person_contact_id,
        v_person_id,
        v_contact_type_id,
        v_contact_value,
        v_is_primary;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PERSON CONTACT: READ (BY PERSON CONTACT ID)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_person_contact(
    p_person_contact_id BIGINT
)
RETURNS TABLE (
    person_contact_id BIGINT,
    person_id BIGINT,
    contact_type_id INT,
    contact_type_name TEXT,
    contact_value TEXT,
    is_primary BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        pc.person_contact_id,
        pc.person_id,
        pc.contact_type_id,
        ct.contact_type_name,
        pc.contact_value,
        pc.is_primary
    FROM person_contact pc
    JOIN contact_type ct ON ct.contact_type_id = pc.contact_type_id
    WHERE pc.person_contact_id = p_person_contact_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[person-contact-not-found] person_contact with person_contact_id=% does not exist', p_person_contact_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PERSON CONTACT: LIST
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_person_contacts(
    p_person_id BIGINT,
    p_contact_type_id INT DEFAULT NULL,
    p_is_primary BOOLEAN DEFAULT NULL
)
RETURNS TABLE (
    total_records BIGINT,
    person_contact_id BIGINT,
    person_id BIGINT,
    contact_type_id INT,
    contact_type_name TEXT,
    contact_value TEXT,
    is_primary BOOLEAN
) AS $$
BEGIN
    PERFORM 1 FROM person p WHERE p.person_id = p_person_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION '[person-not-found] person with person_id=% does not exist', p_person_id;
    END IF;

    RETURN QUERY
    WITH filtered AS (
        SELECT
            pc.person_contact_id,
            pc.person_id,
            pc.contact_type_id,
            ct.contact_type_name,
            pc.contact_value,
            pc.is_primary
        FROM person_contact pc
        JOIN contact_type ct ON ct.contact_type_id = pc.contact_type_id
        WHERE pc.person_id = p_person_id
          AND (p_contact_type_id IS NULL OR pc.contact_type_id = p_contact_type_id)
          AND (p_is_primary IS NULL OR pc.is_primary = p_is_primary)
    )
    SELECT
        count(*) OVER() AS total_records,
        f.person_contact_id,
        f.person_id,
        f.contact_type_id,
        f.contact_type_name,
        f.contact_value,
        f.is_primary
    FROM filtered f
    ORDER BY f.is_primary DESC, f.person_contact_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PERSON CONTACT: UPDATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_update_person_contact(
    p_person_contact_id BIGINT,
    p_contact_value TEXT DEFAULT NULL,
    p_is_primary BOOLEAN DEFAULT NULL,
    p_contact_type_id INT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_person_id BIGINT;
    v_set_primary BOOLEAN;
BEGIN
    SELECT pc.person_id
      INTO v_person_id
      FROM person_contact pc
     WHERE pc.person_contact_id = p_person_contact_id
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[person-contact-not-found] person_contact with person_contact_id=% does not exist', p_person_contact_id;
    END IF;

    IF p_contact_type_id IS NOT NULL THEN
        PERFORM 1 FROM contact_type ct WHERE ct.contact_type_id = p_contact_type_id;
        IF NOT FOUND THEN
            RAISE EXCEPTION '[contact-type-not-found] contact_type with contact_type_id=% does not exist', p_contact_type_id;
        END IF;
    END IF;

    v_set_primary := COALESCE(p_is_primary, FALSE);
    IF v_set_primary THEN
        UPDATE person_contact
           SET is_primary = FALSE
         WHERE person_id = v_person_id
           AND person_contact_id <> p_person_contact_id
           AND is_primary = TRUE;
    END IF;

    UPDATE person_contact
       SET contact_value = COALESCE(NULLIF(btrim(p_contact_value), ''), contact_value),
           is_primary = COALESCE(p_is_primary, is_primary),
           contact_type_id = COALESCE(p_contact_type_id, contact_type_id)
     WHERE person_contact_id = p_person_contact_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PERSON CONTACT: DELETE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_delete_person_contact(
    p_person_contact_id BIGINT
)
RETURNS VOID AS $$
BEGIN
    DELETE FROM person_contact
     WHERE person_contact_id = p_person_contact_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[person-contact-not-found] person_contact with person_contact_id=% does not exist', p_person_contact_id;
    END IF;
END;
$$ LANGUAGE plpgsql;
