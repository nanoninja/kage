// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logger returns a middleware that logs request details using structured logging.
// If l is nil, it uses the default slog logger.
func Logger(l *slog.Logger) func(http.Handler) http.Handler {
	if l == nil {
		l = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := NewWrapResponseWriter(w)

			next.ServeHTTP(ww, r)

			l.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
