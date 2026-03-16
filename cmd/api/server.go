package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (a *applicationDependencies) serve() error {
	apiServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.config.port),
		Handler:      a.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second, // increased for slow demo endpoint
		ErrorLog:     slog.NewLogLogger(a.logger.Handler(), slog.LevelError),
	}

	// create a channel to keep track of any errors during the shutdown process
	shutdownError := make(chan error)

	// create a goroutine that runs in the background listening
	// for the shutdown signals
	go func() {
		quit := make(chan os.Signal, 1)                      // receive the shutdown signal
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // signal occurred
		s := <-quit                                          // blocks until a signal is received

		// print newline to ensure the shutdown message appears on a new line
		fmt.Println()

		// message about shutdown in process
		a.logger.Info("shutting down server", "signal", s.String())

		// create a context
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// initiate the shutdown,
		// send nil if shutdown is successful
		shutdownError <- apiServer.Shutdown(ctx)
	}()

	a.logger.Info("starting server", "address", apiServer.Addr,
		"environment", a.config.environment)

	// server keeps running until we receive a shutdown signal
	err := apiServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// check the error channel to see if there were shutdown errors
	err = <-shutdownError
	if err != nil {
		return err
	}

	// graceful shutdown was successful
	a.logger.Info("stopped server", "address", apiServer.Addr)

	return nil
}
