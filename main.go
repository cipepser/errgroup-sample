package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	os.Exit(run(context.Background()))
}

func run(ctx context.Context) int {
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	eg.Go(func() error {
		return runServer(ctx)
	})
	eg.Go(func() error {
		return Signal(ctx)
	})
	eg.Go(func() error {
		<-ctx.Done()
		return ctx.Err()
	})

	if err := eg.Wait(); err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

func Signal(ctx context.Context) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-ctx.Done():
		signal.Reset()
		return nil
	case sig := <-sigCh:
		return fmt.Errorf("signal received: %v", sig.String())
	}
}

func runServer(ctx context.Context) error {
	s := &http.Server{
		Addr: ":8888",
	}

	errCh := make(chan error)
	go func() {
		defer close(errCh)
		if err := s.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return s.Shutdown(ctx)
	case err := <-errCh:
		return err

	}
}
