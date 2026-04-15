// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout returns a middleware that cancels the request context after the given duration.
//
// If the timeout expires before the handler writes a response, onFailure is called
// to write the error response. If onFailure is nil, it defaults to 503 Service Unavailable.
//
// If the handler has already started writing a response before the deadline is reached,
// the error handler is not called to avoid a duplicate response.
//
// Example:
//
//	r.Use(middleware.Timeout(5*time.Second, nil))
//
//	// With a custom JSON error:
//	r.Use(middleware.Timeout(5*time.Second, func(w http.ResponseWriter, r *http.Request) {
//	    w.Header().Set("Content-Type", "application/json")
//	    w.WriteHeader(http.StatusServiceUnavailable)
//	    _, _ = w.Write([]byte(`{"error":"request timeout"}`))
//	}))
func Timeout(duration time.Duration, onFailure func(http.ResponseWriter, *http.Request)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			ww := NewWrapResponseWriter(w)
			next.ServeHTTP(ww, r.WithContext(ctx))

			if ctx.Err() == context.DeadlineExceeded && !ww.Written() {
				if onFailure != nil {
					onFailure(w, r)
					return
				}
				http.Error(w,
					http.StatusText(http.StatusServiceUnavailable),
					http.StatusServiceUnavailable,
				)
			}
		})
	}
}
