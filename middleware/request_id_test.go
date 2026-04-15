// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	t.Run("generates an ID when none is provided", func(t *testing.T) {
		handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		id := rec.Header().Get("X-Request-ID")
		if id == "" {
			t.Error("expected X-Request-ID header to be set")
		}
		if len(id) != 32 {
			t.Errorf("expected 32-char hex ID, got %q (len=%d)", id, len(id))
		}
	})

	t.Run("reuses existing X-Request-ID from request", func(t *testing.T) {
		existing := "my-custom-request-id"
		handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", existing)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("X-Request-ID"); got != existing {
			t.Errorf("expected %q, got %q", existing, got)
		}
	})

	t.Run("stores ID in request context", func(t *testing.T) {
		var capturedID string
		handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			capturedID = GetRequestID(r)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)

		if capturedID == "" {
			t.Error("expected request ID in context, got empty string")
		}
	})

	t.Run("context ID matches response header", func(t *testing.T) {
		var contextID string
		handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			contextID = GetRequestID(r)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if contextID != rec.Header().Get("X-Request-ID") {
			t.Errorf("context ID %q does not match header %q", contextID, rec.Header().Get("X-Request-ID"))
		}
	})

	t.Run("each request gets a unique ID", func(t *testing.T) {
		handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))

		rec1 := httptest.NewRecorder()
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))
		handler.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))

		id1 := rec1.Header().Get("X-Request-ID")
		id2 := rec2.Header().Get("X-Request-ID")
		if id1 == id2 {
			t.Errorf("expected unique IDs, got identical: %q", id1)
		}
	})
}

func TestGetRequestID(t *testing.T) {
	t.Run("returns empty string when no ID in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if id := GetRequestID(req); id != "" {
			t.Errorf("expected empty string, got %q", id)
		}
	})
}
