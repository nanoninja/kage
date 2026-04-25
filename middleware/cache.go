// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import "net/http"

// CacheControl sets the Cache-Control header on every response.
//
// Example:
//
//	r.Use(middleware.CacheControl("no-store"))
//	r.Use(middleware.CacheControl("public, max-age=3600"))
func CacheControl(val string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", val)
			next.ServeHTTP(w, r)
		})
	}
}

// NoCache sets headers to prevent the response from being cached
// by browsers or intermediary proxies.
//
// Example:
//
//	r.Use(middleware.NoCache)
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, no-transform, must-revalidate, private, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}
