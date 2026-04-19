package errors

import "errors"

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &wrappedError{msg: msg, err: err}
}

type wrappedError struct {
	msg string
	err error
}

func (e *wrappedError) Error() string {
	return e.msg + ": " + e.err.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.err
}

// Database errors.
var (
	ErrDBURLRequired      = errors.New("database: DB_URL is required for postgres")
	ErrDBURLRequiredMySQL = errors.New("database: DB_URL is required for mysql")
	ErrDBTypeNotSupported = errors.New("database: unsupported database type")
)

// Test errors.
var ErrTestFail = errors.New("fail")

// User errors.
var ErrUserNotFound = errors.New("user: not found")

// Auth errors.
var (
	ErrInvalidCredentials   = errors.New("auth: invalid credentials")
	ErrUsernameExists       = errors.New("auth: username already exists")
	ErrInvalidToken         = errors.New("auth: invalid token")
	ErrInvalidTokenClaims   = errors.New("auth: invalid token claims")
	ErrUnexpectedSignMethod = errors.New("auth: unexpected signing method")
)

// Scan errors.
var (
	ErrScanRowNil       = errors.New("scan: nil row")
	ErrScanRowNoScanner = errors.New("scan: row does not implement Scanner interface")
)

var (
	ErrNotFound      = errors.New("database: record not found")
	ErrAlreadyExists = errors.New("database: record already exists")
	ErrInvalidInput  = errors.New("database: invalid input")
)
