package main

import (
	"errors"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// GET /v1/cancellation-reasons
func (a *applicationDependencies) listCancellationReasonsHandler(w http.ResponseWriter, r *http.Request) {
	var filters data.Filters

	v := validator.New()
	queryParams := r.URL.Query()

	filters.Page = a.getSingleIntegerParameter(queryParams, "page", 1, v)
	filters.PageSize = a.getSingleIntegerParameter(queryParams, "page_size", 20, v)
	filters.Sort = a.getSingleQueryParameter(queryParams, "sort", "reason_id")
	filters.SortSafeList = []string{"reason_id", "reason_name", "-reason_id", "-reason_name"}

	if data.ValidateFilters(v, filters); !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	reasons, metadata, err := a.models.CancellationReason.GetAll(filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"cancellation_reasons": reasons, "@metadata": metadata}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/cancellation-reasons/:id
func (a *applicationDependencies) showCancellationReasonHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	reason, err := a.models.CancellationReason.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
			return
		}
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"cancellation_reason": reason}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
