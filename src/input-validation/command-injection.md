Command Injection Prevention
============================

Command injection occurs when an attacker can execute arbitrary system commands
through your application by manipulating input that's passed to a shell or
system command.

The Vulnerability
-----------------

```go
// VULNERABLE - Direct shell execution with user input
func pingHandler(w http.ResponseWriter, r *http.Request) {
    host := r.URL.Query().Get("host")

    // Attacker sends: ?host=google.com; cat /etc/passwd
    cmd := exec.Command("sh", "-c", "ping -c 1 "+host)
    output, _ := cmd.CombinedOutput()
    w.Write(output)
}
```

The attacker can terminate the ping command with `;` and execute any command.

Prevention Strategies
---------------------

### Strategy 1: Avoid Shell Invocation (Recommended)

Pass arguments directly to the command without using a shell:

```go
// SECURE - No shell, arguments passed separately
func pingHandler(w http.ResponseWriter, r *http.Request) {
    host := r.URL.Query().Get("host")

    // Arguments are passed directly, not interpreted by shell
    cmd := exec.Command("ping", "-c", "1", host)
    output, err := cmd.CombinedOutput()
    if err != nil {
        http.Error(w, "Ping failed", http.StatusBadRequest)
        return
    }
    w.Write(output)
}
```

This is secure because:
- No shell (`sh`, `bash`) is invoked
- Special characters like `;`, `|`, `&&` are treated as literal strings
- The `host` value becomes a single argument to `ping`

### Strategy 2: Input Validation with Allowlist

When you must accept user input, validate strictly:

```go
import (
    "net"
    "regexp"
)

// Validate hostname or IP address
func isValidHost(host string) bool {
    // Check if it's a valid IP
    if ip := net.ParseIP(host); ip != nil {
        return true
    }

    // Check if it's a valid hostname (alphanumeric, dots, hyphens)
    validHostname := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
    return validHostname.MatchString(host) && len(host) <= 253
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
    host := r.URL.Query().Get("host")

    if !isValidHost(host) {
        http.Error(w, "Invalid host", http.StatusBadRequest)
        return
    }

    cmd := exec.Command("ping", "-c", "1", host)
    output, _ := cmd.CombinedOutput()
    w.Write(output)
}
```

### Strategy 3: Command Allowlist

Only allow specific, predefined commands:

```go
var allowedCommands = map[string][]string{
    "status":  {"systemctl", "status", "myservice"},
    "restart": {"systemctl", "restart", "myservice"},
    "logs":    {"journalctl", "-u", "myservice", "-n", "100"},
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
    action := r.URL.Query().Get("action")

    cmdArgs, allowed := allowedCommands[action]
    if !allowed {
        http.Error(w, "Unknown action", http.StatusBadRequest)
        return
    }

    cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        http.Error(w, "Command failed", http.StatusInternalServerError)
        return
    }
    w.Write(output)
}
```

Flag Injection Prevention
-------------------------

Some commands interpret arguments starting with `-` as flags. An attacker might
abuse this:

```go
// VULNERABLE to flag injection
func gitLogHandler(w http.ResponseWriter, r *http.Request) {
    branch := r.URL.Query().Get("branch")

    // Attacker sends: ?branch=--exec=malicious_command
    // Some git commands execute arbitrary code with certain flags
    cmd := exec.Command("git", "log", branch)
    // ...
}
```

**Solution**: Use `--` to separate options from arguments:

```go
// SECURE - "--" prevents flag injection
func gitLogHandler(w http.ResponseWriter, r *http.Request) {
    branch := r.URL.Query().Get("branch")

    // "--" ensures branch is treated as an argument, not a flag
    cmd := exec.Command("git", "log", "--", branch)
    output, _ := cmd.CombinedOutput()
    w.Write(output)
}
```

Also validate that arguments don't start with `-`:

```go
func sanitizeArg(arg string) (string, error) {
    if strings.HasPrefix(arg, "-") {
        return "", errors.New("argument cannot start with dash")
    }
    return arg, nil
}
```

Environment Variable Injection
------------------------------

