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

func TestLogger(t *testing.T) {
	t.Run("log successful request", func(t *testing.T) {
		var buf bytes.Buffer
		l := slog.New(slog.NewJSONHandler(&buf, nil))

		mw := Logger(l)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("created"))
		}))

		req := httptest.NewRequest(http.MethodPost, "/users", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		output := buf.String()
		if !strings.Contains(output, `"method":"POST"`) {
			t.Errorf("Expected method POST in logs, got: %s", output)
		}
		if !strings.Contains(output, `"status":201`) {
			t.Errorf("Expected status 201 in logs, got %s", output)
		}
		if !strings.Contains(output, "/users") {
			t.Errorf("Expected path /users in logs, got: %s", output)
		}
	})

	t.Run("default to slog.Default when nil", func(t *testing.T) {
		mw := Logger(nil)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
