package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jennxsierra/rate-limiting-shutdown/internal/data"
	"github.com/jennxsierra/rate-limiting-shutdown/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// Helper to extract :patient_no param
func (a *applicationDependencies) readPatientNoParam(r *http.Request) string {
	params := httprouter.ParamsFromContext(r.Context())
	return params.ByName("patient_no")
}

// POST /v1/patients -- create new patient
func (a *applicationDependencies) createPatientHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PatientNo   string `json:"patient_no"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DateOfBirth string `json:"date_of_birth"`
		Gender      string `json:"gender"`
		SSN         string `json:"ssn"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	patient := &data.Patient{
		PatientNo:   input.PatientNo,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		DateOfBirth: input.DateOfBirth,
		Gender:      input.Gender,
		SSN:         input.SSN,
	}

	v := validator.New()
	data.ValidatePatient(v, patient)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Patient.Insert(patient)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/patients/%s", patient.PatientNo))

	err = a.writeJSON(w, http.StatusCreated, envelope{"patient": patient}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/patients/:patient_no -- show patient by number
func (a *applicationDependencies) showPatientHandler(w http.ResponseWriter, r *http.Request) {
	patientNo := a.readPatientNoParam(r)

	patient, err := a.models.Patient.Get(patientNo)
	if err != nil {
		if errors.Is(err, errors.New("record not found")) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"patient": patient}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// GET /v1/patients?first_name=...&last_name=...&page=...&page_size=... -- list all (optionally filtered and paginated)
func (a *applicationDependencies) listPatientsHandler(w http.ResponseWriter, r *http.Request) {
	// Create a struct to hold the query parameters
	var queryParametersData struct {
		FirstName string
		LastName  string
		data.Filters
	}

	// get the query parameters from the URL
	queryParameters := r.URL.Query()

	// Load the query parameters into our struct
	queryParametersData.FirstName = a.getSingleQueryParameter(
		queryParameters,
		"first_name",
		"")

	queryParametersData.LastName = a.getSingleQueryParameter(
		queryParameters,
		"last_name",
		"")

	// Create a new validator instance
	v := validator.New()
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(
		queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(
		queryParameters, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(
		queryParameters, "sort", "patient_id")

	queryParametersData.Filters.SortSafeList = []string{"patient_id", "first_name", "last_name",
		"created_at", "-patient_id", "-first_name", "-last_name", "-created_at"}

	// Check if our filters are valid
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	patients, metadata, err := a.models.Patient.GetAll(
		queryParametersData.FirstName,
		queryParametersData.LastName,
		queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	resp := envelope{
		"patients":  patients,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// PUT/PATCH /v1/patients/:patient_no -- update (full or partial)
func (a *applicationDependencies) updatePatientHandler(w http.ResponseWriter, r *http.Request) {
	patientNo := a.readPatientNoParam(r)

	patient, err := a.models.Patient.Get(patientNo)
	if err != nil {
		if errors.Is(err, errors.New("record not found")) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		FirstName   *string `json:"first_name"`
		LastName    *string `json:"last_name"`
		DateOfBirth *string `json:"date_of_birth"`
		Gender      *string `json:"gender"`
		SSN         *string `json:"ssn"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Only update fields present in JSON (for PATCH semantics)
	if input.FirstName != nil {
		patient.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		patient.LastName = *input.LastName
	}
	if input.DateOfBirth != nil {
		patient.DateOfBirth = *input.DateOfBirth
	}
	if input.Gender != nil {
		patient.Gender = *input.Gender
	}
	if input.SSN != nil {
		patient.SSN = *input.SSN
	}

	v := validator.New()
	data.ValidatePatient(v, patient)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Patient.Update(patient)
	if err != nil {
		if errors.Is(err, errors.New("record not found")) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"patient": patient}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// DELETE /v1/patients/:patient_no
func (a *applicationDependencies) deletePatientHandler(w http.ResponseWriter, r *http.Request) {
	patientNo := a.readPatientNoParam(r)
	err := a.models.Patient.Delete(patientNo)
	if err != nil {
		if errors.Is(err, errors.New("record not found")) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}
	a.writeJSON(w, http.StatusOK, envelope{"message": "patient successfully deleted"}, nil)
}

// GET /v1/slow -- simulates slow database query for graceful shutdown demo
func (a *applicationDependencies) slowPatientHandler(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("simulating 10 second database query")

	// Simulate a slow database operation
	time.Sleep(10 * time.Second)

	patients, metadata, err := a.models.Patient.GetAll("", "", data.Filters{
		Page:         1,
		PageSize:     10,
		Sort:         "patient_id",
		SortSafeList: []string{"patient_id"},
	})
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{
		"patients":  patients,
		"@metadata": metadata,
		"message":   "This was a slow request (10s delay) to demonstrate graceful shutdown",
	}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
