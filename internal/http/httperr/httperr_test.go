package httperr_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.jolheiser.com/ugit/internal/http/httperr"
)

func successHandler(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func errorHandler(w http.ResponseWriter, r *http.Request) error {
	return errors.New("test error")
}

func statusErrorHandler(status int) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		return httperr.Status(errors.New("test error"), status)
	}
}

func TestHandler_Success(t *testing.T) {
	handler := httperr.Handler(successHandler)

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestHandler_Error(t *testing.T) {
	handler := httperr.Handler(errorHandler)

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestHandler_StatusError(t *testing.T) {
	testCases := []struct {
		name           string
		status         int
		expectedStatus int
	}{
		{
			name:           "not found",
			status:         http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "bad request",
			status:         http.StatusBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			status:         http.StatusUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := httperr.Handler(statusErrorHandler(tc.status))

			req := httptest.NewRequest("GET", "/", nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)
		})
	}
}

type unwrapper interface {
	Unwrap() error
}

func TestError(t *testing.T) {
	originalErr := errors.New("original error")
	httpErr := httperr.Error(originalErr)

	assert.Equal(t, originalErr.Error(), httpErr.Error())

	unwrapper, ok := any(httpErr).(unwrapper)
	assert.True(t, ok)
	assert.Equal(t, originalErr, unwrapper.Unwrap())
}

func TestStatus(t *testing.T) {
	originalErr := errors.New("original error")
	httpErr := httperr.Status(originalErr, http.StatusNotFound)

	assert.Equal(t, originalErr.Error(), httpErr.Error())

	unwrapper, ok := any(httpErr).(unwrapper)
	assert.True(t, ok)
	assert.Equal(t, originalErr, unwrapper.Unwrap())

	handler := httperr.Handler(func(w http.ResponseWriter, r *http.Request) error {
		return httpErr
	})

	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
