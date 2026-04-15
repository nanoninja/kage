// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultCORSConfig(t *testing.T) {
	t.Run("returns wildcard origin by default", func(t *testing.T) {
		cfg := DefaultCORSConfig()
		if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
			t.Errorf("expected [*], got %v", cfg.AllowedOrigins)
		}
	})

	t.Run("uses provided origins", func(t *testing.T) {
		cfg := DefaultCORSConfig("https://a.com", "https://b.com")
		if len(cfg.AllowedOrigins) != 2 {
			t.Fatalf("expected 2 origins, got %d", len(cfg.AllowedOrigins))
		}
		if cfg.AllowedOrigins[0] != "https://a.com" || cfg.AllowedOrigins[1] != "https://b.com" {
			t.Errorf("unexpected origins: %v", cfg.AllowedOrigins)
		}
	})

	t.Run("has sensible method and header defaults", func(t *testing.T) {
		cfg := DefaultCORSConfig()
		if len(cfg.AllowedMethods) == 0 {
			t.Error("expected default AllowedMethods to be non-empty")
		}
		if len(cfg.AllowedHeaders) == 0 {
			t.Error("expected default AllowedHeaders to be non-empty")
		}
		if cfg.MaxAge != 86400 {
			t.Errorf("expected MaxAge 86400, got %d", cfg.MaxAge)
		}
	})
}

func TestCORS(t *testing.T) {
	cfg := CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	mw := CORS(cfg)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("CORS headers are set on normal request", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		checks := map[string]string{
			"Access-Control-Allow-Origin":      "https://example.com",
			"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE",
			"Access-Control-Allow-Headers":     "Content-Type, Authorization",
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Max-Age":           "3600",
		}
		for header, want := range checks {
			if got := rec.Header().Get(header); got != want {
				t.Errorf("%s: got %q, want %q", header, got, want)
			}
		}
	})

	t.Run("preflight OPTIONS returns 204 and does not call next", func(t *testing.T) {
		var nextCalled bool
		preflightHandler := mw(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			nextCalled = true
		}))

		rec := httptest.NewRecorder()
		preflightHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/", nil))

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", rec.Code)
		}
		if nextCalled {
			t.Error("next handler should not be called on preflight")
		}
	})

	t.Run("normal request calls next handler", func(t *testing.T) {
		var nextCalled bool
		normalHandler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		normalHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if !nextCalled {
			t.Error("next handler should be called on normal request")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("credentials header absent when AllowCredentials is false", func(t *testing.T) {
		noCreds := CORS(CORSConfig{AllowedOrigins: []string{"*"}})
		noCredsHandler := noCreds(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		noCredsHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
			t.Errorf("expected no credentials header, got %q", got)
		}
	})

	t.Run("MaxAge header absent when MaxAge is zero", func(t *testing.T) {
		noMaxAge := CORS(CORSConfig{AllowedOrigins: []string{"*"}})
		noMaxAgeHandler := noMaxAge(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec := httptest.NewRecorder()
		noMaxAgeHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

		if got := rec.Header().Get("Access-Control-Max-Age"); got != "" {
			t.Errorf("expected no max-age header, got %q", got)
		}
	})
}