Be careful with environment variables in commands:

```go
// VULNERABLE - User controls environment variable
func handler(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Query().Get("path")

    cmd := exec.Command("myprogram")
    cmd.Env = append(os.Environ(), "USER_PATH="+path)
    // If myprogram uses USER_PATH unsafely, injection is possible
}
```

**Solution**: Sanitize or validate environment values:

```go
cmd.Env = []string{
    "PATH=/usr/bin:/bin",           // Minimal PATH
    "HOME=/tmp",                    // Controlled HOME
    "USER_PATH=" + sanitize(path),  // Sanitized user input
}
```

Working Directory Attacks
-------------------------

Be cautious when setting the working directory:

```go
// VULNERABLE - User controls working directory
func handler(w http.ResponseWriter, r *http.Request) {
    dir := r.URL.Query().Get("dir")

    cmd := exec.Command("ls", "-la")
    cmd.Dir = dir  // Attacker can list any directory!
    output, _ := cmd.CombinedOutput()
    w.Write(output)
}

// SECURE - Validate and restrict directory
func handler(w http.ResponseWriter, r *http.Request) {
    dir := r.URL.Query().Get("dir")

    // Only allow subdirectories of /data
    fullPath := filepath.Join("/data", filepath.Clean(dir))
    if !strings.HasPrefix(fullPath, "/data/") {
        http.Error(w, "Invalid directory", http.StatusBadRequest)
        return
    }

    cmd := exec.Command("ls", "-la")
    cmd.Dir = fullPath
    output, _ := cmd.CombinedOutput()
    w.Write(output)
}
```

Timeout and Resource Limits
---------------------------

Always set timeouts on command execution:

```go
func runWithTimeout(name string, args ...string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, name, args...)

    // Limit output size to prevent memory exhaustion
    var stdout, stderr bytes.Buffer
    stdout.Grow(1 << 20) // 1 MB max
    stderr.Grow(1 << 20)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if ctx.Err() == context.DeadlineExceeded {
        return nil, errors.New("command timed out")
    }

    return stdout.Bytes(), err
}
```

Complete Secure Example
-----------------------

```go
package main

import (
    "bytes"
    "context"
    "net"
    "net/http"
    "os/exec"
    "regexp"
    "time"
)

var validHostPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\.\-]*[a-zA-Z0-9]$`)

func isValidHost(host string) bool {
    if len(host) == 0 || len(host) > 253 {
        return false
    }
    if ip := net.ParseIP(host); ip != nil {
        return true
    }
    return validHostPattern.MatchString(host)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
    host := r.URL.Query().Get("host")

    // 1. Validate input
    if !isValidHost(host) {
        http.Error(w, "Invalid host", http.StatusBadRequest)
        return
    }

    // 2. Create command with timeout context
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()

    // 3. Execute without shell, with explicit arguments
    cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "5", "--", host)

    // 4. Limit output size
    var stdout bytes.Buffer
    stdout.Grow(4096)
    cmd.Stdout = &stdout

    // 5. Run with restricted environment
    cmd.Env = []string{"PATH=/bin:/usr/bin"}

    if err := cmd.Run(); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            http.Error(w, "Request timed out", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, "Ping failed", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write(stdout.Bytes())
}
```

Alternatives to exec.Command
----------------------------

Consider native Go implementations instead of shell commands:

| Shell Command | Go Alternative |
|---------------|----------------|
| `cat file` | `os.ReadFile()` |
| `ls dir` | `os.ReadDir()` |
| `grep pattern file` | `regexp` + `bufio.Scanner` |
| `curl url` | `net/http` client |
| `ping host` | `github.com/go-ping/ping` |
| `dig domain` | `net.LookupHost()` |

References
----------

- [OWASP Command Injection][1]
- [Semgrep Go Command Injection Cheat Sheet][2]
- [Go exec Package Documentation][3]

[1]: https://owasp.org/www-community/attacks/Command_Injection
[2]: https://semgrep.dev/docs/cheat-sheets/go-command-injection
[3]: https://pkg.go.dev/os/exec
