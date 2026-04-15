# Kage

**Kage** is a lightweight, idiomatic, and high-performance HTTP router for **Go 1.22+**. Built directly on top of `net/http`, it provides fluid request routing with zero external dependencies and modern structured logging.

[![Go Version](https://img.shields.io/badge/go-1.22%2B-00ADD8.svg?style=flat&logo=go)](https://github.com/nanoninja/kage)
[![Go Reference](https://pkg.go.dev/badge/github.com/nanoninja/kage.svg)](https://pkg.go.dev/github.com/nanoninja/kage)
[![Go Report Card](https://goreportcard.com/badge/github.com/nanoninja/kage)](https://goreportcard.com/report/github.com/nanoninja/kage)
[![CI](https://github.com/nanoninja/kage/actions/workflows/ci.yaml/badge.svg)](https://github.com/nanoninja/kage/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/nanoninja/kage/branch/main/graph/badge.svg)](https://codecov.io/gh/nanoninja/kage)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](LICENSE)

## Why Kage?

Go 1.22 introduced method-based routing and path parameters directly into the standard `http.ServeMux`. Most developers are still reaching for heavy frameworks when the stdlib already handles the hard part.

Kage builds on that foundation instead of replacing it. You get route grouping, middleware chaining, graceful shutdown, and a ready-to-use middleware package — all without a single external dependency. If you ever need to drop down to plain `net/http`, everything is compatible.

> If chi or gin do more than you need, Kage is the layer you were missing.

## Features

* **Zero Dependencies** — Pure Go standard library.
* **Go 1.22+ Ready** — Native support for HTTP methods and path parameters.
* **Functional Options** — Clean and flexible router initialization.
* **Fluid Middleware Stack** — Global, group-level, or per-route middlewares (FIFO).
* **Route Grouping** — Organize routes under shared prefixes with isolated middleware stacks.
* **Multi-Method Routes** — Register multiple HTTP methods on a single path with `Route`.
* **Sub-Router Mounting** — Attach any `http.Handler` at a prefix with `Mount`.
* **Structured Logging** — Built-in support for `slog`.
* **Panic Recovery** — Robust recovery middleware with custom error handler.
* **CORS** — Configurable CORS middleware with preflight support.
* **Timeout** — Per-request context timeout with custom error handler.
* **Response Instrumentation** — Captured status codes and size via a custom `ResponseWriter`.
* **Static File Serving** — Easy-to-use helpers for serving assets with automatic path stripping.
* **Graceful Shutdown** — Clean server lifecycle management for Docker/Kubernetes.
* **Native Compatibility** — Fully compatible with `http.ResponseController` (Unwrap).

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
    "time"
    "github.com/nanoninja/kage"
    "github.com/nanoninja/kage/middleware"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    r := kage.New(
        kage.WithPrefix("/api"),
    )

    r.Use(middleware.Recoverer(logger, nil))
    r.Use(middleware.Logger(logger))
    r.Use(middleware.CORS(middleware.DefaultCORSConfig()))

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    r.Group("/v1", func(v1 kage.Router) {
        v1.Route("/users/{id}", func(rt kage.Route) {
            rt.Get(getUser)
            rt.Put(updateUser)
            rt.Delete(deleteUser)
        })
    })

    srv := &http.Server{Addr: ":8080", Handler: r}
    kage.ServeGraceful(srv, srv.ListenAndServe, 10*time.Second)
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
* **`/posts/{id}`**: Captures a path parameter. Use `kage.Param(r, "id")` to retrieve it.
* **`/files/` (Subtree)**: A trailing slash creates a subtree match.

### HTTP Verb Shortcuts

```go
r.Get("/users", listUsers)
r.Post("/users", createUser)
r.Put("/users/{id}", updateUser)
r.Patch("/users/{id}", patchUser)
r.Delete("/users/{id}", deleteUser)
```

### Multi-Method Routes

Register multiple methods on the same path without repeating it:

```go
r.Route("/users/{id}", func(rt kage.Route) {
    rt.Get(getUser)
    rt.Put(updateUser)
    rt.Delete(deleteUser)
})
```

### Path Parameters

```go
r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := kage.Param(r, "id")
    fmt.Fprintf(w, "User: %s", id)
})
```

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
    path := kage.Param(r, "path")
    fmt.Fprintf(w, "Fetching file at: %s", path)
})
```

## Groups

Organize related routes under a shared prefix. Each group clones the parent middleware stack, so group-level middlewares do not affect sibling groups or the parent.

```go
r.Group("/api", func(api kage.Router) {
    api.Use(authMiddleware)

    api.Group("/v1", func(v1 kage.Router) {
        v1.Get("/users", listUsers)
        v1.Post("/users", createUser)
    })
})
```

## Mount

Attach any `http.Handler` (including another kage router) at a given prefix. The mounted handler receives requests with the prefix stripped from the path.

```go
// Mount a sub-router
users := kage.New()
users.Get("/", listUsers)
users.Post("/", createUser)

r.Mount("/users", users)

// Mount a third-party handler
r.Mount("/metrics", promhttp.Handler())
```

## Custom 404 Handler

```go
r.NotFound(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprint(w, `{"error": "not_found"}`)
})
```

## Static Files

Kage provides `FileServer` and `FileServerFS` helper functions that return an `http.Handler` ready to be used with `Mount`. The prefix stripping is handled automatically.

### Local Directory

```go
// Files in "./public" are available at http://localhost:8080/assets/*
r.Mount("/assets", kage.FileServer("./public"))
```

### Embedded Files (Go Embed)

```go
//go:embed dist/*
var embedFS embed.FS

// Re-root to hide the dist/ folder from URLs
subFS, _ := fs.Sub(embedFS, "dist")
r.Mount("/static", kage.FileServerFS(http.FS(subFS)))
```

## Middlewares

Middlewares follow a **First In, First Out (FIFO)** execution order. Kage provides essential middlewares in the `middleware` sub-package.

### Applying Middlewares

```go
// Global — applies to all routes
r.Use(middleware.Logger(nil))

// Group-level — applies only to routes in this group
r.Group("/admin", func(admin kage.Router) {
    admin.Use(authMiddleware)
    admin.Get("/dashboard", showDashboard)
})

// Per-route — applies only to this route
r.With(rateLimitMiddleware).Post("/login", handleLogin)
```

### Logger

Logs method, path, status code, and duration using `slog`. Accepts a custom `*slog.Logger` or `nil` for the default.

```go
r.Use(middleware.Logger(nil))
```

### Recoverer

Recovers from panics and logs the error. Accepts an optional `onFailure` handler for custom error responses.

```go
// Default: 500 Internal Server Error
r.Use(middleware.Recoverer(nil, nil))

// Custom JSON error
r.Use(middleware.Recoverer(logger, func(w http.ResponseWriter, r *http.Request, err any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, `{"error":"%v"}`, err)
}))
```

### CORS

Configurable CORS middleware with automatic preflight (`OPTIONS`) handling.

```go
// Development — allow all origins
r.Use(middleware.CORS(middleware.DefaultCORSConfig()))

// Production — restrict to specific origins
r.Use(middleware.CORS(middleware.DefaultCORSConfig("https://myapp.com")))

// Fully custom
r.Use(middleware.CORS(middleware.CORSConfig{
    AllowedOrigins:   []string{"https://myapp.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           3600,
}))
```

### Timeout

Cancels the request context after the given duration. Uses an optional `onFailure` handler for custom error responses.

```go
// Default: 503 Service Unavailable
r.Use(middleware.Timeout(5*time.Second, nil))

// Custom JSON error
r.Use(middleware.Timeout(5*time.Second, func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusServiceUnavailable)
    w.Write([]byte(`{"error":"request timeout"}`))
}))
```

## Advanced: Response Instrumentation

Kage includes a public `WrapResponseWriter` in the `middleware` package. This allows both built-in and custom middlewares to access response metadata that is normally hidden by the standard `http.ResponseWriter`.

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ww := middleware.NewWrapResponseWriter(w)
        next.ServeHTTP(ww, r)
        slog.Info("response", "status", ww.Status())
    })
}
```

| Method | Description |
| --- | --- |
| `rw.Status()` | Returns the captured HTTP status code. |
| `rw.Written()` | Returns `true` if headers have been sent. |
| `rw.Unwrap()` | Access the underlying `http.ResponseWriter` (WebSockets, SSE, `http.ResponseController`). |

## Graceful Shutdown

```go
srv := &http.Server{Addr: ":8080", Handler: r}

if err := kage.ServeGraceful(srv, srv.ListenAndServe, 10*time.Second); err != nil {
    log.Fatalf("server error: %v", err)
}
```

`ServeGraceful` listens for `SIGINT` and `SIGTERM`, then calls `srv.Shutdown` with the provided timeout. It absorbs `http.ErrServerClosed` so a clean shutdown always returns `nil`.

```go
// TLS support
kage.ServeGraceful(srv, func() error {
    return srv.ListenAndServeTLS("cert.pem", "key.pem")
}, 10*time.Second)
```

## License

This project is licensed under the BSD 3-Clause License.
See the [LICENSE](LICENSE) file for details.
