// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_Route(t *testing.T) {
	t.Run("registers multiple methods on the same path", func(t *testing.T) {
		r := New()

		r.Route("/users/{id}", func(rt Route) {
			rt.Get(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			rt.Put(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			})
			rt.Delete(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
		})

		tests := []struct {
			method string
			status int
		}{
			{http.MethodGet, http.StatusOK},
			{http.MethodPut, http.StatusAccepted},
			{http.MethodDelete, http.StatusNoContent},
		}

		for _, tt := range tests {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(tt.method, "/users/42", nil))

			if rec.Code != tt.status {
				t.Errorf("%s /users/42: expected %d, got %d", tt.method, tt.status, rec.Code)
			}
		}
	})

	t.Run("path param is accessible in handler", func(t *testing.T) {
		r := New()
		var capturedID string

		r.Route("/items/{id}", func(rt Route) {
			rt.Get(func(_ http.ResponseWriter, r *http.Request) {
				capturedID = Param(r, "id")
			})
		})

		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items/99", nil))

		if capturedID != "99" {
			t.Errorf("expected id=99, got %q", capturedID)
		}
	})

	t.Run("inherits parent middleware chain", func(t *testing.T) {
		var mwCalled bool
		r := New()

		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				mwCalled = true
				next.ServeHTTP(w, req)
			})
		})

		r.Route("/ping", func(rt Route) {
			rt.Get(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})

		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ping", nil))

		if !mwCalled {
			t.Error("parent middleware should be applied to Route handlers")
		}
	})

	t.Run("route inside a group inherits prefix", func(t *testing.T) {
		r := New()
		var called bool

		r.Group("/api", func(api Router) {
			api.Route("/status", func(rt Route) {
				rt.Get(func(_ http.ResponseWriter, _ *http.Request) {
					called = true
				})
			})
		})

		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/status", nil))

		if !called {
			t.Error("GET /api/status should have been reached via Route inside Group")
		}
	})

	t.Run("all HTTP methods are reachable", func(t *testing.T) {
		r := New()

		r.Route("/resource", func(rt Route) {
			rt.Connect(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			rt.Head(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			rt.Options(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			rt.Patch(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			rt.Post(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
			rt.Trace(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		})

		methods := []string{
			http.MethodConnect,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPatch,
			http.MethodPost,
			http.MethodTrace,
		}

		for _, method := range methods {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(method, "/resource", nil))

			if rec.Code != http.StatusOK {
				t.Errorf("%s /resource: expected 200, got %d", method, rec.Code)
			}
		}
	})
}
