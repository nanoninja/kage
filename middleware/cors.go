// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig holds the configuration for the CORS middleware.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a CORSConfig with permissive defaults suitable for development.
// Specific origins can be provided to restrict access in production.
//
// Example:
//
//	middleware.CORS(middleware.DefaultCORSConfig())
//	middleware.CORS(middleware.DefaultCORSConfig("https://myapp.com"))
func DefaultCORSConfig(origins ...string) CORSConfig {
	allowedOrigins := []string{"*"}
	if len(origins) > 0 {
		allowedOrigins = origins
	}
	return CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Authorization"},
		MaxAge:         86400,
	}
}

// CORS returns a middleware that adds Cross-Origin Resource Sharing headers
// to every response based on the provided configuration.
//
// Preflight requests (OPTIONS) are handled automatically and short-circuited
// with a 204 No Content response.
//
// Example:
//
//	r.Use(middleware.CORS(middleware.DefaultCORSConfig()))
//
//	// With specific origins:
//	r.Use(middleware.CORS(middleware.DefaultCORSConfig("https://myapp.com")))
//
//	// Fully custom:
//	r.Use(middleware.CORS(middleware.CORSConfig{
//	    AllowedOrigins: []string{"https://example.com"},
//	    AllowedMethods: []string{"GET", "POST"},
//	    AllowedHeaders: []string{"Content-Type", "Authorization"},
//	}))
func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	origins := strings.Join(cfg.AllowedOrigins, ", ")
	methods := strings.Join(cfg.AllowedMethods, ", ")
	headers := strings.Join(cfg.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
