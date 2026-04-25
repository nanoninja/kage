// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCacheControl(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"no-store", "no-store"},
		{"public max-age", "public, max-age=3600"},
		{"private", "private, max-age=0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CacheControl(tt.value)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

			if got := rec.Header().Get("Cache-Control"); got != tt.value {
				t.Errorf("Cache-Control: got %q, want %q", got, tt.value)
			}
		})
	}

	t.Run("overwrites existing Cache-Control header", func(t *testing.T) {
		handler := CacheControl("no-store")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=3600")
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if got := rec.Header().Get("Cache-Control"); got != "public, max-age=3600" {
			t.Errorf("expected handler to overwrite middleware header, got %q", got)
		}
	})

	t.Run("applied on error responses", func(t *testing.T) {
		handler := CacheControl("no-store")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control on error: got %q, want %q", got, "no-store")
		}
	})
}

func TestNoCache(t *testing.T) {
	t.Run("sets all no-cache headers", func(t *testing.T) {
		handler := NoCache(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		headers := map[string]string{
			"Cache-Control": "no-cache, no-store, no-transform, must-revalidate, private, max-age=0",
			"Pragma":        "no-cache",
			"Expires":       "0",
		}
		for h, want := range headers {
			if got := rec.Header().Get(h); got != want {
				t.Errorf("%s: got %q, want %q", h, got, want)
			}
		}
	})

	t.Run("calls next handler", func(t *testing.T) {
		var called bool
		handler := NoCache(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			called = true
		}))

		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		if !called {
			t.Error("next handler was not called")
		}
	})
}
