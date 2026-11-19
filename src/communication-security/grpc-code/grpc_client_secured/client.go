package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/OWASP/Go-SCP/src/communication-security/grpc-code/samplebuff"
)

const defaultName = "Art Rosenbaum"

var (
	addr   = flag.String("addr", "localhost:10001", "Address of Server")
	name   = flag.String("name", defaultName, "Name to greet")
	caFile = flag.String("ca", "../cert/ca.cert", "CA certificate used to verify the server")
)

func main() {
	flag.Parse()

	b, err := os.ReadFile(*caFile)
	if err != nil {
		log.Fatalf("Could not read CA certificate: %v", err)
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		log.Fatal("credentials: failed to append certificates")
	}

	// Verify the server certificate against our CA. Never set
	// InsecureSkipVerify to true: that disables certificate verification.
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    cp,
	}

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()
	c := pb.NewSampleServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Greet(ctx, &pb.SendMsg{Name: *name})
	if err != nil {
		log.Fatalf("Could not send message: %v", err)
	}
	log.Printf("Received response: %s", r.GetMessage())
}
