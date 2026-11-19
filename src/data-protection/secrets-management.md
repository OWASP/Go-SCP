Secrets Management
==================

Proper secrets management is critical for application security. This section
covers how to handle API keys, database credentials, encryption keys, and
other sensitive configuration in Go applications.

The Golden Rule
---------------

**Never hardcode secrets in source code.**

```go
// NEVER DO THIS
const apiKey = "sk-live-abc123def456"
const dbPassword = "supersecret123"
```

Hardcoded secrets:
- Get committed to version control
- Appear in build artifacts and logs
- Cannot be rotated without code changes
- May be visible to anyone with code access

Environment Variables (Basic)
-----------------------------

The minimum acceptable approach for secrets:

```go
package main

import (
    "log"
    "os"
)

func main() {
    // Read from environment
    apiKey := os.Getenv("API_KEY")
    if apiKey == "" {
        log.Fatal("API_KEY environment variable required")
    }

    dbPassword := os.Getenv("DB_PASSWORD")
    if dbPassword == "" {
        log.Fatal("DB_PASSWORD environment variable required")
    }

    // Use the secrets...
}
```

### Required vs Optional Secrets

```go
// MustGetEnv panics if the environment variable is not set
func MustGetEnv(key string) string {
    value := os.Getenv(key)
    if value == "" {
        log.Fatalf("Required environment variable %s is not set", key)
    }
    return value
}

// GetEnvOrDefault returns a default value if not set
func GetEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func main() {
    // Required - will fail if not set
    apiKey := MustGetEnv("API_KEY")

    // Optional with safe default
    logLevel := GetEnvOrDefault("LOG_LEVEL", "info")
}
```

Configuration Files with Viper
------------------------------

