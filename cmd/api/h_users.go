package main

import (
	"errors"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
)

// POST /v1/users -- register a user account
func (a *applicationDependencies) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Username:  incomingData.Username,
		Email:     incomingData.Email,
		Activated: false,
	}

	err = user.Password.Set(incomingData.Password)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateUser(v, user)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.User.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
