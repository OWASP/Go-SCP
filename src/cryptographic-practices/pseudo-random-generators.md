Pseudo-Random Generators
========================

In OWASP Secure Coding Practices you'll find what seems to be a really complex
guideline: "_All random numbers, random file names, random GUIDs, and random
strings should be generated using the cryptographic module’s approved random
number generator when these random values are intended to be un-guessable_", so
let's discuss "random numbers".

Cryptography relies on some randomness, but for the sake of correctness, what
most programming languages provide out-of-the-box is a pseudo-random number
generator: for example, [Go's math/rand/v2][1] is not an exception. Note that 
the older `math/rand` is not either a cryptographically secure number generator.

You should carefully read the documentation when it states that "_This package's 
outputs might be easily predictable regardless of how it's seeded_" ([source][2]). 
The `math/rand/v2` package's random values are not as random as values provided 
by `crypto/rand` ([source][3]).

But because we're on Cryptographic Practices section, we
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

You may notice that running [crypto/rand][4] is slower than [math/rand/v2][1], but
this is expected since the fastest algorithm isn't always the safest. Crypto's
rand is also safer to implement. An example of this is the fact that you
_CANNOT_ seed `crypto/rand`, since the library uses OS-randomness for this,
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
hash of a pseudo-random number generated with the [Go's math/rand][5] package's
now deprecated Seed function.

Yes, you guessed it, you would be able to predict the user's password!

Go 1.22: ChaCha8Rand Improvements
---------------------------------

Go 1.22 introduced a significant security improvement: the top-level functions
of `math/rand` and `math/rand/v2` (such as `rand.Int` and `rand.Shuffle`) now
use **ChaCha8Rand**, a cryptographically-strong pseudorandom number generator
based on the ChaCha stream cipher ([source][3]). A generator you build yourself
with `rand.New(rand.NewSource(seed))` in `math/rand` still uses the old
algorithm and is fully predictable from its seed.

### What Changed

| Aspect | Pre-Go 1.22 | Go 1.22+ |
|--------|-------------|----------|
| Algorithm | Additive lagged Fibonacci generator | ChaCha8 stream cipher |
| Seed size | 63 bits | 256 bits |
| Predictability | ~607 observed values predict all future outputs | Cryptographically strong |
| Default seed | Fixed seed up to Go 1.19; random seed since Go 1.20 | Random seed from the operating system |

### Security Implications

**Before Go 1.22:**
Using `math/rand` for security-sensitive operations was a serious
vulnerability: an attacker observing enough output could predict all future
values.

**Go 1.22+:**
Accidental use of the top-level `math/rand` functions instead of `crypto/rand`
is much less dangerous than it used to be. However, `crypto/rand` is **still
required** for security-sensitive operations: the `math/rand` packages make no
security promises, and explicitly seeded generators remain predictable.

Since Go 1.20 the top-level `Seed` function is deprecated, and since Go 1.24
calling it has no effect (unless `GODEBUG=randseednop=0` is set).

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
// Old package (still works; its top-level functions use ChaCha8 since Go 1.22)
import "math/rand"

// New package (recommended for new code)
import "math/rand/v2"
```

The v2 package has cleaner APIs, has no top-level `Seed` function, and its
top-level functions always use ChaCha8Rand. It also offers explicit `ChaCha8`
and `PCG` sources for when you need a reproducible sequence.

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

[1]: https://pkg.go.dev/math/rand/v2
[2]: https://pkg.go.dev/math/rand/v2#pkg-overview
[3]: https://go.dev/blog/chacha8rand
[4]: https://pkg.go.dev/crypto/rand
[5]: https://pkg.go.dev/math/rand
[6]: https://go.dev/doc/security/fips140
