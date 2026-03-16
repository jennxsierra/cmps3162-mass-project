package validator

import (
	"slices"
)

type Validator struct {
	Errors map[string]string
}

// New creates a new Validator instance with an initialized error map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// IsEmpty returns true if the Validator contains no errors, otherwise false
func (v *Validator) IsEmpty() bool {
	return len(v.Errors) == 0
}

// AddError adds a validation error to the Validator's error map,
// only if an entry doesn't already exist for the given key
func (v *Validator) AddError(key string, message string) {
	_, exists := v.Errors[key]
	if !exists {
		v.Errors[key] = message
	}
}

// Check evaluates a condition and adds an error to the Validator's error map
// if the condition is false
func (v *Validator) Check(acceptable bool, key string, message string) {
	if !acceptable {
		v.AddError(key, message)
	}
}

// Check for permitted values
func PermittedValue(value string, permittedValues ...string) bool {
	return slices.Contains(permittedValues, value)
}
