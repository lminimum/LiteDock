package errors

import (
	"errors"
	"net/http"
)

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

// Remote machine errors.
var (
	ErrRemoteMachineNotFound = errors.New("remote machine: not found")
	ErrRemoteMachineExists   = errors.New("remote machine: already exists")
)

// SSH errors.
var (
	ErrSSHConnection = errors.New("ssh: connection failed")
	ErrSSHAuth       = errors.New("ssh: authentication failed")
	ErrSSHKeyParse   = errors.New("ssh: key parse failed")
)

// Docker errors.
var (
	ErrDockerConnection     = errors.New("docker: connection failed")
	ErrDockerOperation      = errors.New("docker: operation failed")
	ErrContainerNotFound    = errors.New("docker: container not found")
	ErrContainerExec        = errors.New("docker: container exec failed")
	ErrNetworkNotFound      = errors.New("docker: network not found")
	ErrNetworkHasContainers = errors.New("docker: network has active containers")
	ErrVolumeNotFound       = errors.New("docker: volume not found")
)

// SQL errors.
var ErrNoRows = errors.New("sql: no rows in result set")

// IsNoRows checks if the error is a "no rows" error from database/sql.
func IsNoRows(err error) bool {
	return err != nil && err.Error() == "sql: no rows in result set"
}

// HTTPStatusMap maps error types to HTTP status codes.
var HTTPStatusMap = map[error]int{
	// Not found errors -> 404
	ErrUserNotFound:          http.StatusNotFound,
	ErrRemoteMachineNotFound: http.StatusNotFound,
	ErrContainerNotFound:     http.StatusNotFound,
	ErrNotFound:              http.StatusNotFound,

	// Authentication errors -> 401
	ErrInvalidCredentials:   http.StatusUnauthorized,
	ErrInvalidToken:         http.StatusUnauthorized,
	ErrInvalidTokenClaims:   http.StatusUnauthorized,
	ErrUnexpectedSignMethod: http.StatusUnauthorized,

	// Bad request errors -> 400
	ErrInvalidInput: http.StatusBadRequest,

	// Conflict errors -> 409
	ErrUsernameExists:      http.StatusConflict,
	ErrRemoteMachineExists: http.StatusConflict,
	ErrAlreadyExists:       http.StatusConflict,

	// Internal server errors -> 500
	ErrDBURLRequired:      http.StatusInternalServerError,
	ErrDBURLRequiredMySQL: http.StatusInternalServerError,
	ErrDBTypeNotSupported: http.StatusInternalServerError,
	ErrSSHConnection:      http.StatusInternalServerError,
	ErrSSHAuth:            http.StatusInternalServerError,
	ErrSSHKeyParse:        http.StatusInternalServerError,
	ErrDockerConnection:   http.StatusInternalServerError,
	ErrDockerOperation:    http.StatusInternalServerError,
	ErrContainerExec:      http.StatusInternalServerError,
	ErrVolumeNotFound:     http.StatusNotFound,
}

// HTTPStatus returns the HTTP status code for an error.
// If the error is not in the map, it returns 500 (Internal Server Error).
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	for knownErr, status := range HTTPStatusMap {
		if errors.Is(err, knownErr) {
			return status
		}
	}

	return http.StatusInternalServerError
}
