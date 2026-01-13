Claude Code Security Development
================================

This section provides guidance for developers using Claude Code (or similar AI
assistants) to write secure Go applications. It includes project configuration,
security-focused prompts, and automated security review workflows.

CLAUDE.md Security Template
---------------------------

Create a `CLAUDE.md` file in your Go project root to provide security context
to Claude Code. This file is automatically read and influences code generation.

```markdown
# CLAUDE.md

## Project Security Requirements

This Go project follows OWASP Secure Coding Practices. All code must:

### Authentication & Authorization
- Use Argon2id for password hashing (see auth/password.go)
- Implement JWT with golang-jwt/jwt/v5
- Validate all tokens server-side

### Input Validation
- Validate ALL user input using go-playground/validator
- Never trust client-side validation alone
- Sanitize output for context (HTML, SQL, shell)

### Database Security
- Use parameterized queries ONLY - never string concatenation
- Validate ORDER BY columns against allowlist
- Use context with timeouts for all queries

### Error Handling
- Never expose internal errors to users
- Log errors with slog including correlation IDs
- Return generic error messages to clients

### Cryptography
- Use crypto/rand for all security-sensitive randomness
- TLS 1.3 minimum for all external connections
- No hardcoded secrets - use environment variables

### HTTP Security
- Set ReadHeaderTimeout on all servers (Slowloris protection)
- Implement rate limiting for public endpoints
- Configure CORS restrictively

## Security Tools
- Run `govulncheck ./...` before committing
- Run `gosec ./...` for static analysis
- All security findings must be addressed

## Prohibited Patterns
- NO hardcoded credentials or API keys
- NO string concatenation in SQL queries
- NO shell command execution with user input
- NO disabled TLS certificate verification
- NO use of math/rand for security purposes
```

OWASP Go Security Subagent
--------------------------

Create a custom Claude Code command for security reviews. Save this as
`.claude/commands/security-review.md` in your project:

```markdown
# Security Review

Perform a comprehensive OWASP-focused security review of the Go code.

## Review Checklist

### 1. Input Validation (A03:2021)
- [ ] All user inputs validated with go-playground/validator
- [ ] Struct tags define validation rules
- [ ] Validation errors don't leak internal details

### 2. Injection Prevention (A03:2021)
- [ ] SQL uses parameterized queries only
- [ ] No string concatenation with user input in queries
- [ ] Command execution uses exec.Command without shell
- [ ] Template rendering uses html/template not text/template

### 3. Authentication (A07:2021)
- [ ] Passwords hashed with Argon2id
- [ ] JWT using golang-jwt/jwt/v5 (not dgrijalva)
- [ ] Token validation checks all claims
- [ ] Session cookies have Secure, HttpOnly, SameSite

### 4. Cryptography (A02:2021)
- [ ] crypto/rand used for security randomness
- [ ] No weak algorithms (MD5, SHA1 for security)
- [ ] TLS 1.3 minimum configured
- [ ] Secrets loaded from environment/vault

### 5. Error Handling
- [ ] Internal errors logged, not returned to users
- [ ] Panic recovery in HTTP handlers
- [ ] Context timeouts on all external calls

### 6. Access Control (A01:2021)
- [ ] Path traversal prevented (filepath.Clean + validation)
- [ ] Rate limiting on sensitive endpoints
- [ ] CORS configured restrictively

## How to Use

Run this command: `/project:security-review`

Claude will analyze the codebase against this checklist and report findings.
```

To use this subagent, run `/project:security-review` in Claude Code.

Security Hooks Configuration
----------------------------

Configure Claude Code hooks to enforce security checks. Create
`.claude/settings.json`:

```json
{
  "hooks": {
    "preCommit": [
      {
        "command": "govulncheck ./...",
        "description": "Check for known vulnerabilities",
        "blocking": true
      },
      {
        "command": "gosec -quiet ./...",
        "description": "Static security analysis",
        "blocking": true
      }
    ],
    "postFileWrite": [
      {
        "pattern": "*.go",
        "command": "gofmt -w $FILE",
        "description": "Format Go files"
      }
    ]
  }
}
```

These hooks ensure security tools run automatically during development.

Pre-Commit Hook Integration
---------------------------

For team-wide enforcement, use pre-commit hooks. Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.55.2
    hooks:
      - id: golangci-lint
        args: ['--enable=gosec']

  - repo: local
    hooks:
      - id: govulncheck
        name: govulncheck
        entry: govulncheck ./...
        language: system
        pass_filenames: false
        types: [go]

      - id: go-test
        name: go test
        entry: go test -race ./...
        language: system
        pass_filenames: false
        types: [go]
