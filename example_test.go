package kage_test

import (
	"fmt"
	"net/http"

	"github.com/nanoninja/kage"
	"github.com/nanoninja/kage/middleware"
)

// Example showing how to initialize the router and add a basic route.
func Example() {
	r := kage.New()

	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, Nanoninja!")
	})

	// To serve:
	// http.ListenAndServe(":8080", r)
}

// Example showing how to use groups to organize routes under a prefix.
func ExampleRouter_Group() {
	r := kage.New()

	r.Group("/api", func(api kage.Router) {
		api.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "OK")
		})
	})
}

// Example showing how to use the With method for inline middlewares.
func ExampleRouter_With() {
	r := kage.New()

	// Only this specific route will use the Logger
	r.With(middleware.Logger(nil)).Get("/private", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Authenticated access")
	})
}
