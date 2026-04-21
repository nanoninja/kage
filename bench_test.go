// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var nopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func BenchmarkRouter_SimpleRoute(b *testing.B) {
	r := New()
	r.Get("/hello", nopHandler)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_ParamRoute(b *testing.B) {
	r := New()
	r.Get("/users/{id}", nopHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_NestedGroup(b *testing.B) {
	r := New()
	r.Group("/api", func(api Router) {
		api.Group("/v1", func(v1 Router) {
			v1.Get("/users/{id}", nopHandler)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/42", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_MiddlewareChain(b *testing.B) {
	nop := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	r := New()
	r.Use(nop, nop, nop)
	r.Get("/hello", nopHandler)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		r.ServeHTTP(rec, req)
	}
}
