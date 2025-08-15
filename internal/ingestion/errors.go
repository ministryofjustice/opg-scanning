package ingestion

import "errors"

var ErrScannedCaseResponseUIDMissing = errors.New("scannedCaseResponse UID missing")

type Problem struct {
	Title            string
	ValidationErrors []string
}

func (p Problem) Error() string {
	return p.Title
}

type FailedToCreateCaseStubError struct {
	Err error
}

func (e FailedToCreateCaseStubError) Error() string { return e.Err.Error() }
func (e FailedToCreateCaseStubError) Unwrap() error { return e.Err }

type ValidateSetError struct {
	Err error
}

func (e ValidateSetError) Error() string { return e.Err.Error() }
func (e ValidateSetError) Unwrap() error { return e.Err }

type ValidateAndSanitizeError struct {
	Err error
}

func (e ValidateAndSanitizeError) Error() string { return e.Err.Error() }
func (e ValidateAndSanitizeError) Unwrap() error { return e.Err }

type PersistSetError struct {
	Err error
}

func (e PersistSetError) Error() string { return e.Err.Error() }
func (e PersistSetError) Unwrap() error { return e.Err }
