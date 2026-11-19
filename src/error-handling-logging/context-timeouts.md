Context Timeouts and Cancellation
==================================

Go's `context` package is essential for managing request lifecycles, preventing
resource exhaustion, and ensuring your application remains responsive under
load or attack.

Why Context Matters for Security
--------------------------------

Without proper context handling:
- **Resource exhaustion**: Slow external services can hang your application
- **Denial of Service**: Attackers can tie up server resources indefinitely
- **Goroutine leaks**: Abandoned operations consume memory forever
- **Cascading failures**: One slow service blocks everything downstream

Basic Context Patterns
----------------------

### HTTP Handlers

Always use the request's context for downstream operations:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // The context is canceled when:
    // - Client disconnects
    // - Request times out (if http.TimeoutHandler is used)
    // - Server is shutting down

    result, err := doWork(ctx)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            // Client disconnected - don't bother responding
            return
        }
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "Request timeout", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, "Error", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(result)
}
```

### Adding Timeouts

Wrap operations with timeout contexts:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Create a timeout context from the request context
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel() // Always call cancel to release resources

    result, err := slowOperation(ctx)
    // ...
}
```

Database Operations
-------------------

Always use context-aware database methods:

```go
func getUserByID(ctx context.Context, db *sql.DB, id int64) (*User, error) {
    // Use QueryRowContext instead of QueryRow
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    var user User
    err := db.QueryRowContext(ctx,
        "SELECT id, name, email FROM users WHERE id = $1", id,
    ).Scan(&user.ID, &user.Name, &user.Email)

    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("database query timed out: %w", err)
        }
        return nil, err
    }

    return &user, nil
}

// For multiple rows
func getUsers(ctx context.Context, db *sql.DB) ([]User, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    rows, err := db.QueryContext(ctx, "SELECT id, name, email FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        // Check context between iterations for long result sets
        if ctx.Err() != nil {
            return nil, ctx.Err()
        }

        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            return nil, err
        }
        users = append(users, u)
    }

    return users, rows.Err()
}
```

External HTTP Calls
-------------------

Always set timeouts on outbound HTTP requests:

```go
// INSECURE - No timeout, can hang forever
func fetchDataInsecure(url string) ([]byte, error) {
    resp, err := http.Get(url)
    // ...
}

// SECURE - Context with timeout
func fetchData(ctx context.Context, url string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("request to %s timed out", url)
        }
        return nil, err
    }
    defer resp.Body.Close()

    // Also limit response body size
    return io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB max
}
```

### Custom HTTP Client with Timeouts

```go
var httpClient = &http.Client{
    Timeout: 30 * time.Second, // Overall timeout
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,  // Connection timeout
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   5 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        IdleConnTimeout:       90 * time.Second,
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
    },
}

func fetchWithClient(ctx context.Context, url string) ([]byte, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := httpClient.Do(req)
    // ...
}
```

Goroutine Management
--------------------

Use context to prevent goroutine leaks:

```go
// INSECURE - Goroutine leak if context canceled
func processItemsInsecure(items []Item) {
    for _, item := range items {
        go processItem(item) // These goroutines may run forever
    }
}

// SECURE - Goroutines respect context
func processItems(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        item := item // Capture for goroutine
        g.Go(func() error {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                return processItem(ctx, item)
            }
        })
    }

    return g.Wait()
}
```

### Worker Pool with Context

```go
func workerPool(ctx context.Context, jobs <-chan Job, results chan<- Result) {
    for {
        select {
        case <-ctx.Done():
            return // Clean shutdown
        case job, ok := <-jobs:
            if !ok {
                return // Channel closed
            }
            result := processJob(ctx, job)
            select {
            case results <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}
```

Context Values (Use Sparingly)
------------------------------

Context values can pass request-scoped data, but use them carefully:

```go
type contextKey string

const (
    requestIDKey contextKey = "requestID"
    userIDKey    contextKey = "userID"
)

// Set values
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), requestIDKey, requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Get values (with type safety)
func getRequestID(ctx context.Context) string {
    if v := ctx.Value(requestIDKey); v != nil {
        return v.(string)
    }
    return ""
}
```

**Best Practices for Context Values:**
- Use custom types for keys to avoid collisions
- Only store request-scoped data (request ID, user ID)
- Never store optional parameters or configuration
- Always provide accessor functions

Timeout Hierarchy
-----------------

Apply timeouts at multiple levels:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Level 1: Request timeout (from http.TimeoutHandler or server config)
    // r.Context() already has deadline

    // Level 2: Operation timeout (more specific)
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    // Level 3: Sub-operation timeouts
    dbCtx, dbCancel := context.WithTimeout(ctx, 2*time.Second)
    defer dbCancel()
    user, err := db.GetUser(dbCtx, userID)
    if err != nil {
        // ...
    }

    apiCtx, apiCancel := context.WithTimeout(ctx, 3*time.Second)
    defer apiCancel()
    data, err := externalAPI.Fetch(apiCtx, user.ID)
    // ...
}
```

Error Handling Patterns
-----------------------

```go
func handleContextError(err error) (int, string) {
    switch {
    case errors.Is(err, context.Canceled):
        // Client disconnected - they won't see this anyway
        return http.StatusClientClosedRequest, "Client disconnected"

    case errors.Is(err, context.DeadlineExceeded):
        // Our timeout was exceeded
        return http.StatusGatewayTimeout, "Request timed out"

    default:
        return http.StatusInternalServerError, "Internal error"
    }
}
```

Complete Example
----------------

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "net/http"
    "time"
)

type Service struct {
    db     *sql.DB
    client *http.Client
}

func (s *Service) GetUserData(ctx context.Context, userID int64) (*UserData, error) {
    // Overall operation timeout
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // Fetch user from database
    dbCtx, dbCancel := context.WithTimeout(ctx, 3*time.Second)
    defer dbCancel()

    user, err := s.getUserFromDB(dbCtx, userID)
    if err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }

    // Fetch additional data from external API
    apiCtx, apiCancel := context.WithTimeout(ctx, 5*time.Second)
    defer apiCancel()

    extra, err := s.fetchExternalData(apiCtx, user.ExternalID)
    if err != nil {
        // Log but don't fail - external data is optional
        log.Printf("external API error: %v", err)
    }

    return &UserData{User: user, Extra: extra}, nil
}

func (s *Service) handler(w http.ResponseWriter, r *http.Request) {
    data, err := s.GetUserData(r.Context(), getUserID(r))
    if err != nil {
        status, msg := handleContextError(err)
        http.Error(w, msg, status)
        return
    }

    json.NewEncoder(w).Encode(data)
}
```

References
----------

- [Go Context Package Documentation][1]
- [Go Blog: Context][2]
- [Go Concurrency Patterns: Context][3]

[1]: https://pkg.go.dev/context
[2]: https://go.dev/blog/context
[3]: https://go.dev/blog/context
