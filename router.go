// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import "net/http"

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

// RouteInfo holds the HTTP method and fully resolved pattern of a registered route.
// It is returned by Router.Routes() for introspection and debugging purposes.
type RouteInfo struct {
	Method  string // HTTP method (e.g. "GET", "POST")
	Pattern string // Full resolved path (e.g. "/api/v1/users/{id}")
}

// Router defines the interface for a layered HTTP router.
type Router interface {
	http.Handler

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

	// Mount attaches an http.Handler at the given prefix.
	// The mounted handler receives requests with the prefix stripped from the path.
	//
	// Example:
	//   sub := kage.New()
	//   sub.Get("/users", listUsers)
	//   r.Mount("/api", sub)
	Mount(prefix string, h http.Handler)

	// Use appends one or more middlewares to the router's global middleware stack.
	Use(...func(http.Handler) http.Handler)

	// Route registers multiple HTTP methods on a single fixed path.
	//
	// Example:
	//
	//	r.Route("/users/{id}", func(rt kage.Route) {
	//	    rt.Get(getUser)
	//	    rt.Put(updateUser)
	//	    rt.Delete(deleteUser)
	//	})
	Route(pattern string, fn func(Route))

	// Routes returns all registered routes.
	Routes() []RouteInfo

	// With returns a new Router instance that includes the provided middlewares
	// in addition to the existing ones. Perfect for method chaining.
	With(middlewares ...func(http.Handler) http.Handler) Router

	// NotFound registers a custom handler for 404 Not Found errors.
	NotFound(h http.HandlerFunc)
}

// New creates a new Router instance with optional configurations.
func New(opts ...Option) Router {
	routes := make([]RouteInfo, 0)
	r := &router{
		mux:    http.NewServeMux(),
		routes: &routes,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// router is an implementation of the Router interface.
type router struct {
	prefix             string
	mux                *http.ServeMux
	routes             *[]RouteInfo
	middlewares        []func(http.Handler) http.Handler
	notFoundRegistered bool
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
