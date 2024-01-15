package httperr

import (
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
)

type httpError struct {
	err    error
	status int
}

func (h httpError) Error() string {
	return h.err.Error()
}

func (h httpError) Unwrap() error {
	return h.err
}

func Error(err error) httpError {
	return Status(err, http.StatusInternalServerError)
}

func Status(err error, status int) httpError {
	return httpError{err: err, status: status}
}

func Handler(fn func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			status := http.StatusInternalServerError

			var httpErr httpError
			if errors.As(err, &httpErr) {
				status = httpErr.status
			}

			log.Error(err)
			http.Error(w, http.StatusText(status), status)
		}
	}
}
