Security Scanning Integration
=============================

This section covers integrating security scanning tools into your Go development
workflow. Automated scanning catches vulnerabilities before they reach production.

Essential Security Tools
------------------------

### govulncheck (Official Go Tool)

The official vulnerability scanner from the Go team:

```bash
# Install
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan current module
govulncheck ./...

# Scan specific package
govulncheck ./cmd/server

# Output JSON for CI/CD
govulncheck -json ./...

# Scan binary
govulncheck -mode=binary ./myapp
```

govulncheck checks against the Go vulnerability database and only reports
vulnerabilities in code paths actually used by your application.

### gosec (Static Analysis)

Security-focused static analyzer:

```bash
# Install
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Basic scan
gosec ./...

# Output formats
gosec -fmt=json -out=results.json ./...
gosec -fmt=sarif -out=results.sarif ./...
gosec -fmt=html -out=results.html ./...

# Exclude specific rules
gosec -exclude=G104,G304 ./...

# Only specific rules
gosec -include=G101,G102 ./...
```

Common gosec rules:

| Rule | Description |
|------|-------------|
| G101 | Hardcoded credentials |
| G102 | Bind to all interfaces |
| G104 | Unhandled errors |
| G107 | URL provided to HTTP request from variable |
| G201 | SQL query with string format |
| G202 | SQL query with string concat |
| G301 | Poor file permissions |
| G304 | File path from variable (traversal) |
| G401 | Use of weak crypto (DES, MD5, SHA1) |
| G501 | Blocklisted imports |

### golangci-lint (Meta Linter)

Aggregates multiple linters including security tools:

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run with security linters
golangci-lint run --enable=gosec,govet,staticcheck
```

Configure `.golangci.yml`:

```yaml
linters:
  enable:
    - gosec        # Security scanner
    - govet        # Go vet
    - staticcheck  # Static analysis
    - errcheck     # Unchecked errors
    - ineffassign  # Unused assignments
    - typecheck    # Type checking
    - bodyclose    # Unclosed HTTP bodies
    - noctx        # HTTP requests without context

linters-settings:
  gosec:
    includes:
      - G101  # Hardcoded credentials
      - G102  # Bind to all interfaces
      - G103  # Audit unsafe block usage
      - G104  # Audit errors not checked
      - G106  # Audit SSH InsecureIgnoreHostKey
      - G107  # URL provided to HTTP request
      - G108  # Profiling endpoint exposed
      - G109  # Integer overflow conversion
      - G110  # Potential DoS from decompression
      - G201  # SQL string formatting
      - G202  # SQL string concatenation
      - G203  # Unescaped HTML templates
      - G204  # Audit command execution
      - G301  # Poor file permissions (mkdir)
      - G302  # Poor file permissions (chmod)
      - G303  # Creating file with predictable name
      - G304  # File path from tainted input
      - G305  # Zip slip traversal
      - G306  # Poor permissions on WriteFile
      - G307  # Deferring Close on writable file
      - G401  # Use of weak crypto
      - G402  # TLS MinVersion too low
      - G403  # RSA key < 2048 bits
      - G404  # Use of weak PRNG
      - G501  # Blocklisted import crypto/md5
      - G502  # Blocklisted import crypto/des
      - G503  # Blocklisted import crypto/rc4
      - G504  # Blocklisted import net/http/cgi
      - G505  # Blocklisted import crypto/sha1
      - G601  # Implicit memory aliasing
      - G602  # Slice access out of bounds

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

Go Fuzzing
----------

Go 1.18+ includes native fuzzing support:

```go
// user_test.go
package user

import (
    "testing"
)

func FuzzValidateEmail(f *testing.F) {
    // Seed corpus
    f.Add("test@example.com")
    f.Add("invalid-email")
    f.Add("")
    f.Add("a@b.c")
    f.Add("<script>@evil.com")

    f.Fuzz(func(t *testing.T, email string) {
        // Should not panic
        _, err := ValidateEmail(email)
        if err != nil {
            // Validation rejected - this is expected for invalid input
            return
        }
        // If validation passed, verify the email is actually valid
        if !isValidEmailFormat(email) {
            t.Errorf("validation passed for invalid email: %q", email)
        }
    })
}

func FuzzParseJSON(f *testing.F) {
    f.Add([]byte(`{"name":"test"}`))
    f.Add([]byte(`{}`))
    f.Add([]byte(`null`))
    f.Add([]byte(`{"name":"` + strings.Repeat("a", 10000) + `"}`))

    f.Fuzz(func(t *testing.T, data []byte) {
        var result map[string]interface{}
        // Should handle any input without panicking
        _ = json.Unmarshal(data, &result)
    })
}
```

