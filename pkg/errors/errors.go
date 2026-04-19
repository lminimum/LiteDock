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
	ErrDockerConnection = errors.New("docker: connection failed")
	ErrDockerOperation  = errors.New("docker: operation failed")
	ErrContainerNotFound = errors.New("docker: container not found")
	ErrContainerExec     = errors.New("docker: container exec failed")
)
