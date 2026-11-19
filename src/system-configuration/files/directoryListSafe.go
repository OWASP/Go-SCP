// Command directoryListSafe serves static files without listing directory
// contents.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM,
	)
	defer stop()

	err := run(ctx, os.Args, os.Stderr)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		stop()
		os.Exit(1)
	}
}

// run serves the static directory until ctx is cancelled.
func run(ctx context.Context, args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	addr := flags.String("addr", ":8080", "address to listen on")
	dir := flags.String("dir", "tmp/static/", "directory to serve")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	fs := justFilesFilesystem{http.Dir(*dir)}
	srv := &http.Server{
		Addr:              *addr,
		Handler:           http.StripPrefix("/tmp/static", http.FileServer(fs)),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	select {
	case err := <-errc:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// justFilesFilesystem serves files from fs but hides directory contents.
type justFilesFilesystem struct {
	fs http.FileSystem
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredReaddirFile{f}, nil
}

// neuteredReaddirFile reports every directory as empty, so http.FileServer
// has nothing to list.
type neuteredReaddirFile struct {
	http.File
}

func (f neuteredReaddirFile) Readdir(_ int) ([]os.FileInfo, error) {
	return nil, nil
}