For complex configuration, use [Viper](https://github.com/spf13/viper):

```bash
go get github.com/spf13/viper
```

```go
package main

import (
    "log"
    "github.com/spf13/viper"
)

type Config struct {
    Database struct {
        Host     string `mapstructure:"host"`
        Port     int    `mapstructure:"port"`
        User     string `mapstructure:"user"`
        Password string `mapstructure:"password"` // From env var
        Name     string `mapstructure:"name"`
    } `mapstructure:"database"`

    API struct {
        Key    string `mapstructure:"key"` // From env var
        Secret string `mapstructure:"secret"` // From env var
    } `mapstructure:"api"`
}

func LoadConfig() (*Config, error) {
    // Set config file
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("/etc/myapp/")

    // Allow environment variables to override
    viper.AutomaticEnv()

    // Map environment variables to config keys
    viper.BindEnv("database.password", "DB_PASSWORD")
    viper.BindEnv("api.key", "API_KEY")
    viper.BindEnv("api.secret", "API_SECRET")

    // Set safe defaults (not real secrets!)
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("database.port", 5432)

    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
        // Config file not found - use env vars only
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    // Validate required secrets
    if config.Database.Password == "" {
        log.Fatal("DB_PASSWORD environment variable required")
    }
    if config.API.Key == "" {
        log.Fatal("API_KEY environment variable required")
    }

    return &config, nil
}
```

Config file (`config.yaml`) - **no secrets**:

```yaml
database:
  host: localhost
  port: 5432
  user: myapp
  name: myapp_production
  # password: loaded from DB_PASSWORD env var

api:
  # key: loaded from API_KEY env var
  # secret: loaded from API_SECRET env var
```

Secret Management Services
--------------------------

For production, use dedicated secret management:

### HashiCorp Vault

```bash
go get github.com/hashicorp/vault/api
```

```go
package main

import (
    "context"
    "log"

    vault "github.com/hashicorp/vault/api"
)

func getSecretFromVault(path, key string) (string, error) {
    config := vault.DefaultConfig()
    client, err := vault.NewClient(config)
    if err != nil {
        return "", err
    }

    // Token from environment or other auth method
    // In production, use AppRole, Kubernetes auth, etc.

    secret, err := client.KVv2("secret").Get(context.Background(), path)
    if err != nil {
        return "", err
    }

    value, ok := secret.Data[key].(string)
    if !ok {
        return "", fmt.Errorf("key %s not found", key)
    }

    return value, nil
}

func main() {
    dbPassword, err := getSecretFromVault("myapp/database", "password")
    if err != nil {
        log.Fatalf("Failed to get database password: %v", err)
    }

    // Use dbPassword...
}
```

### AWS Secrets Manager

```bash
go get github.com/aws/aws-sdk-go-v2/service/secretsmanager
```

```go
package main

import (
    "context"
    "encoding/json"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func getAWSSecret(secretName string) (map[string]string, error) {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        return nil, err
    }

    client := secretsmanager.NewFromConfig(cfg)

    result, err := client.GetSecretValue(context.Background(),
        &secretsmanager.GetSecretValueInput{
            SecretId: &secretName,
        })
    if err != nil {
        return nil, err
    }

    var secrets map[string]string
    if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
        return nil, err
    }

    return secrets, nil
}
```

### Kubernetes Secrets

In Kubernetes, mount secrets as files or environment variables:

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
        - name: myapp
          env:
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: myapp-secrets
                  key: db-password
          volumeMounts:
            - name: secrets
              mountPath: /etc/secrets
              readOnly: true
      volumes:
        - name: secrets
          secret:
            secretName: myapp-secrets
```

Secure Logging
--------------

Never log secrets:

```go
// INSECURE - Logging secrets
log.Printf("Connecting to database with password: %s", password) // NEVER!

// SECURE - Log without secrets
log.Printf("Connecting to database at %s:%d as %s", host, port, user)

// SECURE - Redact sensitive fields
type Config struct {
    Host     string
    Password string
}

func (c Config) String() string {
    return fmt.Sprintf("Config{Host: %s, Password: [REDACTED]}", c.Host)
}
```

### Using slog with Redaction

```go
import "log/slog"

type RedactedString string

func (r RedactedString) LogValue() slog.Value {
    return slog.StringValue("[REDACTED]")
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password RedactedString // Will show [REDACTED] in logs
}

func main() {
    config := DatabaseConfig{
        Host:     "localhost",
        Port:     5432,
        User:     "admin",
        Password: "secret123",
    }

    // Safe to log - password is redacted
    slog.Info("database config loaded", slog.Any("config", config))
    // Output: database config loaded config={Host:localhost Port:5432 User:admin Password:[REDACTED]}
}
```

Secret Rotation
---------------

Design for secret rotation from the start:

```go
type SecretProvider interface {
    GetSecret(ctx context.Context, key string) (string, error)
}

type DatabaseConnection struct {
    secretProvider SecretProvider
    secretKey      string
    currentSecret  string
    mu             sync.RWMutex
}

func (dc *DatabaseConnection) refreshSecret(ctx context.Context) error {
    newSecret, err := dc.secretProvider.GetSecret(ctx, dc.secretKey)
    if err != nil {
        return err
    }

    dc.mu.Lock()
    dc.currentSecret = newSecret
    dc.mu.Unlock()

    // Reconnect with new credentials
    return dc.reconnect()
}

// Run periodic refresh
func (dc *DatabaseConnection) StartSecretRefresh(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    go func() {
        for {
            select {
            case <-ctx.Done():
                ticker.Stop()
                return
            case <-ticker.C:
                if err := dc.refreshSecret(ctx); err != nil {
                    slog.Error("failed to refresh secret", "error", err)
                }
            }
        }
    }()
}
```

Best Practices Summary
----------------------

| Do | Don't |
|----|-------|
| Use environment variables or secret managers | Hardcode secrets in code |
| Validate required secrets at startup | Fail silently if secrets missing |
| Log secret names, not values | Log secret values or connection strings |
| Implement secret rotation | Assume secrets never need changing |
| Use separate secrets per environment | Share secrets across environments |
| Encrypt secrets at rest | Store plaintext secrets in config files |
| Audit secret access | Allow unrestricted secret access |

References
----------

- [12-Factor App: Config][1]
- [HashiCorp Vault][2]
- [AWS Secrets Manager][3]
- [Kubernetes Secrets][4]

[1]: https://12factor.net/config
[2]: https://www.vaultproject.io/
[3]: https://aws.amazon.com/secrets-manager/
[4]: https://kubernetes.io/docs/concepts/configuration/secret/
