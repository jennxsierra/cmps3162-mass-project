package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jennxsierra/mass-project/internal/validator"
	"github.com/julienschmidt/httprouter"
)

// create an envelope type
type envelope map[string]any

func (a *applicationDependencies) writeJSON(w http.ResponseWriter,
	status int, data envelope,
	headers http.Header) error {
	jsResponse, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	jsResponse = append(jsResponse, '\n')

	// additional headers to be set
	for key, value := range headers {
		w.Header()[key] = value
	}

	// set content type header
	w.Header().Set("Content-Type", "application/json")

	// explicitly set the response status code
	w.WriteHeader(status)
	_, err = w.Write(jsResponse)
	if err != nil {
		return err
	}

	return nil
}

func (a *applicationDependencies) readJSON(w http.ResponseWriter,
	r *http.Request,
	destination any) error {

	// what is the max size of the request body (250KB seems reasonable)
	maxBytes := 256_000
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// our decoder will check for unknown fields
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// let start the decoding
	err := dec.Decode(destination)

	if err != nil {
		// check for the different errors
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("the body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// Decode can also send back an io error message
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("the body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("the body contains the incorrect JSON type for field %q",
					unmarshalTypeError.Field)
			}
			return fmt.Errorf("the body contains the incorrect  JSON type (at character %d)",
				unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("the body must not be empty")

			// check for unknown field error
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(),
				"json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// does the body exceed our limit of 250KB?
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("the body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.Is(err, io.EOF):
			return errors.New("the body must not be empty")

		// the programmer messed up
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// some other type of error
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) { // there is more data present
		return errors.New("the body must only contain a single JSON value")
	}

	return nil
}

// getSingleQueryParameter retrieves a single query parameter from the URL
// url.Values is a key:value hash map of the query parameters
func (a *applicationDependencies) getSingleQueryParameter(
	queryParameters url.Values,
	key string,
	defaultValue string) string {
	result := queryParameters.Get(key)
	if result == "" {
		return defaultValue
	}
	return result
}

// getMultipleQueryParameters retrieves comma-separated query parameters from the URL
func (a *applicationDependencies) getMultipleQueryParameters(
	queryParameters url.Values,
	key string,
	defaultValue []string) []string {
	result := queryParameters.Get(key)
	if result == "" {
		return defaultValue
	}
	return strings.Split(result, ",")
}

// getSingleIntegerParameter retrieves an integer parameter from the URL
// This method can cause a validation error when trying to convert the
// string to a valid integer value
func (a *applicationDependencies) getSingleIntegerParameter(
	queryParameters url.Values,
	key string,
	defaultValue int,
	v *validator.Validator) int {
	result := queryParameters.Get(key)
	if result == "" {
		return defaultValue
	}
	// try to convert to an integer
	intValue, err := strconv.Atoi(result)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return intValue
}

// readIntIDParam retrieves an integer ID parameter from the URL path
func (a *applicationDependencies) readIntIDParam(r *http.Request, paramName string) (int, error) {
	params := httprouter.ParamsFromContext(r.Context())
	idString := params.ByName(paramName)

	id, err := strconv.Atoi(idString)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// background accepts a function and runs it in the background as a goroutine
// it also recovers from any panic and logs the error message
func (a *applicationDependencies) background(fn func()) {
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()

		defer func() {
			err := recover()
			if err != nil {
				a.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		fn()
	}()
}
