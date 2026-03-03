// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"path"
	"strings"
)

// Param returns the path value for the given key from the http.Request.
// It is a convenient wrapper around r.PathValue.
func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}

// chain wraps the given http.Handler with all registered middlewares
// in the correct order (FIFO).
func (r *router) chain(h http.Handler) http.Handler {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

// wrapPath combines the router's prefix with the given pattern.
// It ensures the path is clean, starts with a slash, and preserves
// the trailing slash if the pattern specifically requested it.
func (r *router) wrapPath(pattern string) string {
	p := path.Join(r.prefix, pattern)

	if strings.HasSuffix(pattern, "/") && !strings.HasSuffix(p, "/") {
		p += "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	return p
}
