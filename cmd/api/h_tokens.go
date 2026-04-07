package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// POST /v1/tokens/authentication -- create an authentication token
func (a *applicationDependencies) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// read the email and password from the request body
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// validate the email and password provided by the client
	v := validator.New()
	data.ValidateEmail(v, incomingData.Email)
	data.ValidatePasswordPlaintext(v, incomingData.Password)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	// retrieve the user from the database based on the email provided
	user, err := a.models.User.GetByEmail(incomingData.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.invalidCredentialsResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// check if the password provided matches the password stored in the database
	passwordMatch, err := user.Password.Matches(incomingData.Password)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if !passwordMatch { // return 401 if the password doesn't match
		a.invalidCredentialsResponse(w, r)
		return
	}

	// generate a new authentication token for the user that expires in 24 hours
	token, err := a.models.Token.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// return 201 status code along with the token in the response body
	err = a.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
