// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"log/slog"
	"net/http"
)

// Recoverer returns a middleware that recovers from panics and logs the error.
//
// If onFailure is provided, it is called to write the error response, giving
// full control over the format (e.g. JSON). The recovered panic value is passed
// as the third argument. If onFailure is nil, it defaults to 500 Internal Server Error.
//
// Example:
//
//	r.Use(middleware.Recoverer(nil, nil))
//
//	// With a custom JSON error:
//	r.Use(middleware.Recoverer(logger, func(w http.ResponseWriter, r *http.Request, err any) {
//	    w.Header().Set("Content-Type", "application/json")
//	    w.WriteHeader(http.StatusInternalServerError)
//	    _, _ = fmt.Fprintf(w, `{"error":"%v"}`, err)
//	}))
func Recoverer(l *slog.Logger, onFailure func(http.ResponseWriter, *http.Request, any)) func(http.Handler) http.Handler {
	if l == nil {
		l = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rerr := recover(); rerr != nil {
					l.Error("panic recovered",
						slog.Any("error", rerr),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)
					if onFailure != nil {
						onFailure(w, r, rerr)
						return
					}
					http.Error(w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
