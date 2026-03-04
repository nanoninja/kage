// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrapResponseWriter(t *testing.T) {
	tests := []struct {
		name            string
		action          func(rw *WrapResponseWriter)
		expectedStatus  int
		expectedWritten bool
	}{
		{
			name: "capture explicit status code",
			action: func(rw *WrapResponseWriter) {
				rw.WriteHeader(http.StatusCreated)
			},
			expectedStatus:  http.StatusCreated,
			expectedWritten: true,
		},
		{
			name: "default status code is 200 on write",
			action: func(rw *WrapResponseWriter) {
				_, err := rw.Write([]byte("hello"))
				if err != nil {
					t.Fatalf("Failed to write: %v", err)
				}
			},
			expectedStatus:  http.StatusOK,
			expectedWritten: true,
		},
		{
			name: "written returns true after writeheader",
			action: func(rw *WrapResponseWriter) {
				rw.WriteHeader(http.StatusNoContent)
			},
			expectedStatus:  http.StatusNoContent,
			expectedWritten: true,
		},
		{
			name: "writeheader should only record the first call",
			action: func(rw *WrapResponseWriter) {
				rw.WriteHeader(http.StatusAccepted)
				rw.WriteHeader(http.StatusBadRequest)
			},
			expectedStatus:  http.StatusAccepted,
			expectedWritten: true,
		},
		{
			name: "capture headers correctly",
			action: func(rw *WrapResponseWriter) {
				rw.Unwrap().Header().Set("X-Nanoninja", "Test")
				rw.WriteHeader(http.StatusOK)
			},
			expectedStatus:  http.StatusOK,
			expectedWritten: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			rw := NewWrapResponseWriter(rec)

			tt.action(rw)

			if rw.Status() != tt.expectedStatus {
				t.Errorf("Status(): got %d, want %d", rw.Status(), tt.expectedStatus)
			}

			if rw.Written() != tt.expectedWritten {
				t.Errorf("Written(): got %v, want %v", rw.Written(), tt.expectedWritten)
			}
		})
	}
}

func TestResponseWriter_New(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := NewWrapResponseWriter(rec)

	if rw.Status() != http.StatusOK {
		t.Errorf("Initial Status(): got %d, want %d", rw.Status(), http.StatusOK)
	}

	if rw.Written() {
		t.Error("Initial Written(): got true, want false")
	}

	if rec.Body.Len() != 0 {
		t.Errorf("Initial Body: expected empty, got %d bytes", rec.Body.Len())
	}
}

func TestResponseWriter_Unwrap(t *testing.T) {
	t.Run("transparency for ResponseController", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := NewWrapResponseWriter(rec)

		rc := http.NewResponseController(rw)
		err := rc.Flush()

		if err != nil {
			t.Errorf("ResponseController.Flush() failed through wrapper: %v", err)
		}
	})
}
