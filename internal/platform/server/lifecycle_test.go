package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

// TC-009-2: a stop signal drains an in-flight request before Run returns.
func TestRun_DrainsInFlightRequest_TC009_2(t *testing.T) {
	handlerStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	handlerFinished := make(chan struct{})

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(handlerStarted)
			<-releaseHandler
			w.WriteHeader(http.StatusOK)
			close(handlerFinished)
		}),
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	stop := make(chan os.Signal, 1)
	runDone := make(chan error, 1)
	go func() { runDone <- Run(srv, ln, time.Second, stop) }()

	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
		if err == nil {
			resp.Body.Close()
		}
	}()

	select {
	case <-handlerStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("handler never started")
	}

	// Fire the stop signal while the handler is still mid-request, then let
	// it finish; Shutdown must wait for it instead of cutting it off.
	stop <- os.Interrupt
	select {
	case <-time.After(50 * time.Millisecond):
	case <-runDone:
		t.Fatal("Run returned before the in-flight handler was released")
	}
	close(releaseHandler)

	select {
	case <-handlerFinished:
	case <-time.After(2 * time.Second):
		t.Fatal("handler never finished")
	}

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run returned an error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run never returned after the in-flight handler finished")
	}
}

// TC-009-3: a handler that outlives the shutdown deadline is force-closed
// instead of hanging Run forever.
func TestRun_ForceClosesAfterDeadline_TC009_3(t *testing.T) {
	handlerStarted := make(chan struct{})
	neverReleased := make(chan struct{})

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(handlerStarted)
			<-neverReleased
		}),
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	stop := make(chan os.Signal, 1)
	runDone := make(chan error, 1)
	go func() { runDone <- Run(srv, ln, 50*time.Millisecond, stop) }()

	go func() {
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://" + ln.Addr().String() + "/")
		if err == nil {
			resp.Body.Close()
		}
	}()

	select {
	case <-handlerStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("handler never started")
	}

	stop <- os.Interrupt

	select {
	case err := <-runDone:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Run error = %v, want context.DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not force-close after the shutdown deadline")
	}
}

// TC-009-2 (listener failure path): if the listener dies before any stop
// signal arrives, Run reports that error immediately instead of blocking.
func TestRun_ReturnsServeErrorWithoutAStopSignal(t *testing.T) {
	srv := &http.Server{Handler: http.NotFoundHandler()}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	// Force Serve to fail immediately.
	ln.Close()

	stop := make(chan os.Signal, 1)
	err = Run(srv, ln, time.Second, stop)
	if err == nil {
		t.Fatal("expected an error from Serve on a closed listener")
	}
}
