package data

import (
	"strings"

	"github.com/jennxsierra/rate-limiting-shutdown/internal/validator"
)

// The Filters type will contain the fields related to pagination
// and eventually the fields related to sorting.
type Filters struct {
	Page         int      // which page number does the client want
	PageSize     int      // how many records per page
	Sort         string   // sort field
	SortSafeList []string // allowed sort fields
}

// Validate page and PageSize
// We follow the same approach that we used to validate a Patient
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 500, "page", "must be a maximum of 500")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Check if sort fields provided are valid
	v.Check(validator.PermittedValue(f.Sort, f.SortSafeList...), "sort",
		"invalid sort value")
}

// calculate how many records to send back
func (f Filters) limit() int {
	return f.PageSize
}

// calculate the offset so that we remember how many records have
// been sent and how many remain to be sent
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// Implement the sorting feature
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	// don't allow the operation to continue
	// if case of SQL injection attack
	panic("unsafe sort parameter: " + f.Sort)
}

// Get the sort order
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// define a type to hold the metadata
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// calculate the metadata
func calculateMetadata(totalRecords int, currentPage int, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}
