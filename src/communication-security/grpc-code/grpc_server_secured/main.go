// Command grpc_server_secured is a gRPC server that only accepts TLS
// connections.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/OWASP/Go-SCP/src/communication-security/grpc-code/samplebuff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM,
	)
	defer stop()

	err := run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		stop()
		os.Exit(1)
	}
}

// run serves the SampleService over TLS until ctx is cancelled, then stops
// gracefully. It prints the listening address to stdout once it accepts
// connections.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	addr := flags.String("addr", ":10001", "address to listen on")
	certFile := flags.String("cert", "../cert/service.pem", "TLS certificate file")
	keyFile := flags.String("key", "../cert/service.key", "TLS private key file")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	// Load the server certificate and private key used for TLS.
	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	if err != nil {
		return fmt.Errorf("load TLS certificate: %w", err)
	}

	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", *addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	srv := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterSampleServiceServer(srv, &server{logger: slog.New(slog.NewTextHandler(stderr, nil))})

	// Stop accepting new calls and let in-flight ones finish when ctx is
	// cancelled. The done channel lets this goroutine exit if Serve fails.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			srv.GracefulStop()
		case <-done:
		}
	}()

	_, _ = fmt.Fprintf(stdout, "listening on %s\n", lis.Addr())
	if err = srv.Serve(lis); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

// server implements sample.SampleServiceServer.
type server struct {
	pb.UnimplementedSampleServiceServer

	logger *slog.Logger
}

// Greet replies with a greeting for the name in the request.
func (s *server) Greet(ctx context.Context, in *pb.SendMsg) (*pb.SendResp, error) {
	// slog quotes values, so a name containing newlines cannot forge log lines.
	s.logger.InfoContext(ctx, "received greeting", "name", in.GetName())
	return &pb.SendResp{Message: "Hey " + in.GetName()}, nil
}
