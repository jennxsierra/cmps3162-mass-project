package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

func (a *applicationDependencies) createAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		StartTime  string `json:"start_time"`
		EndTime    string `json:"end_time"`
		Reason     string `json:"reason"`
		PatientID  int64  `json:"patient_id"`
		ProviderID int64  `json:"provider_id"`
		CreatedBy  int64  `json:"created_by"`
		ApptTypeID int    `json:"appt_type_id"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"start_time": "must be a valid RFC3339 datetime"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"end_time": "must be a valid RFC3339 datetime"})
		return
	}

	appointment := &data.Appointment{
		StartTime:  startTime,
		EndTime:    endTime,
		Reason:     input.Reason,
		PatientID:  input.PatientID,
		ProviderID: input.ProviderID,
		CreatedBy:  input.CreatedBy,
		ApptTypeID: input.ApptTypeID,
	}

	v := validator.New()
	data.ValidateAppointment(v, appointment)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Appointment.Insert(appointment)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidAppointmentTimeRange):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"end_time": "must be greater than start_time"})
		case errors.Is(err, data.ErrProviderUnavailable):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"provider_id": "has overlapping appointment"})
		case errors.Is(err, data.ErrPatientUnavailable):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"patient_id": "has overlapping appointment"})
		case errors.Is(err, data.ErrRecordNotFound):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, "one or more related records do not exist")
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/appointments/%d", appointment.AppointmentID))

	err = a.writeJSON(w, http.StatusCreated, envelope{"appointment": appointment}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) showAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	appointmentID, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	appointment, err := a.models.Appointment.Get(int64(appointmentID))
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"appointment": appointment}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listAppointmentsHandler(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	v := validator.New()

	var filters data.Filters
	filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 20, v)
	filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "start_time")
	filters.SortSafeList = []string{
		"appointment_id", "start_time", "end_time", "created_at", "updated_at", "patient_id", "provider_id", "appt_type_id",
		"-appointment_id", "-start_time", "-end_time", "-created_at", "-updated_at", "-patient_id", "-provider_id", "-appt_type_id",
	}

	providerID := int64(a.getSingleIntegerParameter(queryParameters, "provider_id", 0, v))
	patientID := int64(a.getSingleIntegerParameter(queryParameters, "patient_id", 0, v))
	apptTypeID := a.getSingleIntegerParameter(queryParameters, "appt_type_id", 0, v)

	includeCancelled := true
	includeCancelledRaw := a.getSingleQueryParameter(queryParameters, "include_cancelled", "true")
	if includeCancelledRaw != "" {
		parsed, err := strconv.ParseBool(includeCancelledRaw)
		if err != nil {
			v.AddError("include_cancelled", "must be a boolean value")
		} else {
			includeCancelled = parsed
		}
	}

	var startFrom *time.Time
	startFromRaw := a.getSingleQueryParameter(queryParameters, "start_from", "")
	if startFromRaw != "" {
		parsed, err := time.Parse(time.RFC3339, startFromRaw)
		if err != nil {
			v.AddError("start_from", "must be a valid RFC3339 datetime")
		} else {
			startFrom = &parsed
		}
	}

	var startTo *time.Time
	startToRaw := a.getSingleQueryParameter(queryParameters, "start_to", "")
	if startToRaw != "" {
		parsed, err := time.Parse(time.RFC3339, startToRaw)
		if err != nil {
			v.AddError("start_to", "must be a valid RFC3339 datetime")
		} else {
			startTo = &parsed
		}
	}

	if startFrom != nil && startTo != nil && startTo.Before(*startFrom) {
		v.AddError("start_to", "must be greater than or equal to start_from")
	}

	if data.ValidateFilters(v, filters); !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	query := data.AppointmentQueryParams{
		ProviderID:       providerID,
		PatientID:        patientID,
		ApptTypeID:       apptTypeID,
		StartFrom:        startFrom,
		StartTo:          startTo,
		IncludeCancelled: includeCancelled,
	}

	appointments, metadata, err := a.models.Appointment.GetAll(query, filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"appointments": appointments, "@metadata": metadata}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) updateAppointmentHandler(w http.ResponseWriter, r *http.Request) {
	appointmentID, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	appointment, err := a.models.Appointment.Get(int64(appointmentID))
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		StartTime  *string `json:"start_time"`
		EndTime    *string `json:"end_time"`
		Reason     *string `json:"reason"`
		PatientID  *int64  `json:"patient_id"`
		ProviderID *int64  `json:"provider_id"`
		CreatedBy  *int64  `json:"created_by"`
		ApptTypeID *int    `json:"appt_type_id"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.StartTime != nil {
		parsed, parseErr := time.Parse(time.RFC3339, *input.StartTime)
		if parseErr != nil {
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"start_time": "must be a valid RFC3339 datetime"})
			return
		}
		appointment.StartTime = parsed
	}

	if input.EndTime != nil {
		parsed, parseErr := time.Parse(time.RFC3339, *input.EndTime)
		if parseErr != nil {
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"end_time": "must be a valid RFC3339 datetime"})
			return
		}
		appointment.EndTime = parsed
	}

	if input.Reason != nil {
		appointment.Reason = *input.Reason
	}
	if input.PatientID != nil {
		appointment.PatientID = *input.PatientID
	}
	if input.ProviderID != nil {
		appointment.ProviderID = *input.ProviderID
	}
	if input.CreatedBy != nil {
		appointment.CreatedBy = *input.CreatedBy
	}
	if input.ApptTypeID != nil {
		appointment.ApptTypeID = *input.ApptTypeID
	}

	v := validator.New()
	data.ValidateAppointment(v, appointment)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Appointment.Update(appointment)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrInvalidAppointmentTimeRange):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"end_time": "must be greater than start_time"})
		case errors.Is(err, data.ErrProviderUnavailable):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"provider_id": "has overlapping appointment"})
		case errors.Is(err, data.ErrPatientUnavailable):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"patient_id": "has overlapping appointment"})
		case errors.Is(err, data.ErrCancelledAppointmentScheduleChange):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, "cannot change schedule fields for cancelled appointment")
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	appointment, err = a.models.Appointment.Get(int64(appointmentID))
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"appointment": appointment}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
