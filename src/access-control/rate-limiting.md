Rate Limiting
=============

Rate limiting is essential for preventing abuse, ensuring service stability,
and protecting against denial-of-service attacks. This section covers
implementing rate limiting in Go applications.

Why Rate Limiting Matters
-------------------------

Without rate limiting, your application is vulnerable to:
- **Brute force attacks**: Unlimited login attempts
- **API abuse**: Single clients consuming all resources
- **Denial of service**: Overwhelming your server with requests
- **Scraping**: Automated extraction of your data

Token Bucket Algorithm
----------------------

Go's `golang.org/x/time/rate` package implements the token bucket algorithm,
which allows bursts of traffic while maintaining an average rate limit.

```bash
go get golang.org/x/time/rate
```

Basic Rate Limiter
------------------

### Global Rate Limiting

Limit all requests to your server:

```go
package main

import (
    "net/http"
    "golang.org/x/time/rate"
)

// Allow 100 requests per second with burst of 200
var limiter = rate.NewLimiter(rate.Limit(100), 200)

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)

    // Apply rate limiting
    http.ListenAndServe(":8080", rateLimitMiddleware(mux))
}
```

Per-Client Rate Limiting
------------------------

Global rate limiting has a flaw: one abusive client can exhaust the limit for
everyone. Per-client limiting solves this:

```go
package main

import (
    "net/http"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

// Client stores rate limiter for each visitor
type Client struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

var (
    clients = make(map[string]*Client)
    mu      sync.RWMutex
)

// getClient returns the rate limiter for a client IP
func getClient(ip string) *rate.Limiter {
    mu.Lock()
    defer mu.Unlock()

    client, exists := clients[ip]
    if !exists {
        // 10 requests per second, burst of 30
        limiter := rate.NewLimiter(rate.Limit(10), 30)
        clients[ip] = &Client{limiter: limiter, lastSeen: time.Now()}
        return limiter
    }

    client.lastSeen = time.Now()
    return client.limiter
}

// cleanupClients removes old entries to prevent memory leaks
func cleanupClients() {
    for {
        time.Sleep(time.Minute)
        mu.Lock()
        for ip, client := range clients {
            if time.Since(client.lastSeen) > 3*time.Minute {
                delete(clients, ip)
            }
        }
        mu.Unlock()
    }
}

func perClientRateLimiter(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get client IP (handle proxies - see security note below)
        ip := r.RemoteAddr

        limiter := getClient(ip)
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
    // Start cleanup goroutine
    go cleanupClients()

    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)

    http.ListenAndServe(":8080", perClientRateLimiter(mux))
}
```

Secure Client IP Detection
--------------------------

**Security Warning**: Extracting client IPs is tricky behind proxies. Naive
implementations are vulnerable to spoofing.

```go
// INSECURE - Attacker can spoof X-Forwarded-For
func getIPInsecure(r *http.Request) string {
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        return strings.Split(xff, ",")[0] // Easily spoofed!
    }
    return r.RemoteAddr
}

// SECURE - Only trust X-Forwarded-For from known proxy
func getIPSecure(r *http.Request, trustedProxies map[string]bool) string {
    remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

    // Only trust headers if request came from known proxy
    if trustedProxies[remoteIP] {
        if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
            ips := strings.Split(xff, ",")
            // Get rightmost non-proxy IP
            for i := len(ips) - 1; i >= 0; i-- {
                ip := strings.TrimSpace(ips[i])
                if !trustedProxies[ip] {
                    return ip
                }
            }
        }
    }

    return remoteIP
}
```

Endpoint-Specific Limits
------------------------

Different endpoints may need different limits:

