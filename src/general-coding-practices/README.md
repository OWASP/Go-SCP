General Coding Practices
========================

This section covers secure coding practices that apply broadly across web
application development. These practices help prevent common vulnerabilities
and improve the overall security posture of your Go applications.

Topics Covered
--------------

- **[Cross-Site Request Forgery (CSRF)](cross-site-request-forgery.md)** -
  Protecting state-changing requests from being forged by attackers.

- **[Regular Expressions](regular-expressions.md)** - Safe use of regex to
  avoid ReDoS (Regular Expression Denial of Service) attacks.

Key Principles
--------------

1. **Defense in Depth**: Never rely on a single security control. Layer
   multiple protections.

2. **Fail Securely**: When errors occur, default to a secure state. Deny
   access rather than allow it.

3. **Least Privilege**: Grant only the minimum permissions necessary for
   functionality.

4. **Input Validation**: Validate all input from untrusted sources. See
   [Input Validation](../input-validation/README.md).

5. **Output Encoding**: Encode output appropriately for the context. See
   [Output Encoding](../output-encoding/README.md).

Security Headers
----------------

Always set security headers on HTTP responses. Here's a recommended middleware:

```go
func securityMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent MIME type sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")

        // Clickjacking protection
        w.Header().Set("X-Frame-Options", "DENY")

        // Referrer policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        // Content Security Policy (customize for your needs)
        // See Output Encoding > Content Security Policy for details
        w.Header().Set("Content-Security-Policy", "default-src 'self'")

        // HTTPS enforcement
        w.Header().Set("Strict-Transport-Security",
            "max-age=31536000; includeSubDomains")

        next.ServeHTTP(w, r)
    })
}
```

Related Sections
----------------

- [Security Tooling](../security-tooling/README.md) - Tools for finding
  vulnerabilities in your code
- [Error Handling](../error-handling-logging/README.md) - Secure error handling
  practices
- [Communication Security](../communication-security/README.md) - TLS and
  secure transport
