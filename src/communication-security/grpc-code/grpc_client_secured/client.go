package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	pb "pentest/grpc/samplebuff"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultName = "Art Rosenbaum"
)

var (
	addr = flag.String("addr", "localhost:10001", "Address of Server")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	b, _ := ioutil.ReadFile("../cert/ca.cert")
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		fmt.Println("credentials: failed to append certificates")
	}

	config := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            cp,
	}

	creds := credentials.NewTLS(config)
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds))
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
		log.Fatalf("could not send message: %v", err)
	}
	log.Printf("Sending message: %s", r.GetMessage())
}
