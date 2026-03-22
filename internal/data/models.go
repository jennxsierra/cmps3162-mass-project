package data

import (
	"database/sql"
)

// Models groups all model types for easy access
type Models struct {
	Patient         PatientModel
	Provider        ProviderModel
	Staff           StaffModel
	AppointmentType AppointmentTypeModel
}

// NewModels builds the Models struct with initialized model handlers
func NewModels(db *sql.DB) Models {
	return Models{
		Patient:         PatientModel{DB: db},
		Provider:        ProviderModel{DB: db},
		Staff:           StaffModel{DB: db},
		AppointmentType: AppointmentTypeModel{DB: db},
	}
}
