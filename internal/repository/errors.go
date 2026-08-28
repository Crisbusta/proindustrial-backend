package repository

import (
	"errors"

	"github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("not found")

	ErrRegistrationNotFound    = errors.New("registration not found")
	ErrRegistrationAlreadyDone = errors.New("registration already processed")
	ErrRegistrationEmailInUse  = errors.New("registration email already in use")
	ErrApprovedCompanyNotFound = errors.New("approved company not found")
)

// IsUniqueViolation reconoce el error 23505 de Postgres, para poder
// traducir una colisión de índice único en un 409 con mensaje útil en
// vez de un 500 genérico.
func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
