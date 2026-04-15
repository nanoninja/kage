// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"strings"
)

func (r *router) Handle(pattern string, h http.Handler) {
	if method, p, ok := strings.Cut(pattern, " "); ok {
		r.Method(method, p, h)
		return
	}
	r.mux.Handle(r.wrapPath(pattern), r.chain(h))
}

func (r *router) HandleFunc(pattern string, h http.HandlerFunc) {
	r.Handle(pattern, h)
}

func (r *router) Method(method, pattern string, h http.Handler) {
	path := r.wrapPath(pattern)
	r.mux.Handle(method+" "+path, r.chain(h))
	*r.routes = append(*r.routes, RouteInfo{Method: method, Pattern: path})
}

func (r *router) MethodFunc(method, pattern string, h http.HandlerFunc) {
	r.Method(method, pattern, h)
}

func (r *router) Connect(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodConnect, pattern, h)
}

func (r *router) Delete(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodDelete, pattern, h)
}

func (r *router) Get(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodGet, pattern, h)
}

func (r *router) Head(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodHead, pattern, h)
}

func (r *router) Options(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodOptions, pattern, h)
}

func (r *router) Patch(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodPatch, pattern, h)
}

func (r *router) Post(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodPost, pattern, h)
}

func (r *router) Put(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodPut, pattern, h)
}

func (r *router) Trace(pattern string, h http.HandlerFunc) {
	r.Method(http.MethodTrace, pattern, h)
}

func (r *router) Routes() []RouteInfo {
	return *r.routes
}

func (r *router) NotFound(h http.HandlerFunc) {
	if r.notFoundRegistered {
		return
	}
	r.notFoundRegistered = true
	r.mux.Handle("/", r.chain(h))
}
