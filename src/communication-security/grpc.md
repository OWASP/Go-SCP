gRPC
====

[gRPC][1] is an open source, high performance Remote Procedure Call (RPC)
framework. It is a popular way for microservices written in different languages
to talk to each other using Protocol Buffers, and it is also used for mobile and
web backends. Go support is provided by the [`google.golang.org/grpc`][1]
module.

Like any other network channel, gRPC traffic must be encrypted with TLS, and
clients must verify the identity of the server they connect to. Let's look at
how to do that on both the server and the client side.

The complete, buildable example lives in the [grpc-code][2] directory. Both
programs expect TLS material in a `cert` directory next to them; you can create
a test CA and a server certificate for `localhost` with OpenSSL:

```bash
mkdir cert && cd cert
openssl req -x509 -newkey rsa:2048 -nodes -keyout ca.key -out ca.cert \
    -days 30 -subj "/CN=Test CA"
openssl req -newkey rsa:2048 -nodes -keyout service.key -out service.csr \
    -subj "/CN=localhost"
printf "subjectAltName=DNS:localhost\n" > service.ext
openssl x509 -req -in service.csr -CA ca.cert -CAkey ca.key -CAcreateserial \
    -out service.pem -days 30 -extfile service.ext
```

**Note:** These certificates are only for local testing. In production, use
certificates issued by your organization's CA or a public CA, and never commit
private keys to source control.

## Secure gRPC Server

The complete code can be found [here][3].

```go
var (
    port     = flag.Int("port", 10001, "The server port")
    certFile = flag.String("cert", "../cert/service.pem", "TLS certificate file")
    keyFile  = flag.String("key", "../cert/service.key", "TLS private key file")
)

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
```

## Secure gRPC Client

The complete code can be found [here][4].

```go
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
```

If the server presents a certificate that was not signed by the trusted CA,
the call fails with an error such as `x509: certificate signed by unknown
authority`, which is exactly what you want.

**Important:** As mentioned in the [HTTP/TLS section][5], TLS certificates
should be valid, not expired, and installed with intermediate certificates when
required, as recommended in the [OWASP SCP Quick Reference Guide][6]. A gRPC
client should refuse to connect to a server whose certificate is missing or
invalid, so `InsecureSkipVerify` must never be set to `true` in production, and
`grpc.WithTransportCredentials(insecure.NewCredentials())` must not be used for
traffic that leaves the host.

[1]: https://pkg.go.dev/google.golang.org/grpc
[2]: ./grpc-code/
[3]: ./grpc-code/grpc_server_secured/server.go
[4]: ./grpc-code/grpc_client_secured/client.go
[5]: http-tls.md
[6]: https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/
