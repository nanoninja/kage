// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var nopHandler = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

func nopMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func BenchmarkRouter_SimpleRoute(b *testing.B) {
	r := New()
	r.Get("/hello", nopHandler)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	for b.Loop() {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_ParamRoute(b *testing.B) {
	r := New()
	r.Get("/users/{id}", nopHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	rec := httptest.NewRecorder()

	for b.Loop() {
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

	for b.Loop() {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_MiddlewareChain(b *testing.B) {
	r := New()
	r.Use(nopMiddleware, nopMiddleware, nopMiddleware)
	r.Get("/hello", nopHandler)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	for b.Loop() {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_With(b *testing.B) {
	r := New()
	r.With(nopMiddleware, nopMiddleware).Get("/hello", nopHandler)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	for b.Loop() {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_Mount(b *testing.B) {
	sub := New()
	sub.Get("/status", nopHandler)

	r := New()
	r.Mount("/api", sub)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()

	for b.Loop() {
		r.ServeHTTP(rec, req)
	}
}

func BenchmarkRouter_RouteMultiMethod(b *testing.B) {
	r := New()
	r.Route("/items", func(rt Route) {
		rt.Get(nopHandler)
		rt.Post(nopHandler)
	})

	reqGet := httptest.NewRequest(http.MethodGet, "/items", nil)
	reqPost := httptest.NewRequest(http.MethodPost, "/items", nil)
	rec := httptest.NewRecorder()

	b.ResetTimer()
	for b.Loop() {
		r.ServeHTTP(rec, reqGet)
		r.ServeHTTP(rec, reqPost)
	}
}

func BenchmarkRouter_Routes(b *testing.B) {
	r := New()
	r.Get("/a", nopHandler)
	r.Post("/b", nopHandler)
	r.Put("/c/{id}", nopHandler)
	r.Delete("/d/{id}", nopHandler)

	for b.Loop() {
		_ = r.Routes()
	}
}

func BenchmarkRouter_Parallel(b *testing.B) {
	r := New()
	r.Use(nopMiddleware)
	r.Get("/hello", nopHandler)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)

	b.RunParallel(func(pb *testing.PB) {
		rec := httptest.NewRecorder()
		for pb.Next() {
			r.ServeHTTP(rec, req)
		}
	})
}
