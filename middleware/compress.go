// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter wraps http.ResponseWriter to compress the response body.
type compressWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w *compressWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

// Compress returns a middleware that compresses responses using gzip or deflate
// based on the client's Accept-Encoding header.
// If the client does not support compression, the response is passed through unchanged.
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Accept-Encoding")

		switch {
		case strings.Contains(encoding, "gzip"):
			gz := gzip.NewWriter(w)

			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")

			next.ServeHTTP(&compressWriter{ResponseWriter: w, writer: gz}, r)
			_ = gz.Close()

		case strings.Contains(encoding, "deflate"):
			fl, _ := flate.NewWriter(w, flate.DefaultCompression)

			w.Header().Set("Content-Encoding", "deflate")
			w.Header().Del("Content-Length")

			next.ServeHTTP(&compressWriter{ResponseWriter: w, writer: fl}, r)
			_ = fl.Close()

		default:
			next.ServeHTTP(w, r)
		}
	})
}
