-- migrations/000005_create_provider_functions.up.sql
-- Creates provider CRUD + specialty management functions.

-- ====================================================================================
-- PROVIDER: CREATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_provider(
    p_first_name TEXT,
    p_last_name TEXT,
    p_date_of_birth DATE DEFAULT NULL,
    p_gender TEXT DEFAULT NULL,
    p_license_no TEXT DEFAULT NULL
)
RETURNS TABLE (
    provider_id BIGINT,
    license_no TEXT,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    gender TEXT,
    created_at TIMESTAMPTZ
) AS $$
DECLARE
    v_provider_id BIGINT;
BEGIN
    IF p_first_name IS NULL OR btrim(p_first_name) = '' THEN
        RAISE EXCEPTION '[provider-invalid] first_name is required';
    END IF;

    IF p_last_name IS NULL OR btrim(p_last_name) = '' THEN
        RAISE EXCEPTION '[provider-invalid] last_name is required';
    END IF;

    IF p_license_no IS NULL OR btrim(p_license_no) = '' THEN
        RAISE EXCEPTION '[provider-invalid] license_no is required';
    END IF;

    INSERT INTO person (first_name, last_name, date_of_birth, gender)
    VALUES (btrim(p_first_name), btrim(p_last_name), p_date_of_birth, p_gender)
    RETURNING person_id INTO v_provider_id;

    INSERT INTO provider (provider_id, license_no)
    VALUES (v_provider_id, btrim(p_license_no));

    RETURN QUERY
    SELECT p.person_id, pr.license_no, p.first_name, p.last_name, p.date_of_birth, p.gender, p.created_at
    FROM person p
    JOIN provider pr ON pr.provider_id = p.person_id
    WHERE p.person_id = v_provider_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PROVIDER: READ (BY LICENSE NO)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_provider(
    p_license_no TEXT
)
RETURNS TABLE (
    provider_id BIGINT,
    license_no TEXT,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    gender TEXT,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT p.person_id, pr.license_no, p.first_name, p.last_name, p.date_of_birth, p.gender, p.created_at
    FROM provider pr
    JOIN person p ON p.person_id = pr.provider_id
    WHERE pr.license_no = p_license_no;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-not-found] provider with license_no=% does not exist', p_license_no;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PROVIDER: LIST / SEARCH
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_providers(
    p_search TEXT DEFAULT NULL,
    p_specialty_id INT DEFAULT NULL,
    p_limit INT DEFAULT 50,
    p_offset INT DEFAULT 0
)
RETURNS TABLE (
    total_records BIGINT,
    provider_id BIGINT,
    license_no TEXT,
    first_name TEXT,
    last_name TEXT,
    date_of_birth DATE,
    gender TEXT,
    created_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    WITH filtered AS (
        SELECT DISTINCT
            p.person_id,
            pr.license_no,
            p.first_name,
            p.last_name,
            p.date_of_birth,
            p.gender,
            p.created_at
        FROM provider pr
        JOIN person p ON p.person_id = pr.provider_id
        LEFT JOIN provider_specialty ps ON ps.provider_id = pr.provider_id
        WHERE (
            p_search IS NULL
            OR p.first_name ILIKE '%' || p_search || '%'
            OR p.last_name ILIKE '%' || p_search || '%'
            OR (p.first_name || ' ' || p.last_name) ILIKE '%' || p_search || '%'
            OR pr.license_no ILIKE '%' || p_search || '%'
        )
          AND (
            p_specialty_id IS NULL
            OR ps.specialty_id = p_specialty_id
        )
    )
    SELECT
        count(*) OVER() AS total_records,
        f.person_id,
        f.license_no,
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
-- PROVIDER: UPDATE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_update_provider(
    p_license_no TEXT,
    p_first_name TEXT DEFAULT NULL,
    p_last_name TEXT DEFAULT NULL,
    p_date_of_birth DATE DEFAULT NULL,
    p_gender TEXT DEFAULT NULL,
    p_new_license_no TEXT DEFAULT NULL
)
RETURNS VOID AS $$
DECLARE
    v_provider_id BIGINT;
BEGIN
    SELECT pr.provider_id
      INTO v_provider_id
      FROM provider pr
     WHERE pr.license_no = p_license_no
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-not-found] provider with license_no=% does not exist', p_license_no;
    END IF;

    UPDATE person
       SET first_name = COALESCE(p_first_name, first_name),
           last_name = COALESCE(p_last_name, last_name),
           date_of_birth = COALESCE(p_date_of_birth, date_of_birth),
           gender = COALESCE(p_gender, gender)
     WHERE person_id = v_provider_id;

    UPDATE provider
       SET license_no = COALESCE(NULLIF(btrim(p_new_license_no), ''), license_no)
     WHERE provider_id = v_provider_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PROVIDER: DELETE
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_delete_provider(
    p_license_no TEXT
)
RETURNS VOID AS $$
DECLARE
    v_provider_id BIGINT;
