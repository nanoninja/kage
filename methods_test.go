// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_Handle(t *testing.T) {
	t.Run("register with Handle and specific method pattern", func(t *testing.T) {
		r := New()

		r.Handle("POST /custom", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))

		req := httptest.NewRequest(http.MethodPost, "/custom", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("Handle failed: got %d, want 201", rec.Code)
		}
	})
}

func TestRouter_GenericMethods(t *testing.T) {
	r := New()

	t.Run("Method", func(t *testing.T) {
		r.Method("PURGE", "/cache", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(204)
		}))

		req := httptest.NewRequest("PURGE", "/cache", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != 204 {
			t.Errorf("Method failed: got %d", rec.Code)
		}
	})

	t.Run("MethodFunc", func(t *testing.T) {
		r.MethodFunc("PROFIND", "webdav", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(207)
		})

		req := httptest.NewRequest("PROFIND", "/webdav", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != 207 {
			t.Errorf("MethodFunc failed: got %d", rec.Code)
		}
	})
}

func TestRouter_Methods(t *testing.T) {
	r := New()

	// Register a route for each HTTP method
	// Each handler returns a unique status code to prove it was called correctly
	tests := []struct {
		name   string
		method string
		action func(pattern string, h http.HandlerFunc)
		status int
	}{
		{"CONNECT", http.MethodConnect, r.Connect, 201},
		{"DELETE", http.MethodDelete, r.Delete, 202},
		{"GET", http.MethodGet, r.Get, 203},
		{"HEAD", http.MethodHead, r.Head, 204},
		{"OPTIONS", http.MethodOptions, r.Options, 205},
		{"PATCH", http.MethodPatch, r.Patch, 206},
		{"POST", http.MethodPost, r.Post, 207},
		{"PUT", http.MethodPut, r.Put, 208},
		{"TRACE", http.MethodTrace, r.Trace, 209},

		// Test HandleFunc variant (which calls Handle internally)
		{"HandleFunc", http.MethodGet, r.HandleFunc, 210},
	}

	for _, tt := range tests {
		// Create a unique path for each case
		path := "/" + tt.name

		tt.action(path, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tt.status)
		})

		t.Run("method"+tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.status {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.status, rec.Code)
			}
		})
	}
}

func TestRouter_NotFound(t *testing.T) {
	t.Run("custom not found handler", func(t *testing.T) {
		r := New()

		r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusTeapot {
			t.Errorf("Expected 418 for custom NotFound, got %d", rec.Code)
		}
	})

	t.Run("middleware chain is applied to NotFound handler", func(t *testing.T) {
		r := New()
		var mwCalled bool

		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				mwCalled = true
				next.ServeHTTP(w, req)
			})
		})

		r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil))

		if !mwCalled {
			t.Error("NotFound handler should pass through the middleware chain")
		}
	})

	t.Run("calling NotFound twice does not panic", func(t *testing.T) {
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("NotFound called twice caused a panic: %v", rec)
			}
		}()

		r := New()
		r.NotFound(func(_ http.ResponseWriter, _ *http.Request) {})
		r.NotFound(func(_ http.ResponseWriter, _ *http.Request) {}) // second call: silently ignored
	})
}

func TestRouter_Routes(t *testing.T) {
	t.Run("returns registered routes", func(t *testing.T) {
		r := New()
		r.Get("/users", func(_ http.ResponseWriter, _ *http.Request) {})
		r.Post("/users", func(_ http.ResponseWriter, _ *http.Request) {})
		r.Delete("/users/{id}", func(_ http.ResponseWriter, _ *http.Request) {})

		routes := r.Routes()
		if len(routes) != 3 {
			t.Fatalf("expected 3 routes, got %d", len(routes))
		}
	})

	t.Run("route info contains correct method and pattern", func(t *testing.T) {
		r := New(WithPrefix("/api"))
		r.Get("/users", func(_ http.ResponseWriter, _ *http.Request) {})

		routes := r.Routes()
		if len(routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(routes))
		}
		if routes[0].Method != http.MethodGet {
			t.Errorf("expected method GET, got %q", routes[0].Method)
		}
		if routes[0].Pattern != "/api/users" {
			t.Errorf("expected pattern /api/users, got %q", routes[0].Pattern)
		}
	})

	t.Run("group routes are visible from parent", func(t *testing.T) {
		r := New()
		r.Group("/api", func(api Router) {
			api.Get("/status", func(_ http.ResponseWriter, _ *http.Request) {})
			api.Post("/users", func(_ http.ResponseWriter, _ *http.Request) {})
		})

		routes := r.Routes()
		if len(routes) != 2 {
			t.Fatalf("expected 2 routes from group, got %d", len(routes))
		}
	})

	t.Run("empty router returns empty slice", func(t *testing.T) {
		r := New()
		if routes := r.Routes(); len(routes) != 0 {
			t.Errorf("expected empty routes, got %d", len(routes))
		}
	})
}
