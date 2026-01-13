Pseudo-Random Generators
========================

In OWASP Secure Coding Practices you'll find what seems to be a really complex
guideline: "_All random numbers, random file names, random GUIDs, and random
strings should be generated using the cryptographic module’s approved random
number generator when these random values are intended to be un-guessable_", so
let's discuss "random numbers".

Cryptography relies on some randomness, but for the sake of correctness, what
most programming languages provide out-of-the-box is a pseudo-random number
generator: for example, [Go's math/rand][1] is not an exception.

You should carefully read the documentation when it states that "_Top-level
functions, such as Float64 and Int, use a default shared Source that produces a
**deterministic sequence** of values each time a program is run._" ([source][2])

What exactly does that mean? Let's see:

```go
package main

import "fmt"
import "math/rand"

func main() {
    fmt.Println("Random Number: ", rand.Intn(1984))
}
```

Running this program several times will lead exactly to the same
number/sequence, but why?

```bash
$ for i in {1..5}; do go run rand.go; done
Random Number:  1825
Random Number:  1825
Random Number:  1825
Random Number:  1825
Random Number:  1825
```

Because [Go's math/rand][1] is a deterministic pseudo-random number generator.
Similar to many others, it uses a source, called a Seed. This Seed is **solely**
responsible for the randomness of the deterministic pseudo-random number
generator. If it is known or predictable, the same will happen to generated
number sequence.

We could "fix" this example quite easily by using the
[math/rand Seed function][3], getting the expected five different values for
each program execution. But because we're on Cryptographic Practices section, we
should follow to [Go's crypto/rand package][4].

```go
package main

import "fmt"
import "math/big"
import "crypto/rand"

func main() {
    rand, err := rand.Int(rand.Reader, big.NewInt(1984))
    if err != nil {
        panic(err)
    }

    fmt.Printf("Random Number: %d\n", rand)
}
```

You may notice that running [crypto/rand][4] is slower than [math/rand][1], but
this is expected since the fastest algorithm isn't always the safest. Crypto's
rand is also safer to implement. An example of this is the fact that you
_CANNOT_ seed crypto/rand, since the library uses OS-randomness for this,
preventing developer misuse.

```bash
$ for i in {1..5}; do go run rand-safe.go; done
Random Number: 277
Random Number: 1572
Random Number: 1793
Random Number: 1328
Random Number: 1378
```

If you're curious about how this can be exploited just think what happens if
your application creates a default password on user signup, by computing the
hash of a pseudo-random number generated with [Go's math/rand][1], as shown in
the first example.

Yes, you guessed it, you would be able to predict the user's password!

Go 1.22: ChaCha8Rand Improvements
---------------------------------

Go 1.22 introduced a significant security improvement: the `math/rand` package
now uses **ChaCha8Rand**, a cryptographically-strong pseudorandom number
generator based on the ChaCha stream cipher.

### What Changed

| Aspect | Pre-Go 1.22 | Go 1.22+ |
|--------|-------------|----------|
| Algorithm | Linear feedback shift register | ChaCha8 stream cipher |
| Seed size | 63 bits | 256 bits |
| Predictability | ~607 values to predict all future outputs | Cryptographically secure |
| Default behavior | Deterministic with same seed | Random seed from crypto/rand |

### Security Implications

**Before Go 1.22:**
Using `math/rand` for security-sensitive operations was a "serious security
catastrophe" - an attacker observing output could predict all future values.

**Go 1.22+:**
Accidental use of `math/rand` instead of `crypto/rand` is now "just a mistake"
rather than a critical vulnerability. However, `crypto/rand` is **still
recommended** for security-sensitive operations.

### Recommendations

```go
// ALWAYS use for security-sensitive operations
import "crypto/rand"

func generateToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}

// OK for non-security uses (shuffling, sampling, etc.)
import "math/rand/v2"

func shuffleItems(items []Item) {
    rand.Shuffle(len(items), func(i, j int) {
        items[i], items[j] = items[j], items[i]
    })
}
```

### math/rand vs math/rand/v2

Go 1.22 introduced `math/rand/v2`:

```go
// Old package (still works, uses ChaCha8 by default now)
import "math/rand"

// New package (recommended for new code)
import "math/rand/v2"
```

The v2 package has cleaner APIs and always uses ChaCha8Rand.

FIPS 140-3 Compliance (Go 1.24+)
--------------------------------

For applications in regulated environments (government, finance, healthcare),
Go 1.24 introduced native FIPS 140-3 support.

### Enabling FIPS Mode

```bash
# Build with FIPS-validated cryptography
GOFIPS140=latest go build ./...

# Or use specific validated version
GOFIPS140=v1.0.0 go build ./...
```

### Runtime Verification

```go
import "crypto/fips140"

func main() {
    if fips140.Enabled() {
        log.Println("Running in FIPS 140-3 compliant mode")
    }
}
```

### What FIPS Mode Changes

- Uses FIPS-validated implementations of cryptographic algorithms
- Restricts to approved algorithms (AES, SHA-2, ECDSA, RSA, etc.)
- Disables non-approved algorithms
- No cgo required - pure Go implementation

### When to Use FIPS Mode

- Government contracts requiring FIPS compliance
- Financial services with regulatory requirements
- Healthcare applications handling PHI
- Any system requiring CMVP validation

Summary: When to Use Each Package
---------------------------------

| Use Case | Package |
|----------|---------|
| Tokens, keys, secrets, passwords | `crypto/rand` |
| Session IDs, CSRF tokens | `crypto/rand` |
| Nonces, IVs for encryption | `crypto/rand` |
| Shuffling, sampling, simulations | `math/rand/v2` |
| Random delays, jitter | `math/rand/v2` |
| Test data generation | `math/rand/v2` |

[1]: https://pkg.go.dev/math/rand
[2]: https://pkg.go.dev/math/rand#pkg-overview
[3]: https://pkg.go.dev/math/rand#Seed
[4]: https://pkg.go.dev/crypto/rand
[5]: https://go.dev/blog/chacha8rand
[6]: https://go.dev/doc/security/fips140
