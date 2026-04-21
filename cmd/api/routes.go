package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {
	router := httprouter.New()

	/**** BACKEND ****/

	// Handle 404 and 405 errors with our custom responses
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	// Healtcheck Route
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.requirePermission("system:read", a.healthcheckHandler))

	// Metrics Route
	router.HandlerFunc(http.MethodGet, "/v1/metrics", a.requirePermission("system:read", func(w http.ResponseWriter, r *http.Request) {
		expvar.Handler().ServeHTTP(w, r)
	}))

	// Demo route for graceful shutdown
	router.HandlerFunc(http.MethodGet, "/v1/slow", a.slowPatientHandler)

	/****  DB SCHEMA ****/

	// Patient routes
	router.HandlerFunc(http.MethodGet, "/v1/patients", a.requirePermission("patients:read", a.listPatientsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/patients", a.requirePermission("patients:write", a.createPatientHandler))
	router.HandlerFunc(http.MethodGet, "/v1/patients/:patient_no", a.requirePermission("patients:read", a.showPatientHandler))
	router.HandlerFunc(http.MethodPut, "/v1/patients/:patient_no", a.requirePermission("patients:write", a.updatePatientHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/patients/:patient_no", a.requirePermission("patients:write", a.updatePatientHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/patients/:patient_no", a.requirePermission("patients:write", a.deletePatientHandler))

	// Provider routes
	router.HandlerFunc(http.MethodGet, "/v1/providers", a.requirePermission("providers:read", a.listProvidersHandler))
	router.HandlerFunc(http.MethodPost, "/v1/providers", a.requirePermission("providers:write", a.createProviderHandler))
	router.HandlerFunc(http.MethodGet, "/v1/providers/:license_no", a.requirePermission("providers:read", a.showProviderHandler))
	router.HandlerFunc(http.MethodPut, "/v1/providers/:license_no", a.requirePermission("providers:write", a.updateProviderHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/providers/:license_no", a.requirePermission("providers:write", a.updateProviderHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/providers/:license_no", a.requirePermission("providers:write", a.deleteProviderHandler))
	router.HandlerFunc(http.MethodPost, "/v1/providers/:license_no/specialties", a.requirePermission("providers:write", a.createProviderSpecialtyHandler))
	router.HandlerFunc(http.MethodGet, "/v1/providers/:license_no/specialties", a.requirePermission("providers:read", a.listProviderSpecialtiesHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/providers/:license_no/specialties/:specialty_id", a.requirePermission("providers:write", a.deleteProviderSpecialtyHandler))

	// Staff routes
	router.HandlerFunc(http.MethodGet, "/v1/staff", a.requirePermission("staff:read", a.listStaffHandler))
	router.HandlerFunc(http.MethodPost, "/v1/staff", a.requirePermission("staff:write", a.createStaffHandler))
	router.HandlerFunc(http.MethodGet, "/v1/staff/:staff_no", a.requirePermission("staff:read", a.showStaffHandler))
	router.HandlerFunc(http.MethodPut, "/v1/staff/:staff_no", a.requirePermission("staff:write", a.updateStaffHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/staff/:staff_no", a.requirePermission("staff:write", a.updateStaffHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/staff/:staff_no", a.requirePermission("staff:write", a.deleteStaffHandler))

	// User routes
	router.HandlerFunc(http.MethodPost, "/v1/users", a.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", a.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", a.createAuthenticationTokenHandler)

	// Appointment Type routes
	router.HandlerFunc(http.MethodGet, "/v1/appointment-types", a.requirePermission("appointment-types:read", a.listAppointmentTypesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/appointment-types", a.requirePermission("appointment-types:write", a.createAppointmentTypeHandler))
	router.HandlerFunc(http.MethodGet, "/v1/appointment-types/:id", a.requirePermission("appointment-types:read", a.showAppointmentTypeHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/appointment-types/:id", a.requirePermission("appointment-types:write", a.updateAppointmentTypeHandler))

	// Specialties routes
	router.HandlerFunc(http.MethodGet, "/v1/specialties", a.requirePermission("specialties:read", a.listSpecialtiesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/specialties", a.requirePermission("specialties:write", a.createSpecialtyHandler))
	router.HandlerFunc(http.MethodGet, "/v1/specialties/:id", a.requirePermission("specialties:read", a.showSpecialtyHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/specialties/:id", a.requirePermission("specialties:write", a.updateSpecialtyHandler))

	// Contact Types routes (read-only)
	router.HandlerFunc(http.MethodGet, "/v1/contact-types", a.requirePermission("contact-types:read", a.listContactTypesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/contact-types/:id", a.requirePermission("contact-types:read", a.showContactTypeHandler))

	// Cancellation Reasons routes (read-only)
	router.HandlerFunc(http.MethodGet, "/v1/cancellation-reasons", a.requirePermission("cancellation-reasons:read", a.listCancellationReasonsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/cancellation-reasons/:id", a.requirePermission("cancellation-reasons:read", a.showCancellationReasonHandler))

	// Person Contacts routes
	router.HandlerFunc(http.MethodPost, "/v1/persons/:id/contacts", a.requirePermission("person-contacts:write", a.createPersonContactHandler))
	router.HandlerFunc(http.MethodGet, "/v1/persons/:id/contacts", a.requirePermission("person-contacts:read", a.listPersonContactsHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/persons/:id/contacts/:contact_id", a.requirePermission("person-contacts:write", a.updatePersonContactHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/persons/:id/contacts/:contact_id", a.requirePermission("person-contacts:write", a.deletePersonContactHandler))

	// Appointment routes
	router.HandlerFunc(http.MethodPost, "/v1/appointments", a.requirePermission("appointments:write", a.createAppointmentHandler))
	router.HandlerFunc(http.MethodGet, "/v1/appointments", a.requirePermission("appointments:read", a.listAppointmentsHandler))
	router.HandlerFunc(http.MethodGet, "/v1/appointments/:id", a.requirePermission("appointments:read", a.showAppointmentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/appointments/:id", a.requirePermission("appointments:write", a.updateAppointmentHandler))
	router.HandlerFunc(http.MethodPost, "/v1/appointments/:id/cancellations", a.requirePermission("appointments:write", a.createAppointmentCancellationHandler))
	router.HandlerFunc(http.MethodGet, "/v1/appointments/:id/cancellations", a.requirePermission("appointments:read", a.showAppointmentCancellationHandler))

	/**** FRONTEND ****/

	router.ServeFiles("/static/*filepath", http.Dir("./ui/static"))

	router.HandlerFunc(http.MethodGet, "/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./ui/static/pages/login.html")
	})

	router.HandlerFunc(http.MethodGet, "/appointments", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./ui/static/pages/appointments.html")
	})

	/** MIDDLEWARE */

	return a.loggingMiddleware(a.metrics(a.recoverPanic(a.enableCORS(a.rateLimit(a.gzip(a.authenticate(router)))))))
}
