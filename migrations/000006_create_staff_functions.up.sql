-- migrations/000006_create_staff_functions.up.sql
-- Creates staff CRUD + search functions.

-- ====================================================================================
-- STAFF: CREATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_staff(
    p_first_name TEXT,
    p_last_name TEXT,
    p_date_of_birth DATE DEFAULT NULL,
    p_gender TEXT DEFAULT NULL,
    p_staff_no TEXT DEFAULT NULL
)
RETURNS TABLE (
    staff_id BIGINT,
    staff_no TEXT,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    gender TEXT,
    created_at TIMESTAMPTZ
) AS $$
DECLARE
    v_staff_id BIGINT;
BEGIN
    IF p_first_name IS NULL OR btrim(p_first_name) = '' THEN
        RAISE EXCEPTION '[staff-invalid] first_name is required';
    END IF;

    IF p_last_name IS NULL OR btrim(p_last_name) = '' THEN
        RAISE EXCEPTION '[staff-invalid] last_name is required';
    END IF;

    IF p_staff_no IS NULL OR btrim(p_staff_no) = '' THEN
        RAISE EXCEPTION '[staff-invalid] staff_no is required';
    END IF;

    INSERT INTO person (first_name, last_name, date_of_birth, gender)
    VALUES (btrim(p_first_name), btrim(p_last_name), p_date_of_birth, p_gender)
    RETURNING person_id INTO v_staff_id;

    INSERT INTO staff (staff_id, staff_no)
    VALUES (v_staff_id, btrim(p_staff_no));

    RETURN QUERY
    SELECT p.person_id, s.staff_no, p.first_name, p.last_name, p.date_of_birth, p.gender, p.created_at
    FROM person p
    JOIN staff s ON s.staff_id = p.person_id
    WHERE p.person_id = v_staff_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- STAFF: READ (BY STAFF NO)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_staff(
    p_staff_no TEXT
)
RETURNS TABLE (
    staff_id BIGINT,
    staff_no TEXT,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    gender TEXT,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT p.person_id, s.staff_no, p.first_name, p.last_name, p.date_of_birth, p.gender, p.created_at
    FROM staff s
    JOIN person p ON p.person_id = s.staff_id
    WHERE s.staff_no = p_staff_no;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[staff-not-found] staff with staff_no=% does not exist', p_staff_no;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- STAFF: LIST / SEARCH
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_staffs(
    p_search TEXT DEFAULT NULL,
    p_limit INT DEFAULT 50,
    p_offset INT DEFAULT 0
)
RETURNS TABLE (
    total_records BIGINT,
    staff_id BIGINT,
    staff_no TEXT,
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
            s.staff_no,
            p.first_name,
            p.last_name,
            p.date_of_birth,
            p.gender,
            p.created_at
        FROM staff s
        JOIN person p ON p.person_id = s.staff_id
        WHERE (
            p_search IS NULL
            OR p.first_name ILIKE '%' || p_search || '%'
            OR p.last_name ILIKE '%' || p_search || '%'
            OR (p.first_name || ' ' || p.last_name) ILIKE '%' || p_search || '%'
            OR s.staff_no ILIKE '%' || p_search || '%'
        )
    )
    SELECT
        count(*) OVER() AS total_records,
        f.person_id,
        f.staff_no,
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
-- STAFF: UPDATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_update_staff(
    p_staff_no TEXT,
    p_first_name TEXT DEFAULT NULL,
    p_last_name TEXT DEFAULT NULL,
    p_date_of_birth DATE DEFAULT NULL,
    p_gender TEXT DEFAULT NULL,
    p_new_staff_no TEXT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_staff_id BIGINT;
BEGIN
    SELECT s.staff_id
      INTO v_staff_id
      FROM staff s
     WHERE s.staff_no = p_staff_no
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[staff-not-found] staff with staff_no=% does not exist', p_staff_no;
    END IF;

    UPDATE person
       SET first_name = COALESCE(p_first_name, first_name),
           last_name = COALESCE(p_last_name, last_name),
           date_of_birth = COALESCE(p_date_of_birth, date_of_birth),
           gender = COALESCE(p_gender, gender)
     WHERE person_id = v_staff_id;

    UPDATE staff
       SET staff_no = COALESCE(NULLIF(btrim(p_new_staff_no), ''), staff_no)
     WHERE staff_id = v_staff_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- STAFF: DELETE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_delete_staff(
    p_staff_no TEXT
)
RETURNS VOID AS $$
DECLARE
    v_staff_id BIGINT;
BEGIN
    SELECT s.staff_id
      INTO v_staff_id
      FROM staff s
     WHERE s.staff_no = p_staff_no
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[staff-not-found] staff with staff_no=% does not exist', p_staff_no;
    END IF;

    DELETE FROM person
     WHERE person_id = v_staff_id;
END;
$$ LANGUAGE plpgsql;