Run fuzzing:

```bash
# Run fuzz test for 30 seconds
go test -fuzz=FuzzValidateEmail -fuzztime=30s

# Run until failure
go test -fuzz=FuzzValidateEmail

# Run with specific corpus
go test -fuzz=FuzzValidateEmail -run=^$ ./...
```

Dependency Scanning
-------------------

### Nancy (Sonatype)

Scans go.sum for known vulnerabilities:

```bash
# Install
go install github.com/sonatype-nexus-community/nancy@latest

# Scan
go list -json -deps ./... | nancy sleuth
```

### Trivy

Container and filesystem vulnerability scanner:

```bash
# Scan filesystem
trivy fs --scanners vuln .

# Scan Go module
trivy fs --scanners vuln --pkg-types library go.mod

# Output as JSON
trivy fs --format json --output results.json .
```

IDE Integration
---------------

### VS Code

Install extensions:
- Go (official)
- Go Test Explorer
- Go Nightly (for latest features)

Configure `.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--enable=gosec"],
  "go.testFlags": ["-race"],
  "go.vetOnSave": "workspace",
  "go.formatTool": "gofumpt"
}
```

### GoLand

Enable inspections:
1. Settings > Editor > Inspections > Go
2. Enable "Security" category
3. Configure gosec integration in External Tools

CI/CD Integration
-----------------

### GitHub Actions

```yaml
name: Security

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  security:
    runs-on: ubuntu-latest
    permissions:
      security-events: write  # For SARIF upload
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: gosec
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec -fmt=sarif -out=gosec.sarif ./...

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: gosec.sarif

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --enable=gosec,govet,staticcheck

      - name: Race detection
        run: go test -race -short ./...
```

### GitLab CI

```yaml
security:
  stage: test
  image: golang:1.24
  script:
    - go install golang.org/x/vuln/cmd/govulncheck@latest
    - go install github.com/securego/gosec/v2/cmd/gosec@latest
    - govulncheck ./...
    - gosec -fmt=junit-xml -out=gosec-report.xml ./...
    - go test -race ./...
  artifacts:
    reports:
      junit: gosec-report.xml
```

SBOM Generation
---------------

Generate Software Bill of Materials for compliance:

```bash
# Using syft
syft . -o spdx-json > sbom.spdx.json

# Using cyclonedx-gomod
go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
cyclonedx-gomod mod -json -output sbom.json
```

Pre-Commit Configuration
------------------------

Complete `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.55.2
    hooks:
      - id: golangci-lint
        args: ['--enable=gosec,govet,errcheck,staticcheck']

  - repo: local
    hooks:
      - id: govulncheck
        name: govulncheck
        entry: bash -c 'govulncheck ./...'
        language: system
        types: [go]
        pass_filenames: false

      - id: go-mod-tidy
        name: go mod tidy
        entry: go mod tidy
        language: system
        types: [go]
        pass_filenames: false

      - id: go-test-race
        name: go test -race
        entry: go test -race -short ./...
        language: system
        types: [go]
        pass_filenames: false

      - id: gofumpt
        name: gofumpt
        entry: gofumpt -l -w .
        language: system
        types: [go]
        pass_filenames: false
```

Makefile Integration
--------------------

```makefile
.PHONY: security lint test

# Run all security checks
security: govulncheck gosec

govulncheck:
	@echo "Running govulncheck..."
	govulncheck ./...

gosec:
	@echo "Running gosec..."
	gosec -quiet ./...

lint:
	@echo "Running golangci-lint..."
	golangci-lint run --enable=gosec,govet,staticcheck

test:
	@echo "Running tests with race detector..."
	go test -race -cover ./...

# Run before committing
precommit: lint security test
	@echo "All checks passed!"
```

Best Practices Summary
----------------------

| Tool | Purpose | When to Run |
|------|---------|-------------|
| govulncheck | Known vulnerabilities | Every commit, CI/CD |
| gosec | Static security analysis | Every commit, CI/CD |
| golangci-lint | Multi-linter runner | Every commit |
| go test -race | Race condition detection | CI/CD, before release |
| go test -fuzz | Fuzz testing | Periodically, CI/CD |
| trivy/nancy | Dependency scanning | CI/CD, scheduled |
| SBOM tools | Compliance | Release builds |

References
----------

- [govulncheck][1]
- [gosec][2]
- [golangci-lint][3]
- [Go Fuzzing][4]
- [Trivy][5]

[1]: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
[2]: https://github.com/securego/gosec
[3]: https://golangci-lint.run/
[4]: https://go.dev/doc/security/fuzz/
[5]: https://github.com/aquasecurity/trivy