BEGIN
    SELECT pr.provider_id
      INTO v_provider_id
      FROM provider pr
     WHERE pr.license_no = p_license_no
     FOR UPDATE;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-not-found] provider with license_no=% does not exist', p_license_no;
    END IF;

    DELETE FROM person
     WHERE person_id = v_provider_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PROVIDER SPECIALTY: CREATE / ATTACH (IDEMPOTENT)
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_create_provider_specialty(
    p_license_no TEXT,
    p_specialty_id INT
)
RETURNS TABLE (
    provider_id BIGINT,
    specialty_id INT
) AS $$
DECLARE
    v_provider_id BIGINT;
BEGIN
    SELECT pr.provider_id
      INTO v_provider_id
      FROM provider pr
     WHERE pr.license_no = p_license_no;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-not-found] provider with license_no=% does not exist', p_license_no;
    END IF;

    PERFORM 1
      FROM specialty s
     WHERE s.specialty_id = p_specialty_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[specialty-not-found] specialty with specialty_id=% does not exist', p_specialty_id;
    END IF;

    INSERT INTO provider_specialty (provider_id, specialty_id)
    VALUES (v_provider_id, p_specialty_id)
    ON CONFLICT DO NOTHING;

    RETURN QUERY
    SELECT v_provider_id, p_specialty_id;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PROVIDER SPECIALTY: DELETE / DETACH
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_delete_provider_specialty(
    p_license_no TEXT,
    p_specialty_id INT
)
RETURNS VOID AS $$
DECLARE
    v_provider_id BIGINT;
BEGIN
    SELECT pr.provider_id
      INTO v_provider_id
      FROM provider pr
     WHERE pr.license_no = p_license_no;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-not-found] provider with license_no=% does not exist', p_license_no;
    END IF;

    DELETE FROM provider_specialty ps
     WHERE ps.provider_id = v_provider_id
       AND ps.specialty_id = p_specialty_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-specialty-not-found] provider specialty mapping does not exist (license_no=%, specialty_id=%)', p_license_no, p_specialty_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ====================================================================================
-- PROVIDER SPECIALTY: LIST
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_get_provider_specialties(
    p_license_no TEXT
)
RETURNS TABLE (
    specialty_id INT,
    specialty_name TEXT
) AS $$
DECLARE
    v_provider_id BIGINT;
BEGIN
    SELECT pr.provider_id
      INTO v_provider_id
      FROM provider pr
     WHERE pr.license_no = p_license_no;

    IF NOT FOUND THEN
        RAISE EXCEPTION '[provider-not-found] provider with license_no=% does not exist', p_license_no;
    END IF;

    RETURN QUERY
    SELECT s.specialty_id, s.specialty_name
    FROM provider_specialty ps
    JOIN specialty s ON s.specialty_id = ps.specialty_id
    WHERE ps.provider_id = v_provider_id
    ORDER BY s.specialty_name;
END;
$$ LANGUAGE plpgsql;
