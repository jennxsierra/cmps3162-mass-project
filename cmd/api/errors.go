package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

// log an error message
func (a *applicationDependencies) logError(r *http.Request, err error) {

	method := r.Method
	uri := r.URL.RequestURI()
	a.logger.Error(err.Error(), "method", method, "uri", uri)

}

// send an error response in JSON
func (a *applicationDependencies) errorResponseJSON(w http.ResponseWriter,
	r *http.Request,
	status int,
	message any) {

	errorData := envelope{"error": message}
	err := a.writeJSON(w, status, errorData, nil)
	if err != nil {
		a.logError(r, err)
		w.WriteHeader(500)
	}
}

// send an error response if our server messes up
func (a *applicationDependencies) serverErrorResponse(w http.ResponseWriter,
	r *http.Request,
	err error) {

	// first thing is to log error message
	a.logError(r, err)
	// prepare a response to send to the client
	message := "the server encountered a problem and could not process your request"
	a.errorResponseJSON(w, r, http.StatusInternalServerError, message)
}

// send an error response if our client messes up with a 404
func (a *applicationDependencies) notFoundResponse(w http.ResponseWriter,
	r *http.Request) {

	// we only log server errors, not client errors
	// prepare a response to send to the client
	message := "the requested resource could not be found"
	a.errorResponseJSON(w, r, http.StatusNotFound, message)
}

// send an error response if our client messes up with a 405
func (a *applicationDependencies) methodNotAllowedResponse(
	w http.ResponseWriter,
	r *http.Request) {

	// we only log server errors, not client errors
	// prepare a formatted response to send to the client
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)

	a.errorResponseJSON(w, r, http.StatusMethodNotAllowed, message)
}

// send an error response if our client messes up with a 400 (bad request)
func (a *applicationDependencies) badRequestResponse(w http.ResponseWriter,
	r *http.Request,
	err error) {

	a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

// send an error response if rate limit exceeded (429 - Too Many Requests)
func (a *applicationDependencies) rateLimitExceededResponse(
	w http.ResponseWriter,
	r *http.Request,
	retry time.Duration,
) {
	if retry <= 0 {
		retry = time.Second
	}

	seconds := int(math.Ceil(retry.Seconds()))
	if seconds < 1 {
		seconds = 1
	}

	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	a.errorResponseJSON(w, r, http.StatusTooManyRequests, "rate limit exceeded")
}

// send an error response if there is an edit conflict (409 - Conflict)
func (a *applicationDependencies) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	a.errorResponseJSON(w, r, http.StatusConflict, message)
}

// send an error response if the client provides invalid authentication credentials (401 - Unauthorized)
func (a *applicationDependencies) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	a.errorResponseJSON(w, r, http.StatusUnauthorized, message)
}
