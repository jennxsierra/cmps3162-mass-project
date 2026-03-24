package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// POST /v1/appointments/:id/cancellations
func (a *applicationDependencies) createAppointmentCancellationHandler(w http.ResponseWriter, r *http.Request) {
	appointmentID, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		Note        string `json:"note"`
		ReasonID    int    `json:"reason_id"`
		RecordedBy  int64  `json:"recorded_by"`
		CancelledAt string `json:"cancelled_at"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	cancellation := &data.AppointmentCancellation{
		AppointmentID: int64(appointmentID),
		Note:          input.Note,
		ReasonID:      input.ReasonID,
		RecordedBy:    input.RecordedBy,
	}

	if input.CancelledAt != "" {
		cancelledAt, parseErr := time.Parse(time.RFC3339, input.CancelledAt)
		if parseErr != nil {
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"cancelled_at": "must be a valid RFC3339 datetime"})
			return
		}
		cancellation.CancelledAt = cancelledAt
	}

	v := validator.New()
	data.ValidateAppointmentCancellation(v, cancellation)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.AppointmentCancellation.Insert(cancellation)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrAppointmentAlreadyCancelled):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, "appointment is already cancelled")
		case errors.Is(err, data.ErrCancellationReasonNotFound):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"reason_id": "does not exist"})
		case errors.Is(err, data.ErrStaffNotFound):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"recorded_by": "does not exist"})
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/appointments/%d/cancellations", appointmentID))

	err = a.writeJSON(w, http.StatusCreated, envelope{"appointment_cancellation": cancellation}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/appointments/:id/cancellations
func (a *applicationDependencies) showAppointmentCancellationHandler(w http.ResponseWriter, r *http.Request) {
	appointmentID, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	cancellation, err := a.models.AppointmentCancellation.GetByAppointment(int64(appointmentID))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrAppointmentCancellationNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"appointment_cancellation": cancellation}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
