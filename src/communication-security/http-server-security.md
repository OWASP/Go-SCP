HTTP Server Security
====================

Securing your Go HTTP server is critical for production deployments. This
section covers timeout configuration, denial-of-service protection, and
secure server setup.

The Default Server Problem
--------------------------

Go's convenience functions for starting HTTP servers are **unsafe for
production** because they have no timeout configuration:

```go
// INSECURE - Never use in production!
http.ListenAndServe(":8080", handler)        // No timeouts
http.ListenAndServeTLS(":443", cert, key, h) // No timeouts
http.Serve(listener, handler)                // No timeouts
```

Without timeouts, your server is vulnerable to:
- **Slowloris attacks**: Attackers open connections and send data very slowly
- **Resource exhaustion**: Connections accumulate until "too many open files"
- **Hanging connections**: Misbehaving clients can hold resources indefinitely

Secure Server Configuration
---------------------------

Always create a custom `http.Server` with explicit timeouts:

```go
package main

import (
    "log"
    "net/http"
    "time"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           mux,
        ReadHeaderTimeout: 10 * time.Second,  // Slowloris protection
        ReadTimeout:       30 * time.Second,  // Max time to read request
        WriteTimeout:      30 * time.Second,  // Max time to write response
        IdleTimeout:       120 * time.Second, // Keep-alive timeout
        MaxHeaderBytes:    1 << 20,           // 1 MB max header size
    }

    log.Printf("Server starting on %s", srv.Addr)
    log.Fatal(srv.ListenAndServe())
}
```

Understanding Timeouts
----------------------

Each timeout serves a specific security purpose:

| Timeout | Purpose | Recommended |
|---------|---------|-------------|
| `ReadHeaderTimeout` | Time to read request headers (Slowloris protection) | 10-30s |
| `ReadTimeout` | Time to read entire request including body | 30-60s |
| `WriteTimeout` | Time to write response (from end of header read) | 30-60s |
| `IdleTimeout` | Time between requests on keep-alive connections | 60-120s |

### Timeout Flow Diagram

```
Connection    ReadHeaderTimeout    ReadTimeout         WriteTimeout
Accepted  ──────────┬─────────────────┬──────────────────────┬──────
                    │                 │                      │
              Headers Read      Body Read            Response Written
```

Slowloris Attack Protection
---------------------------

The Slowloris attack works by:
1. Opening many connections to your server
2. Sending partial HTTP requests very slowly
3. Never completing requests, keeping connections open
4. Eventually exhausting server resources

**Protection requires `ReadHeaderTimeout`:**

```go
// VULNERABLE - No ReadHeaderTimeout
srv := &http.Server{
    ReadTimeout: 30 * time.Second, // Not enough!
}

// SECURE - ReadHeaderTimeout set
srv := &http.Server{
    ReadHeaderTimeout: 10 * time.Second, // Closes slow header sends
    ReadTimeout:       30 * time.Second,
}
```

Handler-Level Timeouts
----------------------

For fine-grained control, use `http.TimeoutHandler` to wrap specific handlers:

```go
// Wrap handler with 30-second timeout
handler := http.TimeoutHandler(myHandler, 30*time.Second, "Request timeout")

mux.Handle("/api/slow-endpoint", handler)
```

This returns HTTP 503 Service Unavailable if the handler exceeds the timeout.

### Per-Request Context Timeouts

For more control, use context timeouts within handlers:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Create timeout context for this request
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    // Use ctx for database calls, external APIs, etc.
    result, err := db.QueryContext(ctx, "SELECT ...")
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Request timeout", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

Streaming Response Considerations
---------------------------------

`WriteTimeout` is problematic for streaming responses (SSE, WebSockets, large
file downloads) because it's an absolute deadline, not an idle timeout.

**Options for streaming:**

1. **Disable WriteTimeout for streaming endpoints** (use separate server):
```go
// Streaming server without WriteTimeout
streamSrv := &http.Server{
    Addr:              ":8081",
    Handler:           streamMux,
    ReadHeaderTimeout: 10 * time.Second,
    // WriteTimeout intentionally omitted for streaming
}
```

2. **Use http.Hijacker** for manual connection control:
```go
func streamHandler(w http.ResponseWriter, r *http.Request) {
    hj, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    conn, _, err := hj.Hijack()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer conn.Close()

    // Manual connection management with per-write deadlines
    for data := range dataChannel {
        conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
        conn.Write(data)
    }
}
```

Connection Limits
-----------------

Limit concurrent connections to prevent resource exhaustion:

```go
import "golang.org/x/net/netutil"

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }

    // Limit to 1000 concurrent connections
    listener = netutil.LimitListener(listener, 1000)

    srv := &http.Server{
        Handler:           mux,
        ReadHeaderTimeout: 10 * time.Second,
        // ... other timeouts
    }

    log.Fatal(srv.Serve(listener))
}
```

Graceful Shutdown
-----------------

Implement graceful shutdown to complete in-flight requests:

```go
func main() {
    srv := &http.Server{
        Addr:              ":8080",
        Handler:           mux,
        ReadHeaderTimeout: 10 * time.Second,
        ReadTimeout:       30 * time.Second,
        WriteTimeout:      30 * time.Second,
    }

    // Start server in goroutine
    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    // Give outstanding requests 30 seconds to complete
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server stopped")
}
```

Complete Secure Server Example
------------------------------

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", homeHandler)
    mux.HandleFunc("/api/", apiHandler)

    // Wrap with security middleware
    handler := securityHeaders(mux)

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           handler,
        ReadHeaderTimeout: 10 * time.Second,
        ReadTimeout:       30 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       120 * time.Second,
        MaxHeaderBytes:    1 << 20, // 1 MB
    }

    // Graceful shutdown
    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        srv.Shutdown(ctx)
    }()

    log.Printf("Server starting on %s", srv.Addr)
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}

func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        next.ServeHTTP(w, r)
    })
}
```

References
----------

- [Cloudflare Go Timeout Guide][1]
- [Go net/http Server Documentation][2]
- [OWASP Denial of Service Cheat Sheet][3]

[1]: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
[2]: https://pkg.go.dev/net/http#Server
[3]: https://cheatsheetseries.owasp.org/cheatsheets/Denial_of_Service_Cheat_Sheet.html
