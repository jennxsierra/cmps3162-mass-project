package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// POST /v1/providers/:license_no/specialties
func (a *applicationDependencies) createProviderSpecialtyHandler(w http.ResponseWriter, r *http.Request) {
	licenseNo, err := a.readProviderLicenseNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	var input struct {
		SpecialtyID int `json:"specialty_id"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.SpecialtyID > 0, "specialty_id", "must be provided")
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	providerSpecialty, err := a.models.ProviderSpecialty.Insert(licenseNo, input.SpecialtyID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrProviderNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrSpecialtyNotFound):
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, envelope{"specialty_id": "does not exist"})
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/providers/%s/specialties/%d", licenseNo, providerSpecialty.SpecialtyID))

	err = a.writeJSON(w, http.StatusCreated, envelope{"provider_specialty": providerSpecialty}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/providers/:license_no/specialties
func (a *applicationDependencies) listProviderSpecialtiesHandler(w http.ResponseWriter, r *http.Request) {
	licenseNo, err := a.readProviderLicenseNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	specialties, err := a.models.ProviderSpecialty.GetAllByProvider(licenseNo)
	if err != nil {
		if errors.Is(err, data.ErrProviderNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"specialties": specialties}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// DELETE /v1/providers/:license_no/specialties/:specialty_id
func (a *applicationDependencies) deleteProviderSpecialtyHandler(w http.ResponseWriter, r *http.Request) {
	licenseNo, err := a.readProviderLicenseNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	specialtyID, err := a.readIntIDParam(r, "specialty_id")
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.models.ProviderSpecialty.Delete(licenseNo, specialtyID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrProviderNotFound):
			a.notFoundResponse(w, r)
		case errors.Is(err, data.ErrProviderSpecialtyNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"message": "provider specialty successfully removed"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
