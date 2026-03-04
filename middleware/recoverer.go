// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"log/slog"
	"net/http"
)

// Recoverer returns a middleware that recovers from panics.
// If a handler is provided, it will be called to manage the error response.
// Otherwise, it defaults to http.Error with a 500 status.
func Recoverer(l *slog.Logger, handler func(w http.ResponseWriter, r *http.Request, err any)) func(http.Handler) http.Handler {
	if l == nil {
		l = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rerr := recover(); rerr != nil {
					// Log the panic with structured fields
					l.Error("panic recovered",
						slog.Any("error", rerr),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)
					if handler != nil {
						handler(w, r, rerr)
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