```go
type RateLimitConfig struct {
    Rate  rate.Limit
    Burst int
}

var endpointLimits = map[string]RateLimitConfig{
    "/api/login":    {Rate: 5, Burst: 10},   // Strict for auth
    "/api/search":   {Rate: 30, Burst: 50},  // Moderate
    "/api/public":   {Rate: 100, Burst: 200}, // Generous
}

func getEndpointLimiter(path string, ip string) *rate.Limiter {
    config, exists := endpointLimits[path]
    if !exists {
        config = RateLimitConfig{Rate: 50, Burst: 100} // Default
    }

    key := path + ":" + ip
    // ... get or create limiter for this key
}
```

Using Third-Party Libraries
---------------------------

### Tollbooth

Full-featured rate limiting middleware:

```bash
go get github.com/didip/tollbooth/v7
```

```go
package main

import (
    "net/http"
    "time"

    "github.com/didip/tollbooth/v7"
    "github.com/didip/tollbooth/v7/limiter"
)

func main() {
    // Create a limiter: 1 request per second per IP
    lmt := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{
        DefaultExpirationTTL: time.Hour,
    })

    // Configure what identifies a client
    lmt.SetIPLookup(limiter.IPLookup{
        Name:           "X-Real-IP",
        IndexFromRight: 0,
    })

    // Set custom message
    lmt.SetMessage("You have exceeded the rate limit")
    lmt.SetMessageContentType("text/plain; charset=utf-8")

    mux := http.NewServeMux()
    mux.Handle("/", tollbooth.LimitHandler(lmt, http.HandlerFunc(handler)))

    http.ListenAndServe(":8080", mux)
}
```

### httprate (Chi middleware)

For chi router users:

```bash
go get github.com/go-chi/httprate
```

```go
package main

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/httprate"
)

func main() {
    r := chi.NewRouter()

    // Global rate limit
    r.Use(httprate.LimitByIP(100, time.Minute))

    // Stricter limit for auth endpoints
    r.Group(func(r chi.Router) {
        r.Use(httprate.LimitByIP(10, time.Minute))
        r.Post("/login", loginHandler)
        r.Post("/register", registerHandler)
    })

    http.ListenAndServe(":8080", r)
}
```

Distributed Rate Limiting
-------------------------

For multi-server deployments, use Redis for shared state:

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func init() {
    rdb = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
}

func rateLimitRedis(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.Background()
        ip := r.RemoteAddr
        key := "ratelimit:" + ip

        // Increment counter
        count, err := rdb.Incr(ctx, key).Result()
        if err != nil {
            // Fail open or closed based on your requirements
            next.ServeHTTP(w, r)
            return
        }

        // Set expiry on first request
        if count == 1 {
            rdb.Expire(ctx, key, time.Minute)
        }

        // Check limit (100 requests per minute)
        if count > 100 {
            w.Header().Set("Retry-After", "60")
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

Response Headers
----------------

Communicate rate limit status to clients:

```go
func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        limiter := getClientLimiter(r)

        // Set rate limit headers
        w.Header().Set("X-RateLimit-Limit", "100")
        w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int(limiter.Tokens())))
        w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))

        if !limiter.Allow() {
            w.Header().Set("Retry-After", "60")
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

Best Practices
--------------

1. **Start conservative**: Begin with stricter limits; loosen based on data
2. **Monitor and alert**: Log rate limit hits to detect attacks
3. **Differentiate by endpoint**: Auth endpoints need stricter limits
4. **Handle burst traffic**: Allow reasonable bursts for legitimate use
5. **Fail safely**: Decide whether to fail open or closed on errors
6. **Use appropriate identifiers**: IP for anonymous, user ID for authenticated
7. **Clean up state**: Prevent memory leaks with periodic cleanup

References
----------

- [golang.org/x/time/rate Documentation][1]
- [Tollbooth Rate Limiter][2]
- [OWASP Blocking Brute Force Attacks][3]

[1]: https://pkg.go.dev/golang.org/x/time/rate
[2]: https://github.com/didip/tollbooth
[3]: https://owasp.org/www-community/controls/Blocking_Brute_Force_Attacks