```

Install with:

```bash
pip install pre-commit
pre-commit install
```

MCP Security Integrations
-------------------------

Claude Code supports Model Context Protocol (MCP) servers for extended
capabilities. Security-relevant MCP integrations include:

### Snyk MCP Server

```json
{
  "mcpServers": {
    "snyk": {
      "command": "npx",
      "args": ["@anthropic/snyk-mcp-server"],
      "env": {
        "SNYK_TOKEN": "${SNYK_TOKEN}"
      }
    }
  }
}
```

This enables Claude to query Snyk's vulnerability database during code review.

### Custom Security MCP Server

For enterprise environments, create a custom MCP server that enforces
organization-specific security policies:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
)

type SecurityCheck struct {
    Name   string `json:"name"`
    Status string `json:"status"`
    Detail string `json:"detail,omitempty"`
}

func runGovulncheck() SecurityCheck {
    cmd := exec.Command("govulncheck", "./...")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return SecurityCheck{
            Name:   "govulncheck",
            Status: "FAIL",
            Detail: string(output),
        }
    }
    return SecurityCheck{
        Name:   "govulncheck",
        Status: "PASS",
    }
}

func runGosec() SecurityCheck {
    cmd := exec.Command("gosec", "-fmt=json", "-quiet", "./...")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return SecurityCheck{
            Name:   "gosec",
            Status: "FAIL",
            Detail: string(output),
        }
    }
    return SecurityCheck{
        Name:   "gosec",
        Status: "PASS",
    }
}

func main() {
    checks := []SecurityCheck{
        runGovulncheck(),
        runGosec(),
    }

    result, _ := json.MarshalIndent(checks, "", "  ")
    fmt.Println(string(result))

    // Exit with error if any check failed
    for _, check := range checks {
        if check.Status == "FAIL" {
            os.Exit(1)
        }
    }
}
```

Security-Focused Prompting
--------------------------

When using Claude Code for Go development, include security context in prompts:

### Good Prompts

```
"Add user registration with Argon2id password hashing and input validation"

"Create a database query function using parameterized queries for PostgreSQL"

"Implement JWT authentication using golang-jwt/jwt/v5 with proper claim validation"

"Add rate limiting to the /api/login endpoint using golang.org/x/time/rate"
```

### Prompts to Avoid

```
"Add user registration" (no security context)

"Query the database for this user" (might generate string concatenation)

"Add authentication" (too vague, might use insecure patterns)
```

Automated Security Review Workflow
----------------------------------

Create a comprehensive security review workflow:

### 1. Initial Code Generation

```
"Create a REST API endpoint for user creation following OWASP guidelines:
- Validate input with go-playground/validator
- Hash passwords with Argon2id
- Use parameterized queries
- Return generic errors to clients
- Log detailed errors server-side"
```

### 2. Security Review Request

```
"Review this code for OWASP Top 10 vulnerabilities:
- A01: Broken Access Control
- A02: Cryptographic Failures
- A03: Injection
- A07: Identification and Authentication Failures"
```

### 3. Fix Implementation

```
"Fix the identified security issues while maintaining functionality.
Explain each fix and the vulnerability it addresses."
```

CI/CD Security Gate
-------------------

Integrate security checks into CI/CD pipelines:

```yaml
# .github/workflows/security.yml
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
          go-version: '1.24'

      - name: Install security tools
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          go install github.com/securego/gosec/v2/cmd/gosec@latest

      - name: Run govulncheck
        run: govulncheck ./...

      - name: Run gosec
        run: gosec -fmt=sarif -out=gosec.sarif ./...

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif

      - name: Run tests with race detector
        run: go test -race ./...
```

Best Practices Summary
----------------------

| Practice | Recommendation |
|----------|----------------|
| CLAUDE.md | Always include security requirements |
| Custom commands | Create project-specific security review |
| Hooks | Automate govulncheck and gosec |
| Prompts | Include security context explicitly |
| Review | Request OWASP-focused code review |
| CI/CD | Gate deployments on security checks |

References
----------

- [Claude Code Documentation][1]
- [OWASP Secure Coding Practices][2]
- [govulncheck][3]
- [gosec][4]

[1]: https://docs.anthropic.com/en/docs/claude-code
[2]: https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/
[3]: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
[4]: https://github.com/securego/gosec
