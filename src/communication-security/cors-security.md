CORS Security
=============

Cross-Origin Resource Sharing (CORS) controls which domains can access your
API. Misconfigured CORS can expose your application to cross-site attacks.

Understanding CORS
------------------

Browsers enforce the Same-Origin Policy, blocking requests from different
origins. CORS headers tell browsers which cross-origin requests to allow.

**Origin** = Protocol + Domain + Port
- `https://example.com` and `https://api.example.com` are different origins
- `http://example.com` and `https://example.com` are different origins

The Security Risk
-----------------

```go
// INSECURE - Allows any origin with credentials
func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Credentials", "true") // DANGEROUS!
}
```

This allows any website to make authenticated requests to your API, enabling:
- **CSRF attacks**: Malicious sites can perform actions as logged-in users
- **Data theft**: Attackers can read sensitive user data
- **Session hijacking**: Credentials can be stolen

**Note**: Browsers block `Allow-Credentials: true` with `Allow-Origin: *`, but
some implementations have bugs or workarounds.

Secure CORS Configuration
-------------------------

### Manual Implementation

```go
func corsMiddleware(allowedOrigins map[string]bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            // Only allow specific origins
            if allowedOrigins[origin] {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Credentials", "true")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
            }

            // Handle preflight requests
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

func main() {
    allowedOrigins := map[string]bool{
        "https://example.com":     true,
        "https://app.example.com": true,
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/api/", apiHandler)

    http.ListenAndServe(":8080", corsMiddleware(allowedOrigins)(mux))
}
```

Using rs/cors Library
---------------------

The `rs/cors` library is the most popular CORS middleware for Go:

```bash
go get github.com/rs/cors
```

### Basic Secure Configuration

```go
package main

import (
    "net/http"
    "github.com/rs/cors"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/", apiHandler)

    // Secure CORS configuration
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
        ExposedHeaders:   []string{"X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           86400, // Preflight cache for 24 hours
    })

    http.ListenAndServe(":8080", c.Handler(mux))
}
```

### Dynamic Origin Validation

For complex origin requirements, use `AllowOriginFunc`:

```go
import (
    "net/url"
    "strings"
)

c := cors.New(cors.Options{
    AllowOriginFunc: func(origin string) bool {
        // Parse the origin
        u, err := url.Parse(origin)
        if err != nil {
            return false
        }

        // Only allow HTTPS
        if u.Scheme != "https" {
            return false
        }

        // Allow subdomains of example.com
        if u.Host == "example.com" || strings.HasSuffix(u.Host, ".example.com") {
            return true
        }

        return false
    },
    AllowCredentials: true,
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Content-Type", "Authorization"},
})
```

**Security Warning**: Be very careful with `AllowOriginFunc`. Overly permissive
validation can lead to security vulnerabilities.

Framework-Specific Examples
---------------------------

### Gin

```go
import "github.com/gin-contrib/cors"

func main() {
    r := gin.Default()

    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"https://example.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

    r.GET("/api/data", handler)
    r.Run(":8080")
}
```

### Chi

```go
import "github.com/go-chi/cors"

func main() {
    r := chi.NewRouter()

    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"https://example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    r.Get("/api/data", handler)
    http.ListenAndServe(":8080", r)
}
```

### Fiber

```go
import "github.com/gofiber/fiber/v2/middleware/cors"

func main() {
    app := fiber.New()

    app.Use(cors.New(cors.Config{
        AllowOrigins:     "https://example.com,https://app.example.com",
        AllowMethods:     "GET,POST,PUT,DELETE",
        AllowHeaders:     "Content-Type,Authorization",
        AllowCredentials: true,
        MaxAge:           86400,
    }))

    app.Get("/api/data", handler)
    app.Listen(":8080")
}
```

Common Mistakes
---------------

### Mistake 1: Reflecting Origin Without Validation

```go
// INSECURE - Reflects any origin
func handler(w http.ResponseWriter, r *http.Request) {
    origin := r.Header.Get("Origin")
    w.Header().Set("Access-Control-Allow-Origin", origin) // Any origin!
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}
```

### Mistake 2: Overly Permissive Subdomain Matching

```go
// INSECURE - Attacker can register evil.example.com.attacker.com
func isAllowed(origin string) bool {
    return strings.Contains(origin, "example.com") // Too loose!
}

// SECURE - Proper suffix check
func isAllowed(origin string) bool {
    u, _ := url.Parse(origin)
    return u.Host == "example.com" || strings.HasSuffix(u.Host, ".example.com")
}
```

### Mistake 3: Allowing null Origin

```go
// INSECURE - 'null' origin comes from sandboxed iframes and file:// URLs
allowedOrigins := []string{"https://example.com", "null"} // Dangerous!
```

### Mistake 4: Missing Vary Header

```go
// When origin-specific responses are cached, include Vary header
w.Header().Set("Vary", "Origin")
w.Header().Set("Access-Control-Allow-Origin", origin)
```

Development vs Production
-------------------------

Use different configurations for development:

```go
func getCORSConfig() cors.Options {
    if os.Getenv("ENV") == "development" {
        return cors.Options{
            AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
            AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders:   []string{"*"},
            AllowCredentials: true,
        }
    }

    // Production - strict configuration
    return cors.Options{
        AllowedOrigins:   []string{"https://example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:           86400,
    }
}
```

Testing CORS Configuration
--------------------------

```bash
# Test preflight request
curl -X OPTIONS https://api.example.com/data \
  -H "Origin: https://example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v

# Test actual request
curl https://api.example.com/data \
  -H "Origin: https://example.com" \
  -v

# Should see:
# Access-Control-Allow-Origin: https://example.com
# Access-Control-Allow-Credentials: true
```

Best Practices Summary
----------------------

| Practice | Recommendation |
|----------|----------------|
| Origins | Explicit allowlist, never `*` with credentials |
| Methods | Only methods your API actually uses |
| Headers | Only headers your API actually needs |
| Credentials | Only enable if cookies/auth headers needed |
| Max-Age | Cache preflight (86400 = 24 hours) |
| Validation | Use exact matching or careful suffix checks |
| Vary Header | Include when responses vary by origin |

References
----------

- [MDN CORS Documentation][1]
- [OWASP CORS Cheat Sheet][2]
- [rs/cors Library][3]

[1]: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
[2]: https://cheatsheetseries.owasp.org/cheatsheets/Cross-Origin_Resource_Sharing_Cheat_Sheet.html
[3]: https://github.com/rs/cors
