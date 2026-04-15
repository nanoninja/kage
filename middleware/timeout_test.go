// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	t.Run("fast handler completes before timeout", func(t *testing.T) {
		mw := Timeout(100*time.Millisecond, nil)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("slow handler triggers default 503", func(t *testing.T) {
		mw := Timeout(1*time.Millisecond, nil)
		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			// Simulate a slow handler by waiting for context cancellation
			<-r.Context().Done()
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("expected 503, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), http.StatusText(http.StatusServiceUnavailable)) {
			t.Errorf("expected 503 body, got %q", rec.Body.String())
		}
	})

	t.Run("slow handler triggers custom onFailure", func(t *testing.T) {
		var onFailureCalled bool

		onFailure := func(w http.ResponseWriter, _ *http.Request) {
			onFailureCalled = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"timeout"}`))
		}

		mw := Timeout(1*time.Millisecond, onFailure)
		handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !onFailureCalled {
			t.Error("onFailure should have been called on timeout")
		}
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("expected 503, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"error":"timeout"`) {
			t.Errorf("expected JSON body, got %q", rec.Body.String())
		}
	})

	t.Run("handler already written before deadline — no double write", func(t *testing.T) {
		mw := Timeout(50*time.Millisecond, nil)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Body.String() != "ok" {
			t.Errorf("expected body %q, got %q", "ok", rec.Body.String())
		}
	})

	t.Run("context deadline is propagated to handler", func(t *testing.T) {
		mw := Timeout(100*time.Millisecond, nil)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := r.Context().Deadline(); !ok {
				t.Error("expected a deadline on the request context")
			}
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

}
