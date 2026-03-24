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
	router.HandlerFunc(http.MethodPost, "/v1/providers/:license_no/specialties", a.createProviderSpecialtyHandler)
	router.HandlerFunc(http.MethodGet, "/v1/providers/:license_no/specialties", a.listProviderSpecialtiesHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/providers/:license_no/specialties/:specialty_id", a.deleteProviderSpecialtyHandler)

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

	// Specialties routes
	router.HandlerFunc(http.MethodGet, "/v1/specialties", a.listSpecialtiesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/specialties", a.createSpecialtyHandler)
	router.HandlerFunc(http.MethodGet, "/v1/specialties/:id", a.showSpecialtyHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/specialties/:id", a.updateSpecialtyHandler)

	// Contact Types routes (read-only)
	router.HandlerFunc(http.MethodGet, "/v1/contact-types", a.listContactTypesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/contact-types/:id", a.showContactTypeHandler)

	// Cancellation Reasons routes (read-only)
	router.HandlerFunc(http.MethodGet, "/v1/cancellation-reasons", a.listCancellationReasonsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/cancellation-reasons/:id", a.showCancellationReasonHandler)

	// Request sent first to recoverPanic() then sent to loggingMiddleware()
	// then sent to rateLimit() and finally sent to the router
	return a.recoverPanic(a.loggingMiddleware(a.rateLimit(router)))
}
