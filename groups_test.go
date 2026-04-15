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
				v1.Get("/health", func(_ http.ResponseWriter, _ *http.Request) {
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
			api.Get("/users", func(_ http.ResponseWriter, _ *http.Request) {})
		})

		r.Get("/root", func(_ http.ResponseWriter, _ *http.Request) {})

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

func TestRouter_Mount(t *testing.T) {
	t.Run("mounted handler receives stripped path", func(t *testing.T) {
		sub := New()
		sub.Get("/users", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r := New()
		r.Mount("/api", sub)

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("parent middleware chain is applied to mounted handler", func(t *testing.T) {
		var mwCalled bool

		sub := New()
		sub.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r := New()
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				mwCalled = true
				next.ServeHTTP(w, req)
			})
		})
		r.Mount("/svc", sub)

		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/svc/hello", nil))

		if !mwCalled {
			t.Error("parent middleware should be applied to the mounted handler")
		}
	})

	t.Run("mount plain http.Handler", func(t *testing.T) {
		plain := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		})

		r := New()
		r.Mount("/plain", plain)

		req := httptest.NewRequest(http.MethodGet, "/plain/anything", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("expected 202, got %d", rec.Code)
		}
	})

	t.Run("mount inside a group inherits prefix", func(t *testing.T) {
		sub := New()
		sub.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r := New()
		r.Group("/v1", func(v1 Router) {
			v1.Mount("/health", sub)
		})

		req := httptest.NewRequest(http.MethodGet, "/v1/health/ping", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("mounted sub-router does not leak routes to parent", func(t *testing.T) {
		sub := New()
		sub.Get("/secret", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r := New()
		r.Mount("/sub", sub)

		// Accessing the route without the mount prefix should 404
		req := httptest.NewRequest(http.MethodGet, "/secret", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for unmounted path, got %d", rec.Code)
		}
	})
}

func TestRouter_With(t *testing.T) {
	t.Run("Create branch with extra middleware without affecting parent", func(t *testing.T) {
		var trace string
		r := New()

		auth := r.With(createTestMiddleware("AUTH", &trace))
		auth.Get("/private", func(_ http.ResponseWriter, _ *http.Request) {})

		r.Get("/public", func(_ http.ResponseWriter, _ *http.Request) {})

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
