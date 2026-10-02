package retry

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Err       error
	Retryable bool
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Retryable(err error) error {
	return &Error{Err: err, Retryable: true}
}

func Permanent(err error) error {
	return &Error{Err: err, Retryable: false}
}

func IsRetryable(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Retryable
}

func FromStatus(code int, body []byte) error {
	err := fmt.Errorf("unexpected status %d: %s", code, body)

	if code == http.StatusTooManyRequests || code >= 500 {
		return Retryable(err)
	}

	return Permanent(err)
}
