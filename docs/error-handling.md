# Error Handling

## Overview

LiteDock uses a unified error handling approach through the `pkg/errors` package.

## Static Errors

Static errors are pre-defined in `pkg/errors/errors.go` using the `"domain: description"` format:

```go
var ErrUserNotFound       = errors.New("user: not found")
var ErrInvalidCredentials = errors.New("auth: invalid credentials")
var ErrNotFound           = errors.New("database: record not found")
```

## Error Wrapping

Use `errors.Wrap()` with `Package.Method.Operation` format (PascalCase, dot-separated):

```go
// Good
return errors.Wrap(err, "UserRepo.CreateUser.Exec")
return errors.Wrap(err, "Auth.Login.VerifyPassword")
return errors.Wrap(err, "Postgres.New.ParseConfig")

// Bad - using fmt.Errorf
return fmt.Errorf("UserRepo - CreateUser - r.db.Exec: %w", err)
```

## Error Format

Wrapped errors produce: `Package.Method.Operation: original_error`

Examples:
- `UserRepo.CreateUser.Exec: connection refused`
- `Auth.Login.VerifyPassword: auth: invalid credentials`
- `Postgres.New.ParseConfig: invalid URL`

## Wrap Function

```go
func Wrap(err error, msg string) error
```

- If `err` is `nil`, returns `nil`
- Otherwise, returns new error with format `msg: err.Error()`
- Supports `Unwrap()` for use with `errors.Is()` and `errors.As()`

## Error Checking

The package re-exports `errors.Is` and `errors.As` for convenience:

```go
import apperrors "github.com/lminimum/LiteDock/pkg/errors"

// Using errors.Is
if apperrors.Is(err, apperrors.ErrUserNotFound) {
    // handle user not found
}

// Using errors.As
var target *MyError
if apperrors.As(err, &target) {
    // handle typed error
}
```

## Usage Rules

1. **Static errors** — Define in `pkg/errors/errors.go` with `"domain: description"` format
2. **Dynamic errors** — Use `errors.Wrap(err, "Package.Method.Operation")`
3. **Context format** — `Package.Method.Operation` (PascalCase, dot-separated)
4. **Error checking** — Use `errors.Is(err, apperrors.ErrXxx)` for sentinel comparison
5. **Never use** — `fmt.Errorf()` for wrapping errors
