// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package middleware

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const compressBody = "Hello, compressed world!"

func compressHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(compressBody))
}

func TestCompress(t *testing.T) {
	handler := Compress(http.HandlerFunc(compressHandler))

	t.Run("gzip encoding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
			t.Errorf("expected Content-Encoding: gzip, got %q", got)
		}
		if rec.Header().Get("Content-Length") != "" {
			t.Error("Content-Length should be removed when compressing")
		}

		gr, err := gzip.NewReader(rec.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer func() { _ = gr.Close() }()

		body, _ := io.ReadAll(gr)
		if string(body) != compressBody {
			t.Errorf("expected %q, got %q", compressBody, string(body))
		}
	})

	t.Run("deflate encoding", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "deflate")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Content-Encoding"); got != "deflate" {
			t.Errorf("expected Content-Encoding: deflate, got %q", got)
		}
		if rec.Header().Get("Content-Length") != "" {
			t.Error("Content-Length should be removed when compressing")
		}

		fr := flate.NewReader(rec.Body)
		defer func() { _ = fr.Close() }()

		body, _ := io.ReadAll(fr)
		if string(body) != compressBody {
			t.Errorf("expected %q, got %q", compressBody, string(body))
		}
	})

	t.Run("no encoding passes through unchanged", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Content-Encoding"); got != "" {
			t.Errorf("expected no Content-Encoding, got %q", got)
		}
		if body := rec.Body.String(); body != compressBody {
			t.Errorf("expected %q, got %q", compressBody, body)
		}
	})

	t.Run("gzip takes priority over deflate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip, deflate")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
			t.Errorf("expected gzip to take priority, got %q", got)
		}
	})

	t.Run("deflate writer error falls back to uncompressed", func(t *testing.T) {
		orig := flateNewWriter
		flateNewWriter = func(_ io.Writer, level int) (*flate.Writer, error) {
			return nil, errors.New("forced error")
		}
		defer func() { flateNewWriter = orig }()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "deflate")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if got := rec.Header().Get("Content-Encoding"); got != "" {
			t.Errorf("expected no Content-Encoding on error, got %q", got)
		}
		if body := rec.Body.String(); body != compressBody {
			t.Errorf("expected uncompressed body %q, got %q", compressBody, body)
		}
	})
}
