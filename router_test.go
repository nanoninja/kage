// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestRouter_Options(t *testing.T) {
	t.Run("WithPrefix sets initial prefix", func(t *testing.T) {
		prefix := "/api/v1"
		r := New(WithPrefix(prefix)).(*router)

		if r.prefix != prefix {
			t.Errorf("Expected prefix %s, got %s", prefix, r.prefix)
		}
	})

	t.Run("WithMux sets custom ServeMux", func(t *testing.T) {
		customMux := http.NewServeMux()
		r := New(WithMux(customMux)).(*router)

		if r.mux != customMux {
			t.Error("Router should use the custom mux provided in options")
		}
	})

	t.Run("WithNotFound registers custom 404 handler", func(t *testing.T) {
		status := http.StatusTeapot
		custom404 := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		}

		r := New(WithNotFound(custom404))

		// Test by performing a request to a non-existent route
		req := httptest.NewRequest(http.MethodGet, "/unknown-route", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != status {
			t.Errorf("WithNotFound failed: expected status %d, got %d", status, rec.Code)
		}
	})

	t.Run("Multiple options together", func(t *testing.T) {
		prefix := "/global"
		customMux := http.NewServeMux()

		r := New(
			WithPrefix(prefix),
			WithMux(customMux),
		).(*router)

		if r.prefix != prefix || r.mux != customMux {
			t.Errorf("Options combination failed: got prefix %s and mux %p", r.prefix, r.mux)
		}
	})

	t.Run("Nil safety for options", func(t *testing.T) {
		// Test nil for Mux
		r := New(WithMux(nil)).(*router)
		if r.mux == nil {
			t.Error("WithMux(nil) should result in default mux")
		}

		// Test nil for NotFound (should not panic)
		r2 := New(WithNotFound(nil))
		req := httptest.NewRequest(http.MethodGet, "/404", nil)
		rec := httptest.NewRecorder()

		// Should not crash, should return Go's default 404
		r2.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("WithNotFound(nil) should fallback to default 404, got %d", rec.Code)
		}
	})
}

func TestRouter_MiddlewareChain(t *testing.T) {
	r := New()
	var trace []string

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			trace = append(trace, "global")
			next.ServeHTTP(w, req)
		})
	})

	r.Group("/api", func(api Router) {
		api.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				trace = append(trace, "group")
				next.ServeHTTP(w, req)
			})
		})

		mwWith := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				trace = append(trace, "with")
				next.ServeHTTP(w, req)
			})
		}

		api.With(mwWith).Get("/users", func(_ http.ResponseWriter, _ *http.Request) {
			trace = append(trace, "handler")
		})
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	expected := []string{"global", "group", "with", "handler"}
	if strings.Join(trace, ">") != strings.Join(expected, ">") {
		t.Errorf("Order mismatch\nwant: %v\ngot:  %v", expected, trace)
	}
}

func TestRouter_MiddlewareOrder(t *testing.T) {
	r := New()
	var trace []string

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			trace = append(trace, "M1_In")
			next.ServeHTTP(w, req)
			trace = append(trace, "M1_Out")
		})
	})

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			trace = append(trace, "M2_In")
			next.ServeHTTP(w, req)
			trace = append(trace, "M2_Out")
		})
	})

	r.Get("/test", func(_ http.ResponseWriter, _ *http.Request) {
		trace = append(trace, "Handler")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	expected := []string{"M1_In", "M2_In", "Handler", "M2_Out", "M1_Out"}

	if len(trace) != len(expected) {
		t.Fatalf("Expected trace length %d, got %d", len(expected), len(trace))
	}

	for i, v := range trace {
		if v != expected[i] {
			t.Errorf("At index %d: expected %s, got %s", i, expected[i], v)
		}
	}
}

func TestStatic(t *testing.T) {
	mockFS := fstest.MapFS{
		"test.txt":      {Data: []byte("hello world")},
		"css/style.css": {Data: []byte("body {}")},
	}

	r := New()

	r.Group("/api", func(api Router) {
		api.StaticFS("/assets", http.FS(mockFS))
	})

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "File at root of static",
			url:            "/api/assets/test.txt",
			expectedStatus: http.StatusOK,
			expectedBody:   "hello world",
		},
		{
			name:           "File in subdirectory",
			url:            "/api/assets/css/style.css",
			expectedStatus: http.StatusOK,
			expectedBody:   "body {}",
		},
		{
			name:           "Non-existent file",
			url:            "/api/assets/404.txt",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			res := rec.Result()
			defer func() { _ = res.Body.Close() }()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			body, _ := io.ReadAll(res.Body)
			if string(body) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
			}
		})
	}
}

func TestStatic_LocalFolder(t *testing.T) {
	r := New()

	t.Run("Should not panic on registration", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Static panicked: %v", r)
			}
		}()
		r.Static("/tmp-test", "./")
	})
}
