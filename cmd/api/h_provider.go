package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) readProviderLicenseNoParam(r *http.Request) (string, error) {
	params := httprouter.ParamsFromContext(r.Context())
	licenseNo := params.ByName("license_no")
	if licenseNo == "" {
		return "", errors.New("invalid license_no")
	}
	return licenseNo, nil
}

func (a *applicationDependencies) createProviderHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		LicenseNo   string `json:"license_no"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DateOfBirth string `json:"date_of_birth"`
		Gender      string `json:"gender"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	provider := &data.Provider{
		LicenseNo:   input.LicenseNo,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		DateOfBirth: input.DateOfBirth,
		Gender:      input.Gender,
	}

	v := validator.New()
	data.ValidateProvider(v, provider)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Provider.Insert(provider)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/providers/%s", provider.LicenseNo))

	err = a.writeJSON(w, http.StatusCreated, envelope{"provider": provider}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) showProviderHandler(w http.ResponseWriter, r *http.Request) {
	licenseNo, err := a.readProviderLicenseNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	provider, err := a.models.Provider.Get(licenseNo)
	if err != nil {
		if err == data.ErrRecordNotFound {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"provider": provider}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listProvidersHandler(w http.ResponseWriter, r *http.Request) {
	var queryParametersData struct {
		Search      string
		SpecialtyID int
		data.Filters
	}

	queryParameters := r.URL.Query()
	queryParametersData.Search = a.getSingleQueryParameter(queryParameters, "search", "")

	v := validator.New()
	queryParametersData.SpecialtyID = a.getSingleIntegerParameter(queryParameters, "specialty_id", 0, v)
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "provider_id")
	queryParametersData.Filters.SortSafeList = []string{
		"provider_id", "license_no", "first_name", "last_name", "created_at",
		"-provider_id", "-license_no", "-first_name", "-last_name", "-created_at",
	}

	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	providers, metadata, err := a.models.Provider.GetAll(
		queryParametersData.Search,
		queryParametersData.SpecialtyID,
		queryParametersData.Filters,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	resp := envelope{
		"providers": providers,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) updateProviderHandler(w http.ResponseWriter, r *http.Request) {
	licenseNo, err := a.readProviderLicenseNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	provider, err := a.models.Provider.Get(licenseNo)
	if err != nil {
		if err == data.ErrRecordNotFound {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		LicenseNo   *string `json:"license_no"`
		FirstName   *string `json:"first_name"`
		LastName    *string `json:"last_name"`
		DateOfBirth *string `json:"date_of_birth"`
		Gender      *string `json:"gender"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.LicenseNo != nil {
		provider.LicenseNo = *input.LicenseNo
	}
	if input.FirstName != nil {
		provider.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		provider.LastName = *input.LastName
	}
	if input.DateOfBirth != nil {
		provider.DateOfBirth = *input.DateOfBirth
	}
	if input.Gender != nil {
		provider.Gender = *input.Gender
	}

	v := validator.New()
	data.ValidateProvider(v, provider)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Provider.Update(provider)
	if err != nil {
		if err == data.ErrRecordNotFound {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"provider": provider}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteProviderHandler(w http.ResponseWriter, r *http.Request) {
	licenseNo, err := a.readProviderLicenseNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.models.Provider.Delete(licenseNo)
	if err != nil {
		switch err {
		case data.ErrRecordNotFound:
			a.notFoundResponse(w, r)
		case data.ErrProviderInUse:
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, "provider has related appointments and cannot be deleted")
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"message": "provider successfully deleted"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
