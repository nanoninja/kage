package kage_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nanoninja/kage"
	"github.com/nanoninja/kage/middleware"
)

// Example showing how to initialize the router and add a basic route.
func Example() {
	r := kage.New()

	r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "Hello, Nanoninja!")
	})

	// To serve:
	// http.ListenAndServe(":8080", r)
}

// Example showing how to use groups to organize routes under a prefix.
func ExampleRouter_Group() {
	r := kage.New()

	r.Group("/api", func(api kage.Router) {
		api.Get("/status", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprint(w, "OK")
		})
	})
}

// Example showing how to use the With method for inline middlewares.
func ExampleRouter_With() {
	r := kage.New()

	r.With(middleware.Logger(nil)).Get("/private", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "Authenticated access")
	})
}

// Example showing how to register multiple HTTP methods on a single path.
func ExampleRouter_Route() {
	r := kage.New()

	r.Route("/users/{id}", func(rt kage.Route) {
		rt.Get(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, "get user %s", kage.Param(r, "id"))
		})
		rt.Put(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, "update user %s", kage.Param(r, "id"))
		})
		rt.Delete(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, "delete user %s", kage.Param(r, "id"))
		})
	})
}

// Example showing how to mount a sub-router at a prefix.
func ExampleRouter_Mount() {
	users := kage.New()
	users.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "list users")
	})
	users.Post("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "create user")
	})

	r := kage.New()
	r.Mount("/users", users)
}

// Example showing how to serve static files from a local directory.
func ExampleFileServer() {
	r := kage.New()

	// Files in "./public" are served at /assets/*
	r.Mount("/assets", kage.FileServer("./public"))
}

// Example showing how to serve static files from an http.FileSystem.
func ExampleFileServerFS() {
	r := kage.New()

	// Useful with Go's embed package:
	// //go:embed dist/*
	// var embeddedFiles embed.FS
	// r.Mount("/assets", kage.FileServerFS(http.FS(embeddedFiles)))

	r.Mount("/assets", kage.FileServerFS(http.Dir("./public")))
}

// Example showing how to inspect all registered routes.
func ExampleRouter_Routes() {
	r := kage.New(kage.WithPrefix("/api"))

	r.Get("/health", func(_ http.ResponseWriter, _ *http.Request) {})
	r.Group("/v1", func(v1 kage.Router) {
		v1.Get("/users", func(_ http.ResponseWriter, _ *http.Request) {})
		v1.Post("/users", func(_ http.ResponseWriter, _ *http.Request) {})
	})

	for _, route := range r.Routes() {
		fmt.Printf("%s %s\n", route.Method, route.Pattern)
	}
	// Output:
	// GET /api/health
	// GET /api/v1/users
	// POST /api/v1/users
}

// Example showing how to assign and retrieve a unique request ID.
func ExampleRequestID() {
	r := kage.New()

	r.Use(middleware.RequestID)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetRequestID(r)
		_, _ = fmt.Fprint(w, id)
	})
}

// Example showing how to enable gzip/deflate response compression.
func ExampleCompress() {
	r := kage.New()

	r.Use(middleware.Compress)

	r.Get("/data", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"message":"hello"}`)
	})
}

// Example showing a complete setup with middlewares, groups, and graceful shutdown.
func ExampleServeGraceful() {
	r := kage.New(kage.WithPrefix("/api"))

	r.Use(middleware.Recoverer(nil, nil))
	r.Use(middleware.Logger(nil))
	r.Use(middleware.CORS(middleware.DefaultCORSConfig()))
	r.Use(middleware.Timeout(10*time.Second, nil))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Group("/v1", func(v1 kage.Router) {
		v1.Route("/articles/{id}", func(rt kage.Route) {
			rt.Get(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintf(w, "get article %s", kage.Param(r, "id"))
			})
			rt.Delete(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
		})
	})

	srv := &http.Server{Addr: ":8080", Handler: r}
	_ = kage.ServeGraceful(srv, srv.ListenAndServe, 10*time.Second)
}
