package main

import (
	"errors"
	"net/http"
	"time"

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

	// generate a new activation token for the user that expires in 3 days
	token, err := a.models.Token.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	responseData := envelope{"user": user}

	// send the welcome email in the background
	a.background(func() {
		payload := map[string]any{
			"Username":        user.Username,
			"Email":           user.Email,
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		sendErr := a.mailer.Send(user.Email, "user_welcome.tmpl", payload)
		if sendErr != nil {
			a.logger.Error(sendErr.Error())
		}
	})

	err = a.writeJSON(w, http.StatusCreated, responseData, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// PUT /v1/users/activated -- activate a user account with activation token
func (a *applicationDependencies) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		TokenPlaintext string `json:"token"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateTokenPlaintext(v, incomingData.TokenPlaintext)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	user, err := a.models.User.GetForToken(data.ScopeActivation, incomingData.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = a.models.User.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			a.editConflictResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.models.Token.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
