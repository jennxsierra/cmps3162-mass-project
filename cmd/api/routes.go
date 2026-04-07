package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {
	router := httprouter.New()

	// Handle 404 and 405 errors with our custom responses
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	// Healtcheck Route
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.requireActivatedUser(a.healthcheckHandler))

	// Metrics Route
	router.HandlerFunc(http.MethodGet, "/v1/metrics", a.requireActivatedUser(func(w http.ResponseWriter, r *http.Request) {
		expvar.Handler().ServeHTTP(w, r)
	}))

	// Demo route for graceful shutdown
	router.HandlerFunc(http.MethodGet, "/v1/slow", a.slowPatientHandler)

	// Patient routes
	router.HandlerFunc(http.MethodGet, "/v1/patients", a.requireActivatedUser(a.listPatientsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/patients", a.requireActivatedUser(a.createPatientHandler))
	router.HandlerFunc(http.MethodGet, "/v1/patients/:patient_no", a.requireActivatedUser(a.showPatientHandler))
	router.HandlerFunc(http.MethodPut, "/v1/patients/:patient_no", a.requireActivatedUser(a.updatePatientHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/patients/:patient_no", a.requireActivatedUser(a.updatePatientHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/patients/:patient_no", a.requireActivatedUser(a.deletePatientHandler))

	// Provider routes
	router.HandlerFunc(http.MethodGet, "/v1/providers", a.requireActivatedUser(a.listProvidersHandler))
	router.HandlerFunc(http.MethodPost, "/v1/providers", a.requireActivatedUser(a.createProviderHandler))
	router.HandlerFunc(http.MethodGet, "/v1/providers/:license_no", a.requireActivatedUser(a.showProviderHandler))
	router.HandlerFunc(http.MethodPut, "/v1/providers/:license_no", a.requireActivatedUser(a.updateProviderHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/providers/:license_no", a.requireActivatedUser(a.updateProviderHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/providers/:license_no", a.requireActivatedUser(a.deleteProviderHandler))
	router.HandlerFunc(http.MethodPost, "/v1/providers/:license_no/specialties", a.requireActivatedUser(a.createProviderSpecialtyHandler))
	router.HandlerFunc(http.MethodGet, "/v1/providers/:license_no/specialties", a.requireActivatedUser(a.listProviderSpecialtiesHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/providers/:license_no/specialties/:specialty_id", a.requireActivatedUser(a.deleteProviderSpecialtyHandler))

	// Staff routes
	router.HandlerFunc(http.MethodGet, "/v1/staff", a.requireActivatedUser(a.listStaffHandler))
	router.HandlerFunc(http.MethodPost, "/v1/staff", a.requireActivatedUser(a.createStaffHandler))
	router.HandlerFunc(http.MethodGet, "/v1/staff/:staff_no", a.requireActivatedUser(a.showStaffHandler))
	router.HandlerFunc(http.MethodPut, "/v1/staff/:staff_no", a.requireActivatedUser(a.updateStaffHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/staff/:staff_no", a.requireActivatedUser(a.updateStaffHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/staff/:staff_no", a.requireActivatedUser(a.deleteStaffHandler))

	// User routes
	router.HandlerFunc(http.MethodPost, "/v1/users", a.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", a.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", a.createAuthenticationTokenHandler)

	// Appointment Type routes
	router.HandlerFunc(http.MethodGet, "/v1/appointment-types", a.requireActivatedUser(a.listAppointmentTypesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/appointment-types", a.requireActivatedUser(a.createAppointmentTypeHandler))
	router.HandlerFunc(http.MethodGet, "/v1/appointment-types/:id", a.requireActivatedUser(a.showAppointmentTypeHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/appointment-types/:id", a.requireActivatedUser(a.updateAppointmentTypeHandler))

	// Specialties routes
	router.HandlerFunc(http.MethodGet, "/v1/specialties", a.requireActivatedUser(a.listSpecialtiesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/specialties", a.requireActivatedUser(a.createSpecialtyHandler))
	router.HandlerFunc(http.MethodGet, "/v1/specialties/:id", a.requireActivatedUser(a.showSpecialtyHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/specialties/:id", a.requireActivatedUser(a.updateSpecialtyHandler))

	// Contact Types routes (read-only)
	router.HandlerFunc(http.MethodGet, "/v1/contact-types", a.requireActivatedUser(a.listContactTypesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/contact-types/:id", a.requireActivatedUser(a.showContactTypeHandler))

	// Cancellation Reasons routes (read-only)
	router.HandlerFunc(http.MethodGet, "/v1/cancellation-reasons", a.requireActivatedUser(a.listCancellationReasonsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/cancellation-reasons/:id", a.requireActivatedUser(a.showCancellationReasonHandler))

	// Person Contacts routes
	router.HandlerFunc(http.MethodPost, "/v1/persons/:id/contacts", a.requireActivatedUser(a.createPersonContactHandler))
	router.HandlerFunc(http.MethodGet, "/v1/persons/:id/contacts", a.requireActivatedUser(a.listPersonContactsHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/persons/:id/contacts/:contact_id", a.requireActivatedUser(a.updatePersonContactHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/persons/:id/contacts/:contact_id", a.requireActivatedUser(a.deletePersonContactHandler))

	// Appointment routes
	router.HandlerFunc(http.MethodPost, "/v1/appointments", a.requireActivatedUser(a.createAppointmentHandler))
	router.HandlerFunc(http.MethodGet, "/v1/appointments", a.requireActivatedUser(a.listAppointmentsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/appointments/:id", a.requireActivatedUser(a.showAppointmentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/appointments/:id", a.requireActivatedUser(a.updateAppointmentHandler))
	router.HandlerFunc(http.MethodPost, "/v1/appointments/:id/cancellations", a.requireActivatedUser(a.createAppointmentCancellationHandler))
	router.HandlerFunc(http.MethodGet, "/v1/appointments/:id/cancellations", a.requireActivatedUser(a.showAppointmentCancellationHandler))

	return a.loggingMiddleware(a.metrics(a.recoverPanic(a.enableCORS(a.rateLimit(a.gzip(a.authenticate(router)))))))
}
