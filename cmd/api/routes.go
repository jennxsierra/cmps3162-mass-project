package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	// setup a new router
	router := httprouter.New()

	// handle 404
	router.NotFound = http.HandlerFunc(a.notFoundResponse)

	// handle 405
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	// setup routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthcheckHandler)

	// Demo route for graceful shutdown
	router.HandlerFunc(http.MethodGet, "/v1/slow", a.slowPatientHandler)

	// Patient routes
	router.HandlerFunc(http.MethodGet, "/v1/patients", a.listPatientsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/patients", a.createPatientHandler)
	router.HandlerFunc(http.MethodGet, "/v1/patients/:patient_no", a.showPatientHandler)
	router.HandlerFunc(http.MethodPut, "/v1/patients/:patient_no", a.updatePatientHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/patients/:patient_no", a.updatePatientHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/patients/:patient_no", a.deletePatientHandler)

	// Provider routes
	router.HandlerFunc(http.MethodGet, "/v1/providers", a.listProvidersHandler)
	router.HandlerFunc(http.MethodPost, "/v1/providers", a.createProviderHandler)
	router.HandlerFunc(http.MethodGet, "/v1/providers/:license_no", a.showProviderHandler)
	router.HandlerFunc(http.MethodPut, "/v1/providers/:license_no", a.updateProviderHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/providers/:license_no", a.updateProviderHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/providers/:license_no", a.deleteProviderHandler)

	// Staff routes
	router.HandlerFunc(http.MethodGet, "/v1/staff", a.listStaffHandler)
	router.HandlerFunc(http.MethodPost, "/v1/staff", a.createStaffHandler)
	router.HandlerFunc(http.MethodGet, "/v1/staff/:staff_no", a.showStaffHandler)
	router.HandlerFunc(http.MethodPut, "/v1/staff/:staff_no", a.updateStaffHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/staff/:staff_no", a.updateStaffHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/staff/:staff_no", a.deleteStaffHandler)

	// Appointment Type routes
	router.HandlerFunc(http.MethodGet, "/v1/appointment-types", a.listAppointmentTypesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/appointment-types", a.createAppointmentTypeHandler)
	router.HandlerFunc(http.MethodGet, "/v1/appointment-types/:id", a.showAppointmentTypeHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/appointment-types/:id", a.updateAppointmentTypeHandler)

	// Request sent first to recoverPanic() then sent to loggingMiddleware()
	// then sent to rateLimit() and finally sent to the router
	return a.recoverPanic(a.loggingMiddleware(a.rateLimit(router)))
}
