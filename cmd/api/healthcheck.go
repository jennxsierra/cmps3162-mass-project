package main

import (
	"net/http"
)

func (a *applicationDependencies) healthcheckHandler(w http.ResponseWriter,
	r *http.Request) {

	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": a.config.environment,
			"version":     appVersion,
		},
	}

	err := a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
