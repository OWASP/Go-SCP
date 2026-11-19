Content Security Policy
=======================

Content Security Policy (CSP) is a critical HTTP response header that helps
prevent Cross-Site Scripting (XSS), clickjacking, and other code injection
attacks. It allows you to specify which resources the browser should load.

Why CSP Matters
---------------

Even with proper output encoding (see [XSS Prevention](cross-site-scripting.md)),
CSP provides defense-in-depth:

- **Blocks inline scripts** by default, eliminating most XSS vectors
- **Controls resource origins** - prevents loading malicious external scripts
- **Reports violations** - helps identify attacks and misconfigurations
- **Mitigates data exfiltration** - controls where data can be sent

Basic CSP Implementation
------------------------

### Setting the Header

```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Basic CSP - customize based on your needs
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; "+
            "script-src 'self'; "+
            "style-src 'self'; "+
            "img-src 'self' data:; "+
            "font-src 'self'; "+
            "connect-src 'self'; "+
            "frame-ancestors 'none'; "+
            "form-action 'self'")

        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)

    // Wrap with security headers middleware
    http.ListenAndServe(":8080", securityHeaders(mux))
}
```

CSP Directives
--------------

| Directive | Purpose | Example |
|-----------|---------|---------|
| `default-src` | Fallback for other directives | `'self'` |
| `script-src` | Valid sources for JavaScript | `'self' 'nonce-xxx'` |
| `style-src` | Valid sources for CSS | `'self' 'unsafe-inline'` |
| `img-src` | Valid sources for images | `'self' data: https:` |
| `connect-src` | Valid targets for fetch/XHR | `'self' api.example.com` |
| `font-src` | Valid sources for fonts | `'self' fonts.googleapis.com` |
| `frame-ancestors` | Who can embed this page | `'none'` |
| `form-action` | Valid targets for forms | `'self'` |
| `base-uri` | Valid targets for `<base>` | `'self'` |
| `report-uri` | Where to send violation reports | `/csp-report` |

Source Values
-------------

| Value | Meaning |
|-------|---------|
| `'self'` | Same origin only |
| `'none'` | Block all |
| `'unsafe-inline'` | Allow inline scripts/styles (avoid!) |
| `'unsafe-eval'` | Allow eval() (avoid!) |
| `'nonce-xxx'` | Allow scripts with matching nonce |
| `'strict-dynamic'` | Trust scripts loaded by trusted scripts |
| `https:` | Any HTTPS origin |
| `data:` | Data URIs |
| `example.com` | Specific domain |

Nonce-Based CSP (Recommended)
-----------------------------

Nonces allow specific inline scripts while blocking injected ones:

```go
package main

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "html/template"
    "net/http"
)

// Generate a cryptographically secure nonce
func generateNonce() (string, error) {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.StdEncoding.EncodeToString(b), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
    // Generate unique nonce for this request
    nonce, err := generateNonce()
    if err != nil {
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }

    // Set CSP header with nonce
    csp := fmt.Sprintf(
        "default-src 'self'; "+
        "script-src 'self' 'nonce-%s' 'strict-dynamic'; "+
        "style-src 'self' 'nonce-%s'; "+
        "base-uri 'self'; "+
        "frame-ancestors 'none'",
        nonce, nonce)
    w.Header().Set("Content-Security-Policy", csp)

    // Pass nonce to template
    tmpl := template.Must(template.ParseFiles("page.html"))
    tmpl.Execute(w, map[string]string{"Nonce": nonce})
}
```

Template using the nonce:

```html
<!DOCTYPE html>
<html>
<head>
    <!-- Inline script allowed because it has the correct nonce -->
    <script nonce="{{.Nonce}}">
        console.log("This script is allowed");
    </script>

    <!-- Inline style allowed with nonce -->
    <style nonce="{{.Nonce}}">
        body { font-family: sans-serif; }
    </style>
</head>
<body>
    <h1>Secure Page</h1>

    <!-- This injected script would be blocked (no nonce) -->
    <!-- <script>alert('XSS')</script> -->
</body>
</html>
```

Report-Only Mode
----------------

Test your CSP without breaking functionality:

```go
// Use Content-Security-Policy-Report-Only for testing
w.Header().Set("Content-Security-Policy-Report-Only",
    "default-src 'self'; "+
    "report-uri /csp-report")
```

### Handling Reports

```go
type CSPReport struct {
    CSPReport struct {
        DocumentURI        string `json:"document-uri"`
        Referrer           string `json:"referrer"`
        ViolatedDirective  string `json:"violated-directive"`
        EffectiveDirective string `json:"effective-directive"`
        OriginalPolicy     string `json:"original-policy"`
        BlockedURI         string `json:"blocked-uri"`
        StatusCode         int    `json:"status-code"`
    } `json:"csp-report"`
}

func cspReportHandler(w http.ResponseWriter, r *http.Request) {
    var report CSPReport
    if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Log the violation for analysis
    log.Printf("CSP Violation: %s blocked by %s on %s",
        report.CSPReport.BlockedURI,
        report.CSPReport.ViolatedDirective,
        report.CSPReport.DocumentURI)

    w.WriteHeader(http.StatusNoContent)
}
```

Strict CSP Configuration
------------------------

For maximum security, use this strict configuration:

```go
func strictCSP(nonce string) string {
    return fmt.Sprintf(
        "default-src 'none'; "+                    // Block everything by default
        "script-src 'nonce-%s' 'strict-dynamic'; "+ // Nonce-based scripts only
        "style-src 'nonce-%s'; "+                  // Nonce-based styles only
        "img-src 'self'; "+                        // Images from same origin
        "font-src 'self'; "+                       // Fonts from same origin
        "connect-src 'self'; "+                    // XHR/fetch to same origin
        "form-action 'self'; "+                    // Forms submit to same origin
        "base-uri 'none'; "+                       // No <base> element
        "frame-ancestors 'none'; "+                // Cannot be framed
        "upgrade-insecure-requests",               // Upgrade HTTP to HTTPS
        nonce, nonce)
}
```

Common Pitfalls
---------------

1. **Using 'unsafe-inline'**: Defeats the purpose of CSP. Use nonces instead.

2. **Overly permissive policies**: `script-src *` provides no protection.

3. **Forgetting frame-ancestors**: Leaves you vulnerable to clickjacking.

4. **Not using report-uri**: You won't know when attacks are blocked.

5. **Missing base-uri**: Allows attackers to hijack relative URLs.

Additional Security Headers
---------------------------

CSP works best with other security headers:

```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // CSP (customize as shown above)
        w.Header().Set("Content-Security-Policy", "...")

        // Prevent MIME sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")

        // Clickjacking protection (older browsers)
        w.Header().Set("X-Frame-Options", "DENY")

        // XSS filter (legacy browsers)
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        // Referrer policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        // Permissions policy (control browser features)
        w.Header().Set("Permissions-Policy",
            "geolocation=(), microphone=(), camera=()")

        next.ServeHTTP(w, r)
    })
}
```

References
----------

- [MDN: Content-Security-Policy][1]
- [OWASP CSP Cheat Sheet][2]
- [CSP Evaluator Tool][3]

[1]: https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
[2]: https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html
[3]: https://csp-evaluator.withgoogle.com/
