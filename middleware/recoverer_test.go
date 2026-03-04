// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoverer(t *testing.T) {
	t.Run("recover from panic and log error", func(t *testing.T) {
		var buf bytes.Buffer
		l := slog.New(slog.NewTextHandler(&buf, nil))

		mw := Recoverer(l, nil)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("unexpected crash")
		}))

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", rec.Code)
		}
		if !strings.Contains(buf.String(), "unexpected crash") {
			t.Errorf("Expected error message in logs, got: %s", buf.String())
		}
	})

	t.Run("custom recovery handler", func(t *testing.T) {
		customCalled := false
		handlerFunc := func(w http.ResponseWriter, r *http.Request, err any) {
			customCalled = true
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("service down"))
		}

		mw := Recoverer(nil, handlerFunc)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("boom")
		}))

		req := httptest.NewRequest(http.MethodGet, "/crash", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if !customCalled {
			t.Errorf("Custom recovery handler was not invoked")
		}
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 503, got %d", rec.Code)
		}
	})
}
