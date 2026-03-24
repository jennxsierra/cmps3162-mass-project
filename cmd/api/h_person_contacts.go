package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// POST /v1/persons/:id/contacts
func (a *applicationDependencies) createPersonContactHandler(w http.ResponseWriter, r *http.Request) {
	personID, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		ContactTypeID int    `json:"contact_type_id"`
		ContactValue  string `json:"contact_value"`
		IsPrimary     bool   `json:"is_primary"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	contact := &data.PersonContact{
		PersonID:      int64(personID),
		ContactTypeID: input.ContactTypeID,
		ContactValue:  input.ContactValue,
		IsPrimary:     input.IsPrimary,
	}

	v := validator.New()
	data.ValidatePersonContact(v, contact)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.PersonContact.Insert(contact)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPersonNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrContactTypeNotFound):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"contact_type_id": "does not exist"})
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/persons/%d/contacts/%d", personID, contact.PersonContactID))

	err = a.writeJSON(w, http.StatusCreated, envelope{"person_contact": contact}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/persons/:id/contacts
func (a *applicationDependencies) listPersonContactsHandler(w http.ResponseWriter, r *http.Request) {
	personID, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	var filters data.Filters
	v := validator.New()

	queryParameters := r.URL.Query()
	filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 20, v)
	filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "person_contact_id")
	filters.SortSafeList = []string{
		"person_contact_id", "contact_type_id", "contact_type_name", "contact_value", "is_primary",
		"-person_contact_id", "-contact_type_id", "-contact_type_name", "-contact_value", "-is_primary",
	}

	if data.ValidateFilters(v, filters); !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	contacts, metadata, err := a.models.PersonContact.GetAllByPerson(int64(personID), filters)
	if err != nil {
		if errors.Is(err, data.ErrPersonNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"person_contacts": contacts, "@metadata": metadata}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// PATCH /v1/persons/:id/contacts/:contact_id
func (a *applicationDependencies) updatePersonContactHandler(w http.ResponseWriter, r *http.Request) {
	personID, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	contactID, err := a.readIntIDParam(r, "contact_id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	contact, err := a.models.PersonContact.GetByPerson(int64(personID), int64(contactID))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPersonNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		ContactTypeID *int    `json:"contact_type_id"`
		ContactValue  *string `json:"contact_value"`
		IsPrimary     *bool   `json:"is_primary"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.ContactTypeID != nil {
		contact.ContactTypeID = *input.ContactTypeID
	}
	if input.ContactValue != nil {
		contact.ContactValue = *input.ContactValue
	}
	if input.IsPrimary != nil {
		contact.IsPrimary = *input.IsPrimary
	}

	v := validator.New()
	data.ValidatePersonContact(v, contact)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.PersonContact.Update(contact)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPersonNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrContactTypeNotFound):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"contact_type_id": "does not exist"})
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"person_contact": contact}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// DELETE /v1/persons/:id/contacts/:contact_id
func (a *applicationDependencies) deletePersonContactHandler(w http.ResponseWriter, r *http.Request) {
	personID, err := a.readIntIDParam(r, "id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	contactID, err := a.readIntIDParam(r, "contact_id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.models.PersonContact.Delete(int64(personID), int64(contactID))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPersonNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"message": "person contact successfully deleted"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
