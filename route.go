// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import "net/http"

// Route defines the interface for registering multiple HTTP methods
// on a single fixed path without repeating the pattern.
type Route interface {
	Connect(h http.HandlerFunc)
	Delete(h http.HandlerFunc)
	Get(h http.HandlerFunc)
	Head(h http.HandlerFunc)
	Options(h http.HandlerFunc)
	Patch(h http.HandlerFunc)
	Post(h http.HandlerFunc)
	Put(h http.HandlerFunc)
	Trace(h http.HandlerFunc)
	Use(...func(http.Handler) http.Handler)
}

// route is the implementation of the Route interface.
type route struct {
	pattern string
	r       *router
}

func (rt *route) Connect(h http.HandlerFunc) {
	rt.r.Method(http.MethodConnect, rt.pattern, h)
}

func (rt *route) Delete(h http.HandlerFunc) {
	rt.r.Method(http.MethodDelete, rt.pattern, h)
}

func (rt *route) Get(h http.HandlerFunc) {
	rt.r.Method(http.MethodGet, rt.pattern, h)
}

func (rt *route) Head(h http.HandlerFunc) {
	rt.r.Method(http.MethodHead, rt.pattern, h)
}

func (rt *route) Options(h http.HandlerFunc) {
	rt.r.Method(http.MethodOptions, rt.pattern, h)
}

func (rt *route) Patch(h http.HandlerFunc) {
	rt.r.Method(http.MethodPatch, rt.pattern, h)
}

func (rt *route) Post(h http.HandlerFunc) {
	rt.r.Method(http.MethodPost, rt.pattern, h)
}

func (rt *route) Put(h http.HandlerFunc) {
	rt.r.Method(http.MethodPut, rt.pattern, h)
}

func (rt *route) Trace(h http.HandlerFunc) {
	rt.r.Method(http.MethodTrace, rt.pattern, h)
}

func (rt *route) Use(middlewares ...func(http.Handler) http.Handler) {
	rt.r.Use(middlewares...)
}
