// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ServerRunner defines a function that starts a server (e.g., srv.ListenAndServe).
type ServerRunner func() error

// ServeGraceful runs the server using the provided runner and ensures
// a clean shutdown when a SIGINT or SIGTERM signal is received.
func ServeGraceful(srv *http.Server, runner ServerRunner, timeout time.Duration) error {
	runnerError := make(chan error, 1)

	go func() {
		runnerError <- runner()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-runnerError:
		if err == http.ErrServerClosed {
			return nil
		}
		return err

	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return srv.Shutdown(ctx)
	}
}
