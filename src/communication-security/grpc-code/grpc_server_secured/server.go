package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "pentest/grpc/samplebuff"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	port = flag.Int("port", 10001, "The server port")
)

// server is used to implement sample.GreeterServer.
type server struct {
	pb.UnimplementedSampleServiceServer
}

// Greet implements sample.GreeterServer
func (s *server) Greet(ctx context.Context, in *pb.SendMsg) (*pb.SendResp, error) {
	log.Printf("Received msg: %v", in.GetName())
	return &pb.SendResp{Message: "Hey " + in.GetName()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Could not start the server: %v", err)
	}

	//Configuring the certificates
	creds, err := credentials.NewServerTLSFromFile("../cert/service.pem", "../cert/service.key")

	if err != nil {
		log.Fatalf("TLS setup failed: %v", err)
	}

	s := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterSampleServiceServer(s, &server{})
	log.Printf("Server started at: %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Could not start the server: %v", err)
	}
}
