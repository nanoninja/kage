# Kage

**Kage** is a lightweight, idiomatic, and high-performance HTTP router for **Go 1.22+**. Built directly on top of `net/http`, it provides fluid request routing with zero external dependencies and modern structured logging.

[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg?style=flat&logo=go)](https://github.com/nanoninja/kage)
[![Go Reference](https://pkg.go.dev/badge/github.com/nanoninja/kage.svg)](https://pkg.go.dev/github.com/nanoninja/kage)
[![Go Report Card](https://goreportcard.com/badge/github.com/nanoninja/kage)](https://goreportcard.com/report/github.com/nanoninja/kage)
[![Tests](https://github.com/nanoninja/kage/actions/workflows/ci.yaml/badge.svg)](https://github.com/nanoninja/kage/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/nanoninja/kage/branch/main/graph/badge.svg)](https://codecov.io/gh/nanoninja/kage)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](LICENSE)

## Features

* **Zero Dependencies**: Pure Go standard library.
* **Go 1.22+ Ready**: Native support for HTTP methods and path parameters.
* **Functional Options**: Clean and flexible router initialization.
* **Fluid Middleware Stack**: Global, group-level, or per-route middlewares (FIFO).
* **Structured Logging**: Built-in support for `slog`.
* **Panic Recovery**: Robust recovery middleware with custom error handler.
* **Response Instrumentation**: Captured status codes and size via a custom `ResponseWriter`.
* **Static File Serving**: Easy-to-use helpers for serving assets with automatic path stripping.
* **Native Compatibility**: Fully compatible with `http.ResponseController` (Unwrap).

## Installation

```bash
go get github.com/nanoninja/kage
```

## Quick Start

```go
package main

import (
    "log/slog"
    "net/http"
    "os"
    "github.com/nanoninja/kage"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    // Initialize router with options
    r := kage.New(
        kage.WithPrefix("/api"),
    )

    // Global Middlewares
    r.Use(kage.Recoverer(logger, nil))
    r.Use(kage.Logger(logger))

    // Simple Route
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    // Sub-group /api/v1
    r.Group("/v1", func(v1 kage.Router) {
        v1.Get("/users", func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("User List"))
        })
    })

    http.ListenAndServe(":8080", r)
}
```

## Configuration and Options

Kage follows the **Functional Options** pattern. You can pass multiple options to the `New()` constructor to tune your router's behavior from the start.

```go
r := kage.New(
    kage.WithPrefix("/api/v1"),             // Set a global prefix for all routes
    kage.WithNotFound(customNotFound),      // Set custom 404 handler during init
    kage.WithMux(http.NewServeMux()),       // Provide a custom ServeMux instance
)
```

| Option | Description |
| --- | --- |
| `WithPrefix(string)` | Sets a global base path (e.g., `/api`) for every route registered. |
| `WithNotFound(HandlerFunc)` | Configures the default 404 response at startup. |
| `WithMux(*http.ServeMux)` | Injects a pre-configured `ServeMux` for advanced stdlib tuning. |

## Routing and Patterns

This router leverages the Go 1.22+ `ServeMux` engine. It supports modern features like method-based routing and strict path matching directly in the pattern string.

### Pattern Syntax

* **`GET /path`**: Restricts the route to a specific HTTP method.
* **`/{$}` (Strict Root)**: Matches **only** the exact path `/`.
* **`/posts/{id}`**: Captures a path parameter. Use `req.PathValue("id")` to retrieve it.
* **`/files/` (Subtree)**: A trailing slash creates a subtree match.

### Path Merging and Grouping

Kage automatically handles slash concatenation using a clean `wrapPath` logic, ensuring robust prefix inheritance without double slashes.

| Prefix | Pattern | Final Registered Path | Behavior |
| --- | --- | --- | --- |
| `/api` | `/{$}` | `/api/{$}` | Matches exactly `/api/` |
| `/api` | `/users` | `/api/users` | Matches exactly `/api/users` |
| `/api/v1` | `/users/{id}` | `/api/v1/users/{id}` | Matches `/api/v1/users/123` |

### Advanced Routing with Wildcards

The `{path...}` wildcard captures everything from that point onwards.

```go
// Matches /gallery/summer/vacation.jpg
r.Get("/gallery/{path...}", func(w http.ResponseWriter, r *http.Request) {
    path := r.PathValue("path")
    fmt.Fprintf(w, "Fetching file at: %s", path)
})
```

## Custom 404 Handler

You can override the default 404 behavior using the `NotFound` method or the `WithNotFound` option.

```go
r.NotFound(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprint(w, `{"error": "not_found"}`)
})
```

## Static Files

Kage handles static file serving by automatically calculating the correct prefix to strip, even when routes are nested within multiple groups.

### Local Directory

The simplest way to serve a physical folder from your disk.
```go
// Files in "./public" are available at http://localhost:8080/assets/*
r.Static("/assets", "./public")
```

### Embedded Files (Go Embed)

When using `embed.FS`, you have two architectural patterns depending on how you want your URLs to look.

**Pattern 1: Simple (Internal folder visible in URL)**
If your files are in a `dist` folder, that folder name will appear in the URL path.

```go
//go:embed dist/*
var embedFS embed.FS

// A file at dist/style.css is accessed via:
// http://localhost:8080/static/dist/style.css
r.StaticFS("/static", http.FS(embedFS))
```

**Pattern 2: Clean URLs (Internal folder hidden)**
If you want to hide the dist folder from your URLs, use fs.Sub to re-root the file system.

```go
//go:embed dist/*
var embedFS embed.FS

func main() {
    r := kage.New()

    // Re-root the filesystem to the "dist" folder
    subFS, _ := fs.Sub(embedFS, "dist")

    // The same file dist/style.css is now accessed via:
    // http://localhost:8080/static/style.css
    r.StaticFS("/static", http.FS(subFS))
}
```

**Within Nested Groups**
Kage correctly calculates the http.StripPrefix even when deeply nested, allowing you to mount assets anywhere in your API tree without manual path management.

```go
r.Group("/api", func(api kage.Router) {
    api.Group("/v1", func(v1 kage.Router) {
        // Automatically strips "/api/v1/docs/"
        // Accessible at http://localhost:8080/api/v1/docs/swagger.json
        v1.StaticFS("/docs", http.FS(myDocs))
    })
})
```

## Middlewares

Middlewares follow a **First In, First Out (FIFO)** execution order.

### Built-in Middlewares

* **Logger**: Uses `slog` to capture method, path, status code, and duration.
* **Recoverer**: Gracefully recovers from panics and logs the error.

### Global and Group Isolation

Middlewares applied to a group are isolated. When a group is created, it **clones** the parent's middleware stack.

### Per-Route Middlewares

Use `With` to apply middlewares to a specific route without affecting its group.

```go
r.With(AuthMiddleware).Get("/private", handlePrivate)
```

## Advanced: Response Instrumentation

The router wraps `http.ResponseWriter` to allow middlewares to access response data:

* **`rw.Status()`**: Returns the captured HTTP status code.
* **`rw.Written()`**: Returns `true` if headers have been sent.
* **`rw.Unwrap()`**: Access the underlying `http.ResponseWriter`. This is essential for **WebSockets**, **SSE**, and **http.ResponseController** compatibility.

## License

Copyright 2026 The Kage Authors. All rights reserved.
Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
