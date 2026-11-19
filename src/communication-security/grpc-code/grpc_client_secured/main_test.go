package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"net"
	"strings"
	"testing"

	"github.com/OWASP/Go-SCP/src/communication-security/grpc-code/internal/testcert"
	pb "github.com/OWASP/Go-SCP/src/communication-security/grpc-code/samplebuff"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestRunVerifiesServerCertificate(t *testing.T) {
	trusted := testcert.New(t)
	untrusted := testcert.New(t)
	addr := testServer(t, &trusted.Server)

	cases := map[string]struct {
		caFile     string
		wantOutput string
		wantErr    string
	}{
		"server signed by trusted CA": {
			caFile:     trusted.CAFile,
			wantOutput: "Received response: Hey tester\n",
		},
		"server signed by unknown CA": {
			caFile:  untrusted.CAFile,
			wantErr: "certificate signed by unknown authority",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var stdout bytes.Buffer
			args := []string{"client", "-addr", addr, "-name", "tester", "-ca", tc.caFile}

			err := run(t.Context(), args, &stdout, &bytes.Buffer{})

			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
			if got := stdout.String(); got != tc.wantOutput {
				t.Fatalf("expected output %q, got %q", tc.wantOutput, got)
			}
		})
	}
}

type greeter struct {
	pb.UnimplementedSampleServiceServer
}

func (greeter) Greet(_ context.Context, in *pb.SendMsg) (*pb.SendResp, error) {
	return &pb.SendResp{Message: "Hey " + in.GetName()}, nil
}

// testServer starts a TLS gRPC server on a free local port and returns its
// address. The server stops when the test ends.
func testServer(t *testing.T, cert *tls.Certificate) string {
	t.Helper()
	var lc net.ListenConfig
	lis, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("testServer: %s", err)
	}
	srv := grpc.NewServer(grpc.Creds(credentials.NewServerTLSFromCert(cert)))
	pb.RegisterSampleServiceServer(srv, greeter{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}
