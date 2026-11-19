Logging
=======

Logging should always be handled by the application and should not rely on a
server configuration.

All logging should be implemented by a master routine on a trusted system, and
the developers should also ensure no sensitive data is included in the logs
(e.g. passwords, session information, system details, etc.), nor is there
any debugging or stack trace information.
Additionally, logging should cover both successful and unsuccessful security
events, with an emphasis on important log event data.

Important event data most commonly refers to all:

* Input validation failures.
* Authentication attempts, especially failures.
* Access control failures.
* Apparent tampering events, including unexpected changes to state data.
* Attempts to connect with invalid or expired session tokens.
* System exceptions.
* Administrative functions, including changes to security configuration
  settings.
* Backend TLS connection failures and cryptographic module failures.

Here's a simple log example which illustrates this:

```go
func main() {
    var buf bytes.Buffer
    var RoleLevel int

    logger := log.New(&buf, "logger: ", log.Lshortfile)

    fmt.Println("Please enter your user level.")
    fmt.Scanf("%d", &RoleLevel) //<--- example

    switch RoleLevel {
    case 1:
        // Log successful login
        logger.Printf("Login successful.")
        fmt.Print(&buf)
    case 2:
        // Log unsuccessful Login
        logger.Printf("Login unsuccessful - Insufficient access level.")
        fmt.Print(&buf)
     default:
        // Unspecified error
        logger.Print("Login error.")
        fmt.Print(&buf)
    }
}
```

It's also good practice to implement generic error messages, or custom error
pages, as a way to make sure that no information is leaked when an error
occurs.

---

[Go's log package][0], as per the documentation, "implements **simple**
logging". Some common and important features are missing, such as leveled
logging (e.g. `debug`, `info`, `warn`, `error`, `fatal`, `panic`) and formatters
support (e.g. logstash). These are two important features to make logs usable
(e.g. for integration with a Security Information and Event Management system).

### log/slog (Recommended - Go 1.21+)

As of Go 1.21, the standard library includes [log/slog][5], a structured logging
package. This is now the recommended approach for most Go applications:

```go
package main

import (
    "log/slog"
    "os"
)

func main() {
    // JSON handler for production (structured logging)
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    // Log with structured fields
    slog.Info("user login attempt",
        slog.String("username", "john"),
        slog.String("ip", "192.168.1.1"),
        slog.Bool("success", true),
    )

    // Log errors with context
    slog.Error("database connection failed",
        slog.String("host", "db.example.com"),
        slog.Int("port", 5432),
        slog.Any("error", err),
    )
}
```

Output (JSON):
```json
{"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"user login attempt","username":"john","ip":"192.168.1.1","success":true}
```

**Security considerations with slog:**
- Use `slog.String()` for user input to prevent log injection
- Never log sensitive data (passwords, tokens, PII)
- Configure appropriate log levels for production vs development

### Third-Party Alternatives

If you need features not in slog, or are using Go < 1.21:

* [Logrus][1] - https://github.com/Sirupsen/logrus (mature, widely used)
* [zerolog][6] - https://github.com/rs/zerolog (zero-allocation JSON logger)
* [zap][7] - https://github.com/uber-go/zap (high-performance structured logging)

Here's an important note regarding [Go's log package][0]: Fatal and Panic
functions have different behaviors after logging: Panic functions call `panic`
but Fatal functions call `os.Exit(1)` that **may terminate the program
preventing deferred statements to run, buffers to be flushed, and/or temporary
data to be removed**.

---

From the perspective of log access, only authorized individuals should have
access to the logs.
Developers should also make sure that a mechanism that allows for log
analysis is set in place, as well as guarantee that no untrusted data will
be executed as code in the intended log viewing software or interface.

Regarding allocated memory cleanup, Go has a built-in Garbage Collector for this
very purpose.

As a final step to guarantee log validity and integrity, a cryptographic
hash function should be used as an additional step to ensure no log
tampering has taken place.

```go
{...}
// Get our known log checksum from the checksum file.
stored, err := root.ReadFile(*checksumFile)
if err != nil {
    return fmt.Errorf("read checksum: %w", err)
}

// Compute our current log's SHA-256 hash.
b, err := ComputeSHA256(root, *logFile)
if err != nil {
    return fmt.Errorf("hash log: %w", err)
}

// Compare our calculated hash with our stored hash.
if hex.EncodeToString(b) != strings.TrimSpace(string(stored)) {
    return errors.New("file tampering detected")
}
_, _ = fmt.Fprintln(stdout, "Log integrity OK.")
{...}
```

Note: The `ComputeSHA256()` function calculates a file's SHA256. It's also
important to note that the log-file hashes must be stored in a safe place, and
compared with the current log hash to verify integrity before any updates to the
log. [Working demo available in the repository][4].

[0]: https://pkg.go.dev/log
[1]: https://github.com/Sirupsen/logrus
[2]: https://github.com/golang/glog
[3]: https://github.com/juju/loggo
[4]: ./assets/log-integrity.go
[5]: https://pkg.go.dev/log/slog
[6]: https://github.com/rs/zerolog
[7]: https://github.com/uber-go/zap
