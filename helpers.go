// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"path"
	"strings"
)

// Param returns the value of the named path parameter from the request.
// It is a convenience wrapper around r.PathValue, introduced in Go 1.22.
//
// Example:
//
//	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
//	    id := kage.Param(r, "id")
//	    fmt.Fprintf(w, "User: %s", id)
//	})
func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}

// FileServer returns an http.Handler that serves static files
// from the given local directory. Designed to be used with Mount.
//
// Example:
//
//	r.Mount("/assets", kage.FileServer("./public"))
func FileServer(root string) http.Handler {
	return http.FileServer(http.Dir(root))
}

// FileServerFS returns an http.Handler that serves static files
// from the given http.FileSystem. Useful with Go's embed package.
//
// Example:
//
//	r.Mount("/assets", kage.FileServerFS(http.FS(embeddedFiles)))
func FileServerFS(fsys http.FileSystem) http.Handler {
	return http.FileServer(fsys)
}

// Redirect returns an http.HandlerFunc that redirects requests to the given URL.
// It is a convenience wrapper around http.Redirect.
//
// Example:
//
//	r.Get("/about", kage.Redirect("/about-us", http.StatusMovedPermanently))
//	r.Handle("/old-api", kage.Redirect("/api/v2", http.StatusFound))
func Redirect(url string, code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	}
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
