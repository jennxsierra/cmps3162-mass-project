package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jennxsierra/mass-project/internal/data"
	"github.com/jennxsierra/mass-project/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) readStaffNoParam(r *http.Request) (string, error) {
	params := httprouter.ParamsFromContext(r.Context())
	staffNo := params.ByName("staff_no")
	if staffNo == "" {
		return "", errors.New("invalid staff_no")
	}
	return staffNo, nil
}

func (a *applicationDependencies) createStaffHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		StaffNo     string `json:"staff_no"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		DateOfBirth string `json:"date_of_birth"`
		Gender      string `json:"gender"`
	}

	err := a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	staff := &data.Staff{
		StaffNo:     input.StaffNo,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		DateOfBirth: input.DateOfBirth,
		Gender:      input.Gender,
	}

	v := validator.New()
	data.ValidateStaff(v, staff)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Staff.Insert(staff)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/staff/%s", staff.StaffNo))

	err = a.writeJSON(w, http.StatusCreated, envelope{"staff": staff}, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) showStaffHandler(w http.ResponseWriter, r *http.Request) {
	staffNo, err := a.readStaffNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	staff, err := a.models.Staff.Get(staffNo)
	if err != nil {
		if err == data.ErrRecordNotFound {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"staff": staff}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listStaffHandler(w http.ResponseWriter, r *http.Request) {
	var queryParametersData struct {
		Search string
		data.Filters
	}

	queryParameters := r.URL.Query()
	queryParametersData.Search = a.getSingleQueryParameter(queryParameters, "search", "")

	v := validator.New()
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "staff_id")
	queryParametersData.Filters.SortSafeList = []string{
		"staff_id", "staff_no", "first_name", "last_name", "created_at",
		"-staff_id", "-staff_no", "-first_name", "-last_name", "-created_at",
	}

	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	staffMembers, metadata, err := a.models.Staff.GetAll(
		queryParametersData.Search,
		queryParametersData.Filters,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	resp := envelope{
		"staff":     staffMembers,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) updateStaffHandler(w http.ResponseWriter, r *http.Request) {
	staffNo, err := a.readStaffNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	staff, err := a.models.Staff.Get(staffNo)
	if err != nil {
		if err == data.ErrRecordNotFound {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		StaffNo     *string `json:"staff_no"`
		FirstName   *string `json:"first_name"`
		LastName    *string `json:"last_name"`
		DateOfBirth *string `json:"date_of_birth"`
		Gender      *string `json:"gender"`
	}

	err = a.readJSON(w, r, &input)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if input.StaffNo != nil {
		staff.StaffNo = *input.StaffNo
	}
	if input.FirstName != nil {
		staff.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		staff.LastName = *input.LastName
	}
	if input.DateOfBirth != nil {
		staff.DateOfBirth = *input.DateOfBirth
	}
	if input.Gender != nil {
		staff.Gender = *input.Gender
	}

	v := validator.New()
	data.ValidateStaff(v, staff)
	if !v.IsEmpty() {
		a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	err = a.models.Staff.Update(staff)
	if err != nil {
		if err == data.ErrRecordNotFound {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"staff": staff}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteStaffHandler(w http.ResponseWriter, r *http.Request) {
	staffNo, err := a.readStaffNoParam(r)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.models.Staff.Delete(staffNo)
	if err != nil {
		switch err {
		case data.ErrRecordNotFound:
			a.notFoundResponse(w, r)
		case data.ErrStaffInUse:
			a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, "staff has related records and cannot be deleted")
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.writeJSON(w, http.StatusOK, envelope{"message": "staff successfully deleted"}, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
