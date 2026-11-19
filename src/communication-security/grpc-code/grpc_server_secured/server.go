package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/OWASP/Go-SCP/src/communication-security/grpc-code/samplebuff"
)

var (
	port     = flag.Int("port", 10001, "The server port")
	certFile = flag.String("cert", "../cert/service.pem", "TLS certificate file")
	keyFile  = flag.String("key", "../cert/service.key", "TLS private key file")
)

// server is used to implement sample.SampleServiceServer.
type server struct {
	pb.UnimplementedSampleServiceServer
}

// Greet implements sample.SampleServiceServer.
func (s *server) Greet(ctx context.Context, in *pb.SendMsg) (*pb.SendResp, error) {
	log.Printf("Received msg: %v", in.GetName())
	return &pb.SendResp{Message: "Hey " + in.GetName()}, nil
}

func main() {
	flag.Parse()

	// Load the server certificate and private key used for TLS.
	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("TLS setup failed: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Could not start the server: %v", err)
	}

	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterSampleServiceServer(s, &server{})
	log.Printf("Server started at: %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Could not start the server: %v", err)
	}
}
