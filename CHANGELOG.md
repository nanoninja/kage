# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.3] - 2026-04-25

### Added
- `Redirect` helper — returns an `http.HandlerFunc` that redirects to the given URL and status code.
- `CacheControl` middleware — sets the `Cache-Control` header on every response.
- `NoCache` middleware — sets `Cache-Control`, `Pragma`, and `Expires` headers to prevent caching.

### Fixed
- `Compress` middleware now falls back to uncompressed response when `flate.NewWriter` fails instead of panicking.
- `notFoundRegistered` replaced by `*atomic.Bool` shared across clones — prevents duplicate `"/"` registration on the same `ServeMux` when `NotFound` is called from a clone or concurrently.

### Changed
- Minimum Go version bumped to 1.24.
- Benchmarks migrated to `b.Loop()` (Go 1.24 idiomatic form).
- New benchmarks: `With`, `Mount`, `Route` multi-method, `Routes` introspection, parallel dispatch.

### CI
- Added `go build` and `go vet` steps to the CI pipeline.
- Extended branch triggers to `feat/**` and `refactor/**`.

## [0.7.0] - 2026-04-20

### Fixed
- `Route` scope isolation: scope is now reset after each method registration to prevent leaking middleware state across methods.
- `Handle` method now normalizes lowercase HTTP method patterns.

## [0.6.0] - 2026-04-15

### Added
- `Routes()` introspection: returns all registered routes as `[]RouteInfo` for debugging and documentation.
- `RequestID` middleware: generates or reuses `X-Request-ID` header and propagates it via request context.
- `Compress` middleware: gzip and deflate response compression with priority negotiation.

### Fixed
- `Compress` middleware: replaced `defer Close` with explicit call to avoid closing after response is written.

## [0.5.0] - 2026-04-15

### Added
- `Route` for registering multiple HTTP methods on a single path with shared middleware.
- `CORS` middleware with `DefaultCORSConfig` and preflight `OPTIONS` handling.
- `Timeout` middleware with context deadline propagation and custom failure handler.
- `Mount` for attaching any `http.Handler` as a sub-router at a path prefix.

### Changed
- `FileServer` / `FileServerFS` replace the previous `Static` / `StaticFS` helpers.
- `Recoverer` signature harmonized with `Timeout` and `Logger`.

### Fixed
- Middleware chain ordering, `NotFound` double-registration guard, and `Handle` delegation.

## [0.4.0] - 2026-03-04

### Changed
- Middlewares moved to a dedicated `middleware` sub-package.
- `WrapResponseWriter` is now exported.

## [0.3.0] - 2026-03-04

### Changed
- Removed redundant `ServeHTTP` from the `Router` interface (breaking: implementations no longer need to implement it explicitly).

## [0.2.0] - 2026-03-03

### Added
- `ServeGraceful` for clean server shutdown on `SIGINT` / `SIGTERM`.

## [0.1.0] - 2026-03-03

### Added
- Initial release: `Router` with `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options`, `Connect`, `Trace`.
- Route grouping with `Group` and prefix inheritance.
- Middleware chaining with `Use` and `With`.
- `Logger`, `Recoverer` middlewares.
- `WrapResponseWriter` for status code capture.
- `Param` helper for path parameter extraction.
- `FileServer` / `FileServerFS` for static file serving.
- Functional options: `WithMux`, `WithPrefix`, `WithNotFound`.

[Unreleased]: https://github.com/nanoninja/kage/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/nanoninja/kage/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/nanoninja/kage/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/nanoninja/kage/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/nanoninja/kage/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/nanoninja/kage/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/nanoninja/kage/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/nanoninja/kage/releases/tag/v0.1.0
