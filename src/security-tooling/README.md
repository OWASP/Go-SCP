Security Tooling
================

Modern Go development benefits from a rich ecosystem of security tools. This
section covers essential tools that should be part of every Go project's
security workflow.

Vulnerability Scanning with govulncheck
---------------------------------------

[govulncheck][1] is the official Go vulnerability scanner, maintained by the Go
security team. It analyzes your code and dependencies to find known
vulnerabilities.

### Installation

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Usage

```bash
# Scan the current module
govulncheck ./...

# Scan a specific package
govulncheck ./pkg/...

# Scan in binary mode (analyzes compiled binary)
govulncheck -mode=binary ./myapp
```

### CI/CD Integration

Add govulncheck to your CI pipeline to catch vulnerabilities early:

```yaml
# GitHub Actions example
- name: Install govulncheck
  run: go install golang.org/x/vuln/cmd/govulncheck@latest

- name: Run govulncheck
  run: govulncheck ./...
```

govulncheck is more accurate than generic vulnerability scanners because it
analyzes actual code paths - it won't report vulnerabilities in code your
application doesn't use.

Static Analysis with gosec
--------------------------

[gosec][2] (Go Security Checker) scans Go source code for common security
issues such as:

- Hardcoded credentials
- SQL injection
- Command injection
- Weak cryptographic primitives
- Insecure file permissions

### Installation

```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

### Usage

```bash
# Scan the current directory
gosec ./...

# Output as JSON
gosec -fmt=json -out=results.json ./...

# Exclude specific rules
gosec -exclude=G104 ./...
```

### Common Rules

| Rule ID | Description |
|---------|-------------|
| G101 | Hardcoded credentials |
| G102 | Bind to all interfaces |
| G103 | Unsafe block |
| G104 | Audit errors not checked |
| G201-G203 | SQL injection |
| G301-G307 | File permission issues |
| G401-G407 | Weak cryptographic primitives |
| G501-G505 | Import blocklist (crypto/md5, etc.) |

Race Detection
--------------

Go's built-in race detector finds data races in concurrent code. Data races
are a common source of security vulnerabilities and unpredictable behavior.

### Usage

```bash
# Run tests with race detection
go test -race ./...

# Build with race detection (for testing only - adds overhead)
go build -race -o myapp-race ./cmd/myapp

# Run the race-enabled binary
./myapp-race
```

### Example

```go
// INSECURE - Data race
var counter int

func increment() {
    counter++ // Race condition!
}

func main() {
    go increment()
    go increment()
    // counter may be 1 or 2 depending on timing
}

// SECURE - Use synchronization
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}
```

**Important:** The race detector adds significant overhead (2-10x slower,
5-10x more memory). Use it in testing, not production.

Fuzz Testing
------------

Go 1.18 introduced native fuzzing support. Fuzzing automatically generates
test inputs to find edge cases and security issues like buffer overflows,
panics, and injection vulnerabilities.

### Writing a Fuzz Test

```go
// In your_test.go file
func FuzzParseInput(f *testing.F) {
    // Add seed corpus (known valid inputs)
    f.Add("valid input")
    f.Add("another valid input")
    f.Add("")
    f.Add("<script>alert(1)</script>")

    f.Fuzz(func(t *testing.T, input string) {
        // Call the function under test
        result, err := ParseInput(input)

        // The fuzzer will try to find inputs that cause:
        // - Panics
        // - Hangs
        // - Unexpected behavior

        if err != nil {
            return // Expected errors are fine
        }

        // Add invariant checks
        if result == nil {
            t.Error("result should not be nil when err is nil")
        }
    })
}
```

### Running Fuzz Tests

```bash
# Run for 30 seconds
go test -fuzz=FuzzParseInput -fuzztime=30s ./...

# Run until stopped (Ctrl+C)
go test -fuzz=FuzzParseInput ./...

# Run with specific seed corpus
go test -fuzz=FuzzParseInput -fuzztime=1m ./... -test.fuzzcachedir=testdata/fuzz
```

### Security-Focused Fuzzing

Focus on parsing and input handling functions:

```go
func FuzzJSONDecoder(f *testing.F) {
    f.Add([]byte(`{"key": "value"}`))
    f.Add([]byte(`{"key": 123}`))
    f.Add([]byte(`[1, 2, 3]`))

    f.Fuzz(func(t *testing.T, data []byte) {
        var result map[string]interface{}
        // Should not panic on any input
        _ = json.Unmarshal(data, &result)
    })
}
```

Supply Chain Security
---------------------

Securing your dependencies is critical. Go provides several mechanisms for
supply chain security.

### Module Verification

```bash
# Verify all dependencies match go.sum checksums
go mod verify

# Download and verify all dependencies
go mod download -x
```

### Environment Variables

Configure Go's module behavior for security:

```bash
# Use the public module proxy (default)
export GOPROXY="https://proxy.golang.org,direct"

# Use checksum database for verification (default)
export GOSUMDB="sum.golang.org"

# For private modules, configure GOPRIVATE
export GOPRIVATE="github.com/your-org/*"
```

### SBOM Generation

Generate a Software Bill of Materials for your project:

```bash
# Using go version -m (basic)
go version -m ./myapp

# Using cyclonedx-gomod (detailed SBOM)
go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
cyclonedx-gomod mod -json -output sbom.json
```

### Dependency Updates

Regularly update dependencies and audit for vulnerabilities:

```bash
# Check for available updates
go list -m -u all

# Update all dependencies
go get -u ./...

# Update to latest minor/patch versions only
go get -u=patch ./...

# Then scan for vulnerabilities
govulncheck ./...
```

### Vendoring (Optional)

For additional supply chain control, vendor your dependencies:

```bash
# Create vendor directory
go mod vendor

# Build using vendor
go build -mod=vendor ./...
```

Recommended CI Pipeline
-----------------------

A security-focused CI pipeline should include:

```yaml
name: Security Checks

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Verify dependencies
        run: go mod verify

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Run gosec
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec ./...

      - name: Run tests with race detector
        run: go test -race ./...

      - name: Run fuzz tests (time-limited)
        run: |
          go test -fuzz=. -fuzztime=30s ./...
```

[1]: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
[2]: https://github.com/securego/gosec
