Validation and storing authentication data
==========================================

## Validation
----------

The key subject of this section is the "authentication data storage", since more
often than desirable, user account databases are leaked on the Internet.
Of course, this is not guaranteed to happen. But in the case of such an event,
collateral damages can be avoided if authentication data, especially passwords,
are stored properly.

First, let's be clear that "_all authentication controls should fail
securely_". We recommend you read all other "Authentication and Password
Management" sections, since they cover recommendations about reporting back
wrong authentication data and how to handle logging.

One other preliminary recommendation is as follow: for sequential authentication
implementations (like Google does nowadays), validation should happen only on
the completion of all data input, on a trusted system (e.g. the server).


## Storing password securely: the theory

Now let's discuss storing passwords.

You really don't need to store passwords, since they are provided by the users
(plaintext). But you will need to validate on each authentication whether users
are providing the same token.

So, for security reasons, what you need is a "one way" function `H`, so that for
every password `p1` and `p2`, `p1` is different from `p2`, `H(p1)` is also
different from `H(p2)`[^1].

Does this sound, or look like math?
Pay attention to this last requirement: `H` should be such a function that
there's no function `H⁻¹` so that `H⁻¹(H(p1))` is equal to `p1`. This means
that there's no way back to the original `p1`, unless you try all possible
values of `p`.

If `H` is one-way only, what's the real problem about account leakage?

Well, if you know all possible passwords, you can pre-compute their hashes and
then run a rainbow table attack.

Certainly you were already told that passwords are hard to manage from user's
point of view, and that users are not only able to re-use passwords, but they
also tend to use something that's easy to remember, hence somehow guessable.

How can we avoid this?

The point is: if two different users provide the same password `p1`, we should
store a different hashed value.
It may sound impossible, but the answer is `salt`: a pseudo-random **unique per
user password** value which is prepended to `p1`, so that the resulting hash is
computed as follows: `H(salt + p1)`.

So each entry on a passwords store should keep the resulting hash, and the
`salt` itself in plaintext: `salt` is not required to remain private.

Last recommendations.
* Avoid using deprecated hashing algorithms (e.g. SHA-1, MD5, etc)
* Read the [Pseudo-Random Generators section][1].

The following code-sample shows a basic example of how this works:

```go
package main

import (
    "crypto/rand"
    "crypto/sha256"
    "database/sql"
    "context"
    "fmt"
)

const saltSize = 32

func main() {
    ctx := context.Background()
    email := []byte("john.doe@somedomain.com")
    password := []byte("47;u5:B(95m72;Xq")

    // create random word
    salt := make([]byte, saltSize)
    _, err := rand.Read(salt)
    if err != nil {
        panic(err)
    }

    // let's create SHA256(salt+password)
    hash := sha256.New()
    hash.Write(salt)
    hash.Write(password)
    h := hash.Sum(nil)

    // this is here just for demo purposes
    //
    // fmt.Printf("email   : %s\n", string(email))
    // fmt.Printf("password: %s\n", string(password))
    // fmt.Printf("salt    : %x\n", salt)
    // fmt.Printf("hash    : %x\n", h)

    // you're supposed to have a database connection
    stmt, err := db.PrepareContext(ctx, "INSERT INTO accounts SET hash=?, salt=?, email=?")
    if err != nil {
        panic(err)
    }
    result, err := stmt.ExecContext(ctx, h, salt, email)
    if err != nil {
        panic(err)
    }

}
```

However, this approach has several flaws and should not be used. It is shown
here only to illustrate the theory with a practical example. The next section
explains how to correctly salt passwords in real life.

## Storing password securely: the practice

One of the most important sayings in cryptography is: **never roll your own
crypto**. By doing so, one can put the entire application at risk. It is a
sensitive and complex topic. Hopefully, cryptography provides tools and
standards reviewed and approved by experts. It is therefore important to use
them instead of trying to re-invent the wheel.

In the case of password storage, the hashing algorithms recommended by
[OWASP][2] are [`bcrypt`][2], [`PDKDF2`][3], [`Argon2`][4] and [`scrypt`][5].
Those take care of hashing and salting passwords in a robust way. Go authors
provide an extended package for cryptography, that is not part of the standard
library. It provides robust implementations for most of the aforementioned
algorithms. It can be downloaded using  `go get`:

```
go get golang.org/x/crypto
```

The following example shows how to use bcrypt, which should be good enough for
most of the situations. The advantage of bcrypt is that it is simpler to use,
and is therefore less error-prone.

```go
package main

import (
    "database/sql"
    "context"
    "fmt"

    "golang.org/x/crypto/bcrypt"
)

func main() {
    ctx := context.Background()
    email := []byte("john.doe@somedomain.com")
    password := []byte("47;u5:B(95m72;Xq")

    // Hash the password with bcrypt
    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }

    // this is here just for demo purposes
    //
    // fmt.Printf("email          : %s\n", string(email))
    // fmt.Printf("password       : %s\n", string(password))
    // fmt.Printf("hashed password: %x\n", hashedPassword)

    // you're supposed to have a database connection
    stmt, err := db.PrepareContext(ctx, "INSERT INTO accounts SET hash=?, email=?")
    if err != nil {
        panic(err)
    }
    result, err := stmt.ExecContext(ctx, hashedPassword, email)
    if err != nil {
        panic(err)
    }
}
```

Bcrypt also provides a simple and secure way to compare a plaintext password
with an already hashed password:

 ```go
 ctx := context.Background()

 // credentials to validate
 email := []byte("john.doe@somedomain.com")
 password := []byte("47;u5:B(95m72;Xq")

// fetch the hashed password corresponding to the provided email
record := db.QueryRowContext(ctx, "SELECT hash FROM accounts WHERE email = ? LIMIT 1", email)

var expectedPassword string
if err := record.Scan(&expectedPassword); err != nil {
    // user does not exist

    // this should be logged (see Error Handling and Logging) but execution
    // should continue
}

if bcrypt.CompareHashAndPassword(password, []byte(expectedPassword)) != nil {
    // passwords do not match

    // passwords mismatch should be logged (see Error Handling and Logging)
    // error should be returned so that a GENERIC message "Sign-in attempt has
    // failed, please check your credentials" can be shown to the user.
}
 ```

If you're not comfortable with password hashing and comparison
options/parameters, better delegating the task to specialized third-party
package with safe defaults. Always opt for an actively maintained package and
remind to check for known issues.

* [passwd][6] - A Go package that provides a safe default abstraction for
  password hashing and comparison. It has support for original go bcrypt
  implementation, argon2, scrypt, parameters masking and key'ed (uncrackable)
  hashes.

[^1]: Hashing functions are the subject of Collisions but recommended hashing functions have a really low collisions probability

[1]: ../cryptographic-practices/pseudo-random-generators.md
[2]: https://www.owasp.org/index.php/Password_Storage_Cheat_Sheet
[3]: https://godoc.org/golang.org/x/crypto/bcrypt
[4]: https://github.com/p-h-c/phc-winner-argon2
[5]: https://godoc.org/golang.org/x/crypto/pbkdf2
[6]: https://github.com/ermites-io/passwd
