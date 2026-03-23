package main

import (
	"errors"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// GET /v1/contact-types
func (a *applicationDependencies) listContactTypesHandler(w http.ResponseWriter, r *http.Request) {
	var filters data.Filters

	v := validator.New()
	queryParams := r.URL.Query()

	filters.Page = a.getSingleIntegerParameter(queryParams, "page", 1, v)
	filters.PageSize = a.getSingleIntegerParameter(queryParams, "page_size", 20, v)
	filters.Sort = a.getSingleQueryParameter(queryParams, "sort", "contact_type_id")
	filters.SortSafeList = []string{"contact_type_id", "contact_type_name", "-contact_type_id", "-contact_type_name"}

	if data.ValidateFilters(v, filters); !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	contactTypes, metadata, err := a.models.ContactType.GetAll(filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"contact_types": contactTypes, "@metadata": metadata}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/contact-types/:id
func (a *applicationDependencies) showContactTypeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	contactType, err := a.models.ContactType.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"contact_type": contactType}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
