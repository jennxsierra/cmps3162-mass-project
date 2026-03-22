package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// POST /v1/appointment-types -- create new appointment type
func (a *applicationDependencies) createAppointmentTypeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AppointmentName string `json:"appointment_name"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	appointmentType := &data.AppointmentType{
		AppointmentName: input.AppointmentName,
	}

	v := validator.New()
	data.ValidateAppointmentType(v, appointmentType)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.AppointmentType.Insert(appointmentType)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/appointment-types/%d", appointmentType.AppointmentTypeID))

	err = a.writeJSON(w, http.StatusCreated, envelope{"appointment_type": appointmentType}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/appointment-types -- list all appointment types
func (a *applicationDependencies) listAppointmentTypesHandler(w http.ResponseWriter, r *http.Request) {
	var filters data.Filters

	v := validator.New()

	queryParams := r.URL.Query()
	filters.Page = a.getSingleIntegerParameter(queryParams, "page", 1, v)
	filters.PageSize = a.getSingleIntegerParameter(queryParams, "page_size", 20, v)
	filters.Sort = a.getSingleQueryParameter(queryParams, "sort", "appointment_type_id")
	filters.SortSafeList = []string{"appointment_type_id", "appointment_name", "-appointment_type_id", "-appointment_name"}

	if data.ValidateFilters(v, filters); !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	appointmentTypes, metadata, err := a.models.AppointmentType.GetAll(filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"appointment_types": appointmentTypes, "@metadata": metadata}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/appointment-types/:id -- show appointment type by id
func (a *applicationDependencies) showAppointmentTypeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	appointmentType, err := a.models.AppointmentType.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"appointment_type": appointmentType}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// PATCH /v1/appointment-types/:id -- update appointment type
func (a *applicationDependencies) updateAppointmentTypeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	appointmentType, err := a.models.AppointmentType.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}

	var input struct {
		AppointmentName *string `json:"appointment_name"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.AppointmentName != nil {
		appointmentType.AppointmentName = *input.AppointmentName
	}

	v := validator.New()
	data.ValidateAppointmentType(v, appointmentType)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.AppointmentType.Update(appointmentType)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"appointment_type": appointmentType}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
