// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParam(t *testing.T) {
	t.Run("extract path value from request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/1230", nil)

		req.SetPathValue("id", "123")
		req.SetPathValue("slug", "john-doe")

		tests := []struct {
			key      string
			expected string
		}{
			{"id", "123"},
			{"slug", "john-doe"},
			{"unknown", ""},
		}

		for _, tt := range tests {
			if val := Param(req, tt.key); val != tt.expected {
				t.Errorf("Param(%q): got %q, want %q", tt.key, val, tt.expected)
			}
		}
	})
}

func TestRouter_Chain(t *testing.T) {
	t.Run("execute middleware in FIFO order", func(t *testing.T) {
		var result string

		mw := func(tag string) func(http.Handler) http.Handler {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					result += tag
					next.ServeHTTP(w, r)
				})
			}
		}

		r := &router{
			middlewares: []func(http.Handler) http.Handler{
				mw("1"),
				mw("2"),
				mw("3"),
			},
		}

		handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			result += "H"
		})

		chained := r.chain(handler)
		chained.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

		if expected := "123H"; result != expected {
			t.Errorf("Middleware chain order: got %q, want %q", result, expected)
		}
	})
}

func TestRouter_WrapPath(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		pattern  string
		expected string
	}{
		// Cas sans préfixe
		{"no prefix, empty pattern", "", "", "/"},
		{"no prefix, root pattern", "", "/", "/"},
		{"no prefix, simple pattern", "", "users", "/users"},
		{"no prefix, strict root", "", "/{$}", "/{$}"},

		// Cas avec préfixe (Join classique)
		{"prefix and simple pattern", "/api", "users", "/api/users"},
		{"prefix with slash and pattern", "/api/", "users", "/api/users"},
		{"prefix and pattern with slash", "/api", "/users", "/api/users"},

		// Cas du Subtree Routing (La zone critique)
		{"preserve trailing slash", "/api", "users/", "/api/users/"},
		{"preserve trailing slash with prefix slash", "/api/", "users/", "/api/users/"},
		{"root pattern with prefix", "/api", "/", "/api/"},

		// Cas spécifiques Go 1.22
		{"strict root with prefix", "/api", "/{$}", "/api/{$}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &router{prefix: tt.prefix}
			result := r.wrapPath(tt.pattern)

			if result != tt.expected {
				t.Errorf("wrapPath(%q) with prefix %q: got %q, want %q",
					tt.pattern, tt.prefix, result, tt.expected)
			}
		})
	}
}
