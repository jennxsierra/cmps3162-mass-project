package data

import (
	"errors"
	"time"

	"github.com/jennxsierra/mass-project/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// define the password type: plaintext and hash
// lowercase to keep it from being exported and made public
type password struct {
	plaintext *string
	hash      []byte
}

// Set method to hash the plaintext password and store it in the password struct
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches method to compare the plaintext password with the hashed password
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil // password matches
}

// ValidateEmail checks if the email is provided and valid
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// ValidatePasswordPlaintext checks if the plaintext password
// is provided and meets length requirements
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

// ValidateUser checks if the user struct has valid fields
func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Username != "", "username", "must be provided")
	v.Check(len(user.Username) <= 200, "username", "must not be more than 200 bytes long")

	// validate the email for the user
	ValidateEmail(v, user.Email)

	// if the plaintext password is not nil, validate it
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	// check if the password hash is present, if not, panic
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
