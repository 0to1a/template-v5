package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"time"
)

// Run serves ln with srv until a value arrives on stop, then drains
// in-flight requests for at most shutdownTimeout before forcing the
// listener closed. It returns nil once shutdown completes cleanly (drained
// or force-closed after the deadline), or the error that made the listener
// fail before a stop signal ever arrived.
func Run(srv *http.Server, ln net.Listener, shutdownTimeout time.Duration, stop <-chan os.Signal) error {
	serveErr := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErr <- err
	}()

	select {
	case err := <-serveErr:
		// The listener died on its own (e.g. bind failure) before any stop
		// signal arrived; there is nothing left to drain.
		return err
	case <-stop:
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	// Shutdown's own error (context deadline exceeded) already reports the
	// force-close case; either way the serve goroutine has by now returned
	// (Shutdown blocks until Serve does), so draining serveErr never hangs.
	shutdownErr := srv.Shutdown(ctx)
	<-serveErr
	return shutdownErr
}
