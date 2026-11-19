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

The complete code is in [`grpc_server_secured/main.go`][3]. As in the other
examples, `main` only sets up signal handling and calls `run`, which takes its
arguments and output streams as parameters so it can be tested.

```go
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
```

## Secure gRPC Client

The complete code is in [`grpc_client_secured/main.go`][4].

```go
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
```

If the server presents a certificate that was not signed by the trusted CA,
the call fails with an error such as `x509: certificate signed by unknown
authority`, which is exactly what you want.
The client's [tests][7] check exactly this: the call succeeds against a server
whose certificate is signed by the trusted CA, and fails against one signed by
an unknown CA.

**Important:** As mentioned in the [HTTP/TLS section][5], TLS certificates
should be valid, not expired, and installed with intermediate certificates when
required, as recommended in the [OWASP SCP Quick Reference Guide][6]. A gRPC
client should refuse to connect to a server whose certificate is missing or
invalid, so `InsecureSkipVerify` must never be set to `true` in production, and
`grpc.WithTransportCredentials(insecure.NewCredentials())` must not be used for
traffic that leaves the host.

[1]: https://pkg.go.dev/google.golang.org/grpc
[2]: ./grpc-code/
[3]: ./grpc-code/grpc_server_secured/main.go
[4]: ./grpc-code/grpc_client_secured/main.go
[5]: http-tls.md
[6]: https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/
[7]: ./grpc-code/grpc_client_secured/main_test.go
