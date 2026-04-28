package repository

import "errors"

var (
	ErrNotFound = errors.New("not found")

	ErrRegistrationNotFound    = errors.New("registration not found")
	ErrRegistrationAlreadyDone = errors.New("registration already processed")
	ErrRegistrationEmailInUse  = errors.New("registration email already in use")
	ErrApprovedCompanyNotFound = errors.New("approved company not found")
)
