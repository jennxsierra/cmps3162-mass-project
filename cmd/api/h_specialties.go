package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// POST /v1/specialties -- create new specialty
func (a *applicationDependencies) createSpecialtyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SpecialtyName string `json:"specialty_name"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	specialty := &data.Specialty{
		SpecialtyName: input.SpecialtyName,
	}

	v := validator.New()
	data.ValidateSpecialty(v, specialty)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Specialty.Insert(specialty)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/specialties/%d", specialty.SpecialtyID))

	err = a.writeJSON(w, http.StatusCreated, envelope{"specialty": specialty}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/specialties -- list all specialties
func (a *applicationDependencies) listSpecialtiesHandler(w http.ResponseWriter, r *http.Request) {
	var filters data.Filters

	v := validator.New()

	queryParams := r.URL.Query()
	filters.Page = a.getSingleIntegerParameter(queryParams, "page", 1, v)
	filters.PageSize = a.getSingleIntegerParameter(queryParams, "page_size", 20, v)
	filters.Sort = a.getSingleQueryParameter(queryParams, "sort", "specialty_id")
	filters.SortSafeList = []string{"specialty_id", "specialty_name", "-specialty_id", "-specialty_name"}

	if data.ValidateFilters(v, filters); !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	specialties, metadata, err := a.models.Specialty.GetAll(filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"specialties": specialties, "@metadata": metadata}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/specialties/:id -- show specialty by id
func (a *applicationDependencies) showSpecialtyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	specialty, err := a.models.Specialty.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"specialty": specialty}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// PATCH /v1/specialties/:id -- update specialty
func (a *applicationDependencies) updateSpecialtyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	specialty, err := a.models.Specialty.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}

	var input struct {
		SpecialtyName *string `json:"specialty_name"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.SpecialtyName != nil {
		specialty.SpecialtyName = *input.SpecialtyName
	}

	v := validator.New()
	data.ValidateSpecialty(v, specialty)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Specialty.Update(specialty)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"specialty": specialty}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
