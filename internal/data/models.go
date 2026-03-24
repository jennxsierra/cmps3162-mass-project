package data

import (
	"database/sql"
)

// Models groups all model types for easy access
type Models struct {
	Patient            PatientModel
	Provider           ProviderModel
	Staff              StaffModel
	ProviderSpecialty  ProviderSpecialtyModel
	PersonContact      PersonContactModel
	Appointment        AppointmentModel
	AppointmentType    AppointmentTypeModel
	Specialty          SpecialtyModel
	ContactType        ContactTypeModel
	CancellationReason CancellationReasonModel
}

// NewModels builds the Models struct with initialized model handlers
func NewModels(db *sql.DB) Models {
	return Models{
		Patient:            PatientModel{DB: db},
		Provider:           ProviderModel{DB: db},
		ProviderSpecialty:  ProviderSpecialtyModel{DB: db},
		PersonContact:      PersonContactModel{DB: db},
		Appointment:        AppointmentModel{DB: db},
		Staff:              StaffModel{DB: db},
		AppointmentType:    AppointmentTypeModel{DB: db},
		Specialty:          SpecialtyModel{DB: db},
		ContactType:        ContactTypeModel{DB: db},
		CancellationReason: CancellationReasonModel{DB: db},
	}
}
