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

[1]: https://golang.org/pkg/math/rand/
[2]: https://golang.org/pkg/math/rand/#pkg-overview
[3]: https://golang.org/pkg/math/rand/#Seed
[4]: https://golang.org/pkg/crypto/rand/
