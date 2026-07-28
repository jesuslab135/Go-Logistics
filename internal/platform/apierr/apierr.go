package apierr

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Error struct {
	Status  int
	Code    string
	Message string
	Details any
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) WithDetails(d any) *Error {
	e.Details = d
	return e
}

func (e *Error) Wrap(cause error) *Error {
	e.cause = cause
	return e
}

func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func BadRequest(message string) *Error { return New(http.StatusBadRequest, "bad_request", message) }
func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, "unauthorized", message)
}
func Forbidden(message string) *Error { return New(http.StatusForbidden, "forbidden", message) }
func NotFound(message string) *Error  { return New(http.StatusNotFound, "not_found", message) }
func Conflict(message string) *Error  { return New(http.StatusConflict, "conflict", message) }
func PayloadTooLarge(message string) *Error {
	return New(http.StatusRequestEntityTooLarge, "payload_too_large", message)
}
func UnsupportedMediaType(message string) *Error {
	return New(http.StatusUnsupportedMediaType, "unsupported_media_type", message)
}

func Validation(details any) *Error {
	return New(http.StatusUnprocessableEntity, "validation_failed", "validation failed").WithDetails(details)
}

func Internal(cause error) *Error {
	return New(http.StatusInternalServerError, "internal", "internal server error").Wrap(cause)
}

// Map turns any error into an *Error. Already-typed errors pass through; pgx and
// Postgres driver errors are translated to their HTTP equivalents so the sqlc
// layer never has to know about transport concerns. Everything else is a 500.
func Map(err error) *Error {
	if err == nil {
		return nil
	}

	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return NotFound("resource not found").Wrap(err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return Conflict("resource already exists").Wrap(err)
		case "23503": // foreign_key_violation
			return New(http.StatusConflict, "foreign_key_violation", "referenced resource constraint violated").Wrap(err)
		case "23502": // not_null_violation
			return BadRequest(fmt.Sprintf("missing required field %q", pgErr.ColumnName)).Wrap(err)
		case "23514": // check_violation
			return BadRequest("value violates a constraint").Wrap(err)
		}
	}

	return Internal(err)
}
