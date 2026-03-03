// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper to track middleware execution order
func createTestMiddleware(tag string, trace *string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*trace += tag
			next.ServeHTTP(w, r)
		})
	}
}

func TestRouter_Group(t *testing.T) {
	t.Run("Nested groups and prefix concatenation", func(t *testing.T) {
		r := New()
		var called bool

		r.Group("/api", func(api Router) {
			api.Group("/v1/", func(v1 Router) { // Testing trailing slash normalization too
				v1.Get("/health", func(w http.ResponseWriter, r *http.Request) {
					called = true
				})
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		r.ServeHTTP(httptest.NewRecorder(), req)

		if !called {
			t.Error("Nested group route /api/v1/health was not reached")
		}
	})

	t.Run("Middleware isolation between parent and groups", func(t *testing.T) {
		var trace string
		r := New()
		r.Use(createTestMiddleware("A", &trace))

		r.Group("/api", func(api Router) {
			api.Use(createTestMiddleware("B", &trace))
			api.Get("/users", func(w http.ResponseWriter, r *http.Request) {})
		})

		r.Get("/root", func(w http.ResponseWriter, r *http.Request) {})

		// Test root: Should only execute "A"
		trace = ""
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/root", nil))
		if trace != "A" {
			t.Errorf("Root trace: got %q, want %q (parent polluted by group)", trace, "A")
		}

		// Test group: Should execute "AB"
		trace = ""
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/api/users", nil))
		if trace != "AB" {
			t.Errorf("Group trace: got %q, want %q", trace, "AB")
		}
	})

	t.Run("Sibling group isolation", func(t *testing.T) {
		r := New()

		r.Group("/a", func(a Router) {
			a.Use(func(next http.Handler) http.Handler { return next })
		})

		r.Group("/b", func(b Router) {
			if length := len(b.(*router).middlewares); length != 0 {
				t.Errorf("Group B should be empty, but inherited MW from sibling Group A (got %d)", length)
			}
		})
	})
}

func TestRouter_With(t *testing.T) {
	t.Run("Create branch with extra middleware without affecting parent", func(t *testing.T) {
		var trace string
		r := New()

		auth := r.With(createTestMiddleware("AUTH", &trace))
		auth.Get("/private", func(w http.ResponseWriter, r *http.Request) {})

		r.Get("/public", func(w http.ResponseWriter, r *http.Request) {})

		// Test public: trace should remain empty
		trace = ""
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/public", nil))
		if trace != "" {
			t.Errorf("Public trace: got %q, want empty", trace)
		}

		// Test private: trace should have "AUTH"
		trace = ""
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/private", nil))
		if trace != "AUTH" {
			t.Errorf("Private trace: got %q, want %q", trace, "AUTH")
		}
	})
}

func TestRouter_Clone(t *testing.T) {
	t.Run("Perform deep copy of middleware slice", func(t *testing.T) {
		r1 := New().(*router)
		r1.Use(func(h http.Handler) http.Handler { return h })

		r2 := r1.clone()
		r2.Use(func(h http.Handler) http.Handler { return h })

		if len(r1.middlewares) != 1 {
			t.Errorf("Original router polluted: got %d middlewares, want 1", len(r1.middlewares))
		}

		if len(r2.middlewares) != 2 {
			t.Errorf("Cloned router failed to add middleware: got %d, want 2", len(r2.middlewares))
		}

		if r1.prefix != r2.prefix {
			t.Errorf("Prefix mismatch: expected %q, got %q", r1.prefix, r2.prefix)
		}
	})
}
