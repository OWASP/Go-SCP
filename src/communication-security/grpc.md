gRPC
==========
gRPC is a new open source high performance Remote Procedure Call(RPC) framework that can be used in number of scenarios such as distributed computing, mobile & web applications.
gRPC is one of the popular ways to make microservices implemented in different languages/technolgies to seemlessly interact with each other with protocol buffers. Go provides inbuilt support for [gRPC][1].

Let's look at how we can implement secure gRPC calls using TLS Encryption on both client and server side.

### Secure gRPC Server:
The complete code can be found [here][2].
```go
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

	//Reading the certificate key and private key needed for credentials
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
```

### Secure gRPC client:
The complete code can be found [here][3].
```go
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

```

**Important** : As mentioned in the previous sections that the TLS certificates should be valid, should not be expired and should be installed with intermediate certificates when required as recommended in OWASP SCP Quick
Reference Guide][4]. 
Also, an ideal gRPC client should avoid connecting to a gRPC server which is missing or has invalid certificates. The `InsecureSkipVerify` flag should never be set to false for services deployed in production.

[1]: https://pkg.go.dev/google.golang.org/grpc
[2]: ./grpc-code/grpc_server_secured/server.go
[3]: ./grpc-code/grpc_client_secured/client.go
[4]: https://www.owasp.org/images/0/08/OWASP_SCP_Quick_Reference_Guide_v2.pdf
