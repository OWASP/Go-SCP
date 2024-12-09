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

[1]: https://pkg.go.dev/math/rand/v2
[2]: https://pkg.go.dev/math/rand/v2#pkg-overview
[3]: https://go.dev/blog/chacha8rand
[4]: https://golang.org/pkg/crypto/rand/
[5]: https://pkg.go.dev/math/rand
