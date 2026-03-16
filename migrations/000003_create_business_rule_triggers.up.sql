-- migrations/000003_create_business_rule_triggers.up.sql
-- Creates business-rule trigger functions and triggers for MASS schema.

-- ====================================================================================
-- APPOINTMENT: updated_at management
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_appointment_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at := NOW();
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_appointment_set_updated_at
BEFORE UPDATE ON appointment
FOR EACH ROW
EXECUTE FUNCTION fn_appointment_set_updated_at();

-- ====================================================================================
-- APPOINTMENT: prevent overlapping active appointments for same provider
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_appointment_prevent_overlap_provider()
RETURNS TRIGGER AS $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM appointment a
		LEFT JOIN appt_cancellation ac
			   ON ac.appointment_id = a.appointment_id
		WHERE a.provider_id = NEW.provider_id
		  AND a.appointment_id <> COALESCE(NEW.appointment_id, -1)
		  AND ac.appointment_id IS NULL
		  AND tstzrange(a.start_time, a.end_time, '[)')
			  && tstzrange(NEW.start_time, NEW.end_time, '[)')
	) THEN
		RAISE EXCEPTION 'Provider % has an overlapping appointment in [% - %).',
			NEW.provider_id, NEW.start_time, NEW.end_time
			USING ERRCODE = '23514';
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_appointment_prevent_overlap_provider
BEFORE INSERT OR UPDATE ON appointment
FOR EACH ROW
EXECUTE FUNCTION fn_appointment_prevent_overlap_provider();

-- ====================================================================================
-- APPOINTMENT: prevent overlapping active appointments for same patient
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_appointment_prevent_overlap_patient()
RETURNS TRIGGER AS $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM appointment a
		LEFT JOIN appt_cancellation ac
			   ON ac.appointment_id = a.appointment_id
		WHERE a.patient_id = NEW.patient_id
		  AND a.appointment_id <> COALESCE(NEW.appointment_id, -1)
		  AND ac.appointment_id IS NULL
		  AND tstzrange(a.start_time, a.end_time, '[)')
			  && tstzrange(NEW.start_time, NEW.end_time, '[)')
	) THEN
		RAISE EXCEPTION 'Patient % has an overlapping appointment in [% - %).',
			NEW.patient_id, NEW.start_time, NEW.end_time
			USING ERRCODE = '23514';
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_appointment_prevent_overlap_patient
BEFORE INSERT OR UPDATE ON appointment
FOR EACH ROW
EXECUTE FUNCTION fn_appointment_prevent_overlap_patient();

-- ====================================================================================
-- PERSON CONTACT: enforce one primary contact per person
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_person_contact_single_primary()
RETURNS TRIGGER AS $$
BEGIN
	IF NEW.is_primary THEN
		IF EXISTS (
			SELECT 1
			FROM person_contact pc
			WHERE pc.person_id = NEW.person_id
			  AND pc.is_primary = TRUE
			  AND pc.person_contact_id <> COALESCE(NEW.person_contact_id, -1)
		) THEN
			RAISE EXCEPTION 'Person % already has a primary contact.', NEW.person_id
				USING ERRCODE = '23514';
		END IF;
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_person_contact_single_primary
BEFORE INSERT OR UPDATE ON person_contact
FOR EACH ROW
EXECUTE FUNCTION fn_person_contact_single_primary();

-- ====================================================================================
-- PERSON CONTACT: normalize contact value
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_person_contact_normalize_value()
RETURNS TRIGGER AS $$
DECLARE
	v_contact_type_name TEXT;
BEGIN
	NEW.contact_value := btrim(NEW.contact_value);

	SELECT ct.contact_type_name
	  INTO v_contact_type_name
	  FROM contact_type ct
	 WHERE ct.contact_type_id = NEW.contact_type_id;

	IF v_contact_type_name = 'email' THEN
		NEW.contact_value := lower(NEW.contact_value);
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_person_contact_normalize_value
BEFORE INSERT OR UPDATE ON person_contact
FOR EACH ROW
EXECUTE FUNCTION fn_person_contact_normalize_value();

-- ====================================================================================
-- CANCELLATION: validate cancellation timestamp against appointment window
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_appt_cancellation_validate_window()
RETURNS TRIGGER AS $$
DECLARE
	v_end_time TIMESTAMPTZ;
BEGIN
	IF NEW.cancelled_at IS NULL THEN
		NEW.cancelled_at := NOW();
	END IF;

	SELECT a.end_time
	  INTO v_end_time
	  FROM appointment a
	 WHERE a.appointment_id = NEW.appointment_id;

	IF NOT FOUND THEN
		RAISE EXCEPTION 'Appointment % does not exist.', NEW.appointment_id
			USING ERRCODE = '23503';
	END IF;

	IF NEW.cancelled_at > v_end_time THEN
		RAISE EXCEPTION 'Cannot cancel appointment % after it has ended.', NEW.appointment_id
			USING ERRCODE = '23514';
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_appt_cancellation_validate_window
BEFORE INSERT ON appt_cancellation
FOR EACH ROW
EXECUTE FUNCTION fn_appt_cancellation_validate_window();

-- ====================================================================================
-- APPOINTMENT: block scheduling edits for cancelled appointments
-- ====================================================================================

CREATE OR REPLACE FUNCTION fn_appointment_block_edit_if_cancelled()
RETURNS TRIGGER AS $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM appt_cancellation ac
		WHERE ac.appointment_id = NEW.appointment_id
	) THEN
		IF NEW.start_time   IS DISTINCT FROM OLD.start_time
		OR NEW.end_time     IS DISTINCT FROM OLD.end_time
		OR NEW.patient_id   IS DISTINCT FROM OLD.patient_id
		OR NEW.provider_id  IS DISTINCT FROM OLD.provider_id
		OR NEW.appt_type_id IS DISTINCT FROM OLD.appt_type_id THEN
			RAISE EXCEPTION 'Cannot change schedule fields for cancelled appointment %.', NEW.appointment_id
				USING ERRCODE = '23514';
		END IF;
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_appointment_block_edit_if_cancelled
BEFORE UPDATE ON appointment
FOR EACH ROW
EXECUTE FUNCTION fn_appointment_block_edit_if_cancelled();
