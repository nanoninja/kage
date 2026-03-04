// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import "net/http"

// WrapResponseWriter is a wrapper around http.WrapResponseWriter that captures
// the HTTP status code and tracks if headers have been written.
type WrapResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// NewWrapResponseWriter creates a new responseWriter with a default status of 200 OK.
func NewWrapResponseWriter(w http.ResponseWriter) *WrapResponseWriter {
	return &WrapResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// Status returns the captured HTTP status code.
func (rw *WrapResponseWriter) Status() int {
	return rw.status
}

// Written returns true if the HTTP response headers have been sent.
func (rw *WrapResponseWriter) Written() bool {
	return rw.wroteHeader
}

// WriteHeader captures the status code and delegates to the underlying ResponseWriter.
func (rw *WrapResponseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write ensures that WriteHeader is called with http.StatusOK if it hasn't been called yet.
func (rw *WrapResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter.
// This is essential for http.ResponseController to access advanced features
// like Hijack or Flush through the wrapper.
func (rw *WrapResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
