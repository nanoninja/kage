// Copyright 2026 The Nanoninja Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kage

import (
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

// TestServeGraceful_Signal verifies that the server shuts down correctly
// when a system interrupt signal (SIGINT) is received.
func TestServeGraceful_Signal(t *testing.T) {
	srv := &http.Server{Addr: ":9999"}

	runner := func() error {
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- ServeGraceful(srv, runner, 1*time.Second)
	}()

	// Give the goroutine a moment to start
	time.Sleep(100 * time.Millisecond)

	// Simulate SIGINT (Ctrl+C)
	process, _ := os.FindProcess(os.Getpid())
	_ = process.Signal(syscall.SIGINT)

	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("ServeGraceful returned an unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ServeGraceful timed out waiting for signal shutdown")
	}
}

// TestServeGraceful_ServerError verifies that ServeGraceful returns immediately
// if the server runner encounters an error (e.g., port already in use).
func TestServeGraceful_ServerError(t *testing.T) {
	srv := &http.Server{Addr: ":8080"}

	// Mocking a specific server error
	forcedError := http.ErrServerClosed
	runner := func() error {
		return forcedError
	}

	err := ServeGraceful(srv, runner, 1*time.Second)

	if err != forcedError {
		t.Errorf("expected error %v, got %v", forcedError, err)
	}
}
