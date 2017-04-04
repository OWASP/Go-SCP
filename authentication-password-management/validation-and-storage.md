Validation and Storing authentication data
==========================================

The key subject of this section is the authentication data storage, as more
often than desirable, user account databases are leaked on the internet.
Of course that this is not guaranteed to happen, but in the case of such
an event, collateral damages can be avoided if authentication data,
especially passwords, are stored properly.

First, let's make it clear that "_all authentication controls should fail
securely_". You're recommended to read all other Authentication and Password
Management sections as they cover recommendations about
reporting back wrong authentication data and how to handle logging.

One other preliminary recommendation: for sequential authentication
implementations (like Google does nowadays), validation should happen only on
the completion of all data input, on a trusted system (e.g. the server).

Now let's talk about storing passwords.

You don't really need to store passwords as they are provided by the users
(plaintext) but you'll need to validate on each authentication whether users
are providing the same token.

So, for security reasons, what you need is a "one way" function `H` so that for
every password `p1` and `p2`, `p1` is different from `p2`, `H(p1)` is also
different from `H(p2)`[^1].

Does this sound, or look, like Math?
Pay attention to this last requirement: `H` should be such a function that
there's no function `H⁻¹` so that `H⁻¹(H(p1))` is equal to `p1`. This means
that there's no way back to the original `p1`.

If `H` is one-way only, what's the real problem about account leakage?

Well, if you know all possible passwords, you can pre-compute their hashes and
then run a rainbow table attack.

Certainly you were already told that passwords are hard to manage from user's
point of view, and that users are not only able re-use passwords but they also
tend to use something easy to remember, which makes the universe really small.

How can we avoid this?

The point is: if two different users provide the same password `p1` we should
store a different hashed value.
It may sound impossible but the answer is `salt`: a pseudo-random value which is
append to `p1` so that the resulting hash is computed as follow: `H(p1 + salt)`.

So each entry on passwords store should keep the resulting hash and the `salt`
itself in plaintext: `salt` does not offer any security concerns.

Last recommendations.
* Avoid using MD5 hashing algorithm whenever possible;
* Avoid using SHA1 as it has been cracked recently.
* Read the [Pseudo-Random Generators section][1].

```go
package main

import "crypto/rand"
import "crypto/sha256"
import "database/sql"
import "fmt"
import "io"

const SaltSize = 16

func main() {
    email := []byte("john.doe@somedomain.com")
    password := []byte("47;u5:B(95m72;Xq")

    // create random word
    salt := make([]byte, SaltSize)
    _, err := io.ReadFull(rand.Reader, salt)
    if err != nil {
        panic(err)
    }

    // let's create SHA256(password+salt)
    hash := sha256.New()
    hash.Write(password)
    hash.Write(salt)

    // this is here just for demo purposes
    //
    // fmt.Printf("email   : %s\n", string(email))
    // fmt.Printf("password: %s\n", string(password))
    // fmt.Printf("salt    : %x\n", salt)
    // fmt.Printf("hash    : %x\n", hash.Sum(nil))

    // you're supposed to have a database connection
    stmt, err := db.Prepare("INSERT INTO accounts SET hash=?,salt=?,email=?")
    if err != nil {
        panic(err)
    }
    result, err := stmt.Exec(email, h, salt)
    if err != nil {
        panic(err)
    }

}
```

[^1]: Hashing functions are the subject of Collisions but recommended hashing functions have a really low collisions probability

[1]: /cryptographic-practices/pseudo-random-generators.md
