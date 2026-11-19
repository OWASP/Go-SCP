// Command grpc_client_secured calls the gRPC server over TLS, verifying the
// server certificate against a trusted CA.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

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

// run sends one Greet call to the server and prints the response to stdout.
// It fails if the server's certificate is not signed by the CA in -ca.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) (err error) {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	addr := flags.String("addr", "localhost:10001", "address of the server")
	name := flags.String("name", "Art Rosenbaum", "name to greet")
	caFile := flags.String("ca", "../cert/ca.cert", "CA certificate used to verify the server")
	timeout := flags.Duration("timeout", time.Second, "time limit for the call")
	if err = flags.Parse(args[1:]); err != nil {
		return err
	}

	// Trust only servers whose certificate is signed by the CA in -ca. gRPC
	// requires TLS 1.2 or later by default. Never build credentials from a
	// tls.Config with InsecureSkipVerify set to true: that disables
	// certificate verification.
	creds, err := credentials.NewClientTLSFromFile(*caFile, "")
	if err != nil {
		return fmt.Errorf("load CA certificate: %w", err)
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	defer func() { err = errors.Join(err, conn.Close()) }()
	c := pb.NewSampleServiceClient(conn)

	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()

	r, err := c.Greet(ctx, &pb.SendMsg{Name: *name})
	if err != nil {
		return fmt.Errorf("greet: %w", err)
	}
	_, _ = fmt.Fprintf(stdout, "Received response: %s\n", r.GetMessage())
	return nil
}
