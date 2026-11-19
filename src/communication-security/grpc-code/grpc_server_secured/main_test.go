package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/OWASP/Go-SCP/src/communication-security/grpc-code/internal/testcert"
	pb "github.com/OWASP/Go-SCP/src/communication-security/grpc-code/samplebuff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestRunServesOverTLSAndStopsOnCancel(t *testing.T) {
	certs := testcert.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	stdoutReader, stdout := io.Pipe()

	// Start the server on a free port; it reports the address on stdout.
	args := []string{
		"server",
		"-addr",
		"127.0.0.1:0",
		"-cert",
		certs.CertFile,
		"-key",
		certs.KeyFile,
	}
	errc := make(chan error, 1)
	go func() {
		errc <- run(ctx, args, stdout, io.Discard)
		_ = stdout.Close()
	}()
	line, err := bufio.NewReader(stdoutReader).ReadString('\n')
	if err != nil {
		t.Fatalf("reading listen address: %s (server error: %v)", err, <-errc)
	}
	go func() { _, _ = io.Copy(io.Discard, stdoutReader) }()
	addr := strings.TrimSpace(strings.TrimPrefix(line, "listening on "))

	// Call it as a client that trusts the test CA.
	caPEM, err := os.ReadFile(certs.CAFile)
	if err != nil {
		t.Fatalf("reading CA: %s", err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caPEM)
	creds := credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12, RootCAs: pool})
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		t.Fatalf("creating client: %s", err)
	}
	t.Cleanup(func() {
		if closeErr := conn.Close(); closeErr != nil {
			t.Errorf("closing client: %s", closeErr)
		}
	})

	resp, err := pb.NewSampleServiceClient(conn).Greet(ctx, &pb.SendMsg{Name: "tester"})
	if err != nil {
		t.Fatalf("Greet: %s", err)
	}
	if got, want := resp.GetMessage(), "Hey tester"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	// Cancelling the context stops the server cleanly.
	cancel()
	select {
	case runErr := <-errc:
		if runErr != nil {
			t.Fatalf("run returned error after cancel: %s", runErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not return after cancel")
	}
}
