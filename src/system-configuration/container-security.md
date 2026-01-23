Container Security
==================

Containerizing Go applications requires attention to security at every layer.
This section covers secure Docker images, runtime configuration, and
Kubernetes deployment best practices.

Minimal Container Images
------------------------

Go's static binary compilation enables extremely small, secure containers:

### Multi-Stage Build (Recommended)

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-w -s -extldflags '-static'" \
    -o /app/server \
    ./cmd/server

# Final stage - minimal image
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data (if needed)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /app/server /server

# Run as non-root user
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/server"]
```

### Build Flags Explained

| Flag | Purpose |
|------|---------|
| `CGO_ENABLED=0` | Disable cgo for static binary |
| `GOOS=linux` | Target Linux |
| `GOARCH=amd64` | Target architecture |
| `-trimpath` | Remove file paths from binary |
| `-ldflags="-w -s"` | Strip debug info and symbol table |
| `-extldflags '-static'` | Force static linking |

### Image Size Comparison

| Base Image | Typical Size | Attack Surface |
|------------|--------------|----------------|
| `ubuntu:22.04` | ~77MB | High |
| `alpine:3.19` | ~7MB | Medium |
| `gcr.io/distroless/static` | ~2MB | Low |
| `scratch` | 0MB | Minimal |

Non-Root Execution
------------------

Never run containers as root:

### Using scratch (No User Database)

```dockerfile
# In scratch images, use numeric UID/GID
USER 65534:65534  # nobody:nogroup
```

### Using distroless

```dockerfile
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/server /server
USER nonroot:nonroot
ENTRYPOINT ["/server"]
```

### Using Alpine

```dockerfile
FROM alpine:3.19

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server /server

USER appuser:appgroup
ENTRYPOINT ["/server"]
```

Security Scanning
-----------------

Scan images for vulnerabilities before deployment:

### In CI/CD Pipeline

```yaml
# GitHub Actions example
- name: Build image
  run: docker build -t myapp:${{ github.sha }} .

- name: Scan with Trivy
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: myapp:${{ github.sha }}
    format: 'table'
    exit-code: '1'
    severity: 'CRITICAL,HIGH'

- name: Scan with Snyk
  uses: snyk/actions/docker@master
  with:
    image: myapp:${{ github.sha }}
    args: --severity-threshold=high
```

### Local Scanning

```bash
# Using Trivy
trivy image myapp:latest

# Using Grype
grype myapp:latest

# Using Docker Scout
docker scout cves myapp:latest
```

Kubernetes Security
-------------------

### Pod Security Standards

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: secure-app
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 65534
    runAsGroup: 65534
    fsGroup: 65534
    seccompProfile:
      type: RuntimeDefault

  containers:
    - name: app
      image: myapp:latest
      securityContext:
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
        capabilities:
          drop:
            - ALL

      resources:
        limits:
          cpu: "500m"
          memory: "256Mi"
        requests:
          cpu: "100m"
          memory: "128Mi"

      # Health checks
      livenessProbe:
        httpGet:
          path: /health
          port: 8080
        initialDelaySeconds: 5
        periodSeconds: 10

      readinessProbe:
        httpGet:
          path: /ready
          port: 8080
        initialDelaySeconds: 5
        periodSeconds: 5
```

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: app-network-policy
spec:
  podSelector:
    matchLabels:
      app: myapp
  policyTypes:
    - Ingress
    - Egress

  ingress:
    - from:
        - podSelector:
            matchLabels:
              role: frontend
      ports:
        - protocol: TCP
          port: 8080

  egress:
    - to:
        - podSelector:
            matchLabels:
              role: database
      ports:
        - protocol: TCP
          port: 5432
    # Allow DNS
    - to:
        - namespaceSelector: {}
      ports:
        - protocol: UDP
          port: 53
```

### Secrets Management

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
stringData:
  database-url: "postgres://user:pass@db:5432/app"
  api-key: "secret-key-here"

---
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: app
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: database-url
            - name: API_KEY
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: api-key
```

Health Endpoints in Go
----------------------

Implement proper health checks:

```go
package main

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "sync/atomic"
)

var (
    ready   int32 = 0
    db      *sql.DB
)

// /health - Liveness probe
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

// /ready - Readiness probe
func readyHandler(w http.ResponseWriter, r *http.Request) {
    if atomic.LoadInt32(&ready) == 0 {
        http.Error(w, "Not ready", http.StatusServiceUnavailable)
        return
    }

    // Check dependencies
    if err := db.Ping(); err != nil {
        http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Ready"))
}

func main() {
    // Initialize dependencies...
    db = initDatabase()

    // Mark as ready when initialization complete
    atomic.StoreInt32(&ready, 1)

    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("/ready", readyHandler)
    // ...
}
```

Docker Compose Security
-----------------------

For local development:

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    user: "65534:65534"
    read_only: true
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    tmpfs:
      - /tmp:noexec,nosuid,size=64m
    environment:
      - DB_HOST=db
    secrets:
      - db_password
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:16-alpine
    user: "999:999"
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password
    secrets:
      - db_password
    volumes:
      - db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

secrets:
  db_password:
    file: ./secrets/db_password.txt

volumes:
  db_data:
```

Best Practices Summary
----------------------

| Practice | Recommendation |
|----------|----------------|
| Base image | Use `scratch` or `distroless` |
| User | Run as non-root (UID 65534) |
| Filesystem | Read-only root filesystem |
| Capabilities | Drop ALL, add only what's needed |
| Network | Use network policies to restrict traffic |
| Secrets | Use Kubernetes secrets or secret managers |
| Scanning | Scan images in CI/CD before deployment |
| Resources | Set CPU/memory limits |
| Health checks | Implement liveness and readiness probes |

References
----------

- [Docker Security Best Practices][1]
- [Kubernetes Pod Security Standards][2]
- [Google Distroless Images][3]
- [Trivy Scanner][4]

[1]: https://docs.docker.com/develop/security-best-practices/
[2]: https://kubernetes.io/docs/concepts/security/pod-security-standards/
[3]: https://github.com/GoogleContainerTools/distroless
[4]: https://github.com/aquasecurity/trivy
