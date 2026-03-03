// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"strings"
)

// Option defines a function type for configuring the router.
type Option func(*router)

// WithMux allows providing a custom http.ServeMux instance.
func WithMux(mux *http.ServeMux) Option {
	return func(r *router) {
		if mux != nil {
			r.mux = mux
		}
	}
}

// WithNotFound sets a custom handler for 404 Not Found errors during initialization.
func WithNotFound(h http.HandlerFunc) Option {
	return func(r *router) {
		if h != nil {
			r.NotFound(h)
		}
	}
}

// WithPrefix sets an initial global prefix for all routes.
func WithPrefix(prefix string) Option {
	return func(r *router) {
		r.prefix = prefix
	}
}

// Router defines the interface for a layered HTTP router.
type Router interface {
	http.Handler

	ServeHTTP(http.ResponseWriter, *http.Request)

	// Handle registers a new route with a handler for the given pattern.
	Handle(pattern string, h http.Handler)

	// HandleFunc registers a new route with a handler function.
	HandleFunc(pattern string, h http.HandlerFunc)

	// Method registers a new route with a specific HTTP method and handler.
	Method(method, pattern string, h http.Handler)

	// MethodFunc registers a new route with a specific HTTP method and handler function.
	MethodFunc(method, pattern string, h http.HandlerFunc)

	// HTTP Verb shortcuts.
	Connect(pattern string, h http.HandlerFunc)
	Delete(pattern string, h http.HandlerFunc)
	Get(pattern string, h http.HandlerFunc)
	Head(pattern string, h http.HandlerFunc)
	Options(pattern string, h http.HandlerFunc)
	Patch(pattern string, h http.HandlerFunc)
	Post(pattern string, h http.HandlerFunc)
	Put(pattern string, h http.HandlerFunc)
	Trace(pattern string, h http.HandlerFunc)

	// Group creates a new sub-router with a prefix.
	// All routes registered within the provided function will inherit this prefix.
	Group(prefix string, fn func(Router))

	// Use appends one or more middlewares to the router's global middleware stack.
	Use(...func(http.Handler) http.Handler)

	// With returns a new Router instance that includes the provided middlewares
	// in addition to the existing ones. Perfect for method chaining.
	With(middlewares ...func(http.Handler) http.Handler) Router

	// NotFound registers a custom handler for 404 Not Found errors.
	NotFound(h http.HandlerFunc)

	// Static registers a route to serve static files from a local directory.
	// It automatically handles path prefix stripping and subtree matching
	// using the Go 1.22+ "{path...}" wildcard.
	//
	// Example:
	//   r.Static("/static", "./public")
	Static(prefix, root string)

	// StaticFS registers a route to serve static files from an http.FileSystem.
	// This is particularly useful for serving files embedded in the binary
	// using the 'embed' package.
	//
	// Example:
	//   r.StaticFS("/assets", http.FS(embeddedFiles))
	StaticFS(prefix string, fsys http.FileSystem)
}

// New creates a new Router instance with optional configurations.
func New(opts ...Option) Router {
	r := &router{
		mux: http.NewServeMux(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// router is an implementation of the Router interface.
type router struct {
	prefix      string
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// StaticFS registers a route to serve static files from an http.FileSystem.
// It automatically handles prefix stripping and wildcard routing.
func (r *router) StaticFS(prefix string, fsys http.FileSystem) {
	fullPath := r.wrapPath(prefix)

	pattern := fullPath
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}
	pattern += "{path...}"
	handler := http.StripPrefix(fullPath, http.FileServer(fsys))

	r.mux.Handle("GET "+pattern, r.chain(handler))
}

func (r *router) Static(prefix, root string) {
	r.StaticFS(prefix, http.Dir(root))
}
