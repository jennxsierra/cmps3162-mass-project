package main

import (
	"context"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
)

type contextKey string

const userContextKey = contextKey("user") 

// contextSetUser adds the provided user to the request context and returns the modified request.
func (a *applicationDependencies) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser retrieves the user from the request context. If no user is found, it panics.
func (a *applicationDependencies) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
