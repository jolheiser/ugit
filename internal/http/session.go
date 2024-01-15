package http

import (
	"context"
	"net/http"
)

// Session fulfills git.ReadWriteContexter for an HTTP request
type Session struct {
	w http.ResponseWriter
	r *http.Request
}

// Read implements io.Reader
func (s Session) Read(p []byte) (n int, err error) {
	return s.r.Body.Read(p)
}

// Write implements io.Writer
func (s Session) Write(p []byte) (n int, err error) {
	return s.w.Write(p)
}

// Close implements io.Closer
func (s Session) Close() error {
	return s.r.Body.Close()
}

// Context implements git.ReadWriteContexter
func (s Session) Context() context.Context {
	return s.r.Context()
}
