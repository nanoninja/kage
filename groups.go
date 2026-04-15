// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"strings"
)

// Group creates a new router group with the given prefix and optional middlewares.
// It executes the provided callback function with the new group.
func (r *router) Group(prefix string, fn func(Router)) {
	g := r.clone()
	g.prefix = r.wrapPath(prefix)

	if fn != nil {
		fn(g)
	}
}

// Mount attaches an http.Handler at the given prefix.
// The mounted handler receives requests with the prefix stripped from the path.
func (r *router) Mount(prefix string, h http.Handler) {
	fullPath := r.wrapPath(prefix)
	pattern := strings.TrimRight(fullPath, "/") + "/"

	r.mux.Handle(pattern, http.StripPrefix(strings.TrimRight(fullPath, "/"), r.chain(h)))
}

// Route creates a new Route for the given pattern, allowing multiple
// HTTP methods to be registered on the same path without repetition.
func (r *router) Route(pattern string, fn func(Route)) {
	rt := &route{
		pattern: pattern,
		r:       r,
	}
	if fn != nil {
		fn(rt)
	}
}

// Use appends the given middlewares to the router's middleware stack.
func (r *router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// With creates a new temporary router (a clone) with the provided
// middlewares appended to the current stack.
// It is useful for applying middlewares to specific routes without
// modifying the main router's state.
func (r *router) With(middlewares ...func(http.Handler) http.Handler) Router {
	g := r.clone()
	g.Use(middlewares...)
	return g
}

// clone creates a shallow copy of the router, including its prefix,
// mux, and a copy of the current middleware stack.
func (r *router) clone() *router {
	mws := make([]func(http.Handler) http.Handler, len(r.middlewares))
	copy(mws, r.middlewares)

	return &router{
		prefix:             r.prefix,
		mux:                r.mux,
		routes:             r.routes,
		middlewares:        mws,
		notFoundRegistered: r.notFoundRegistered,
	}
}
